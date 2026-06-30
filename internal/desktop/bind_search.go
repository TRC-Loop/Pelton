package desktop

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TRC-Loop/Pelton/internal/search"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// settingSearchWatermark records the highest message id already indexed, so the
// search backfill only walks newly synced rows instead of reindexing everything.
const settingSearchWatermark = "search_indexed_max_id"

// indexFileName is the on-disk Bleve index directory, kept next to the database.
const indexFileName = "search.bleve"

// searchBatchSize bounds how many messages are read and indexed per commit during
// a backfill.
const searchBatchSize = 500

// openSearchIndex opens (creating it if needed) the Bleve index alongside the
// database file.
func openSearchIndex() (*search.Index, error) {
	dbPath, err := storage.DefaultPath()
	if err != nil {
		return nil, err
	}
	return search.Open(filepath.Join(filepath.Dir(dbPath), indexFileName))
}

// backfillSearch brings the index up to date with the cached messages. It runs at
// startup and is cheap once the watermark has caught up.
func (a *App) backfillSearch() {
	a.indexNewMessages()
}

// indexNewMessages indexes every message past the stored watermark in batches,
// advancing the watermark as it goes. It is safe to call repeatedly (after each
// sync); the mutex keeps concurrent passes from racing on the watermark.
func (a *App) indexNewMessages() {
	if a.index == nil {
		return
	}
	a.searchMu.Lock()
	defer a.searchMu.Unlock()

	watermark := a.searchWatermark()
	for {
		msgs, err := a.store.ListMessagesForIndex(a.ctx, watermark, searchBatchSize)
		if err != nil {
			a.log.Error("read messages for search index", "err", err)
			return
		}
		if len(msgs) == 0 {
			return
		}

		docs := make([]search.Doc, 0, len(msgs))
		for _, m := range msgs {
			docs = append(docs, toSearchDoc(m))
			if m.ID > watermark {
				watermark = m.ID
			}
		}
		if err := a.index.IndexBatch(docs); err != nil {
			a.log.Error("index message batch", "err", err)
			return
		}
		if err := a.store.Set(a.ctx, settingSearchWatermark, strconv.FormatInt(watermark, 10)); err != nil {
			a.log.Error("persist search watermark", "err", err)
		}
		if len(msgs) < searchBatchSize {
			return
		}
	}
}

// searchWatermark reads the highest indexed message id, defaulting to 0 (index
// everything) when unset or unparsable.
func (a *App) searchWatermark() int64 {
	raw := a.stringSetting(settingSearchWatermark, "0")
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// toSearchDoc projects a stored message into the search document. The sender name
// and address are combined so a search for either finds the mail.
func toSearchDoc(m storage.Message) search.Doc {
	return search.Doc{
		ID:        m.ID,
		AccountID: m.AccountID,
		FolderID:  m.FolderID,
		Subject:   m.Subject,
		From:      strings.TrimSpace(m.FromName + " " + m.FromAddress),
		To:        m.ToAddresses,
		Cc:        m.CcAddresses,
		Body:      m.BodyPlain,
		Date:      m.Date,
	}
}

// SearchRequestDTO is a search query from the ui: free text plus an optional date
// window. AfterUnix/BeforeUnix are unix seconds; 0 means that side is open.
type SearchRequestDTO struct {
	Query      string `json:"query"`
	AfterUnix  int64  `json:"afterUnix"`
	BeforeUnix int64  `json:"beforeUnix"`
	Limit      int    `json:"limit"`
}

// Search runs a ranked, typo-tolerant search over cached messages and returns the
// matching summaries in relevance order. An empty query with a date window lists
// messages in that window; an empty request returns nothing.
func (a *App) Search(req SearchRequestDTO) ([]MessageSummaryDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	if a.index == nil {
		return nil, errSearchUnavailable
	}
	if strings.TrimSpace(req.Query) == "" && req.AfterUnix == 0 && req.BeforeUnix == 0 {
		return []MessageSummaryDTO{}, nil
	}

	q := search.Query{
		Text:  strings.TrimSpace(req.Query),
		Limit: req.Limit,
	}
	if req.AfterUnix > 0 {
		q.After = time.Unix(req.AfterUnix, 0)
	}
	if req.BeforeUnix > 0 {
		q.Before = time.Unix(req.BeforeUnix, 0)
	}

	hits, err := a.index.Search(q)
	if err != nil {
		return nil, err
	}

	// hits are ranked; fetch each full message so rows render like the normal
	// list. a missing message (deleted since indexing) is simply skipped, which
	// also covers stale index entries without a separate cleanup pass.
	out := make([]MessageSummaryDTO, 0, len(hits))
	for _, h := range hits {
		m, err := a.store.GetMessage(a.ctx, h.ID)
		if err != nil {
			continue
		}
		email, folderName := a.lookupContext(a.ctx, m.AccountID, m.FolderID)
		out = append(out, toSummaryDTO(*m, email, folderName))
	}
	return out, nil
}

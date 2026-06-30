// Package search is Pelton's local full-text search over cached messages, built
// on Bleve. It replaces the sqlite fts5 path with a ranked, typo-tolerant index:
// queries match across subject, sender, recipients and body with fuzzy matching
// so small spelling slips still find mail, and results come back scored by
// relevance. The index lives on disk next to the database and is kept current by
// the desktop layer (incremental indexing on sync, with a startup backfill).
package search

import (
	"fmt"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// field boosts weight a hit by where the term matched: a subject or sender match
// ranks above a body match.
const (
	boostSubject = 3.0
	boostFrom    = 2.0
	boostRecip   = 1.5
	boostBody    = 1.0

	// fuzziness is the edit distance tolerated per term (1 catches most typos
	// without drowning results in noise).
	fuzziness = 1

	defaultLimit = 50
)

// Doc is the indexable projection of a message. Only the fields worth searching
// are kept; the id ties a hit back to the stored message.
type Doc struct {
	ID        int64
	AccountID int64
	FolderID  int64
	Subject   string
	From      string
	To        string
	Cc        string
	Body      string
	Date      time.Time
}

// Query is a search request: free text plus an optional date window.
type Query struct {
	Text string
	// After/Before bound the message date. A zero time means the side is open.
	After  time.Time
	Before time.Time
	Limit  int
}

// Hit is one search result: the message id and its relevance score.
type Hit struct {
	ID    int64
	Score float64
}

// Index is the Bleve-backed search index. It is safe for concurrent use; Bleve
// serializes writes internally.
type Index struct {
	idx bleve.Index
}

// Open opens the index at path, creating it with the message mapping if it does
// not exist yet.
func Open(path string) (*Index, error) {
	idx, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		idx, err = bleve.New(path, buildMapping())
	}
	if err != nil {
		return nil, fmt.Errorf("search: open index %q: %w", path, err)
	}
	return &Index{idx: idx}, nil
}

// Close releases the index.
func (i *Index) Close() error {
	return i.idx.Close()
}

// DocCount reports how many documents are indexed, used to decide whether a
// backfill is needed.
func (i *Index) DocCount() (uint64, error) {
	return i.idx.DocCount()
}

// IndexDoc adds or replaces one message in the index.
func (i *Index) IndexDoc(d Doc) error {
	if err := i.idx.Index(docID(d.ID), toIndexable(d)); err != nil {
		return fmt.Errorf("search: index message %d: %w", d.ID, err)
	}
	return nil
}

// IndexBatch indexes many messages in one commit, far faster than one at a time
// for the backfill.
func (i *Index) IndexBatch(docs []Doc) error {
	batch := i.idx.NewBatch()
	for _, d := range docs {
		if err := batch.Index(docID(d.ID), toIndexable(d)); err != nil {
			return fmt.Errorf("search: batch message %d: %w", d.ID, err)
		}
	}
	if err := i.idx.Batch(batch); err != nil {
		return fmt.Errorf("search: commit batch: %w", err)
	}
	return nil
}

// Delete removes a message from the index.
func (i *Index) Delete(id int64) error {
	if err := i.idx.Delete(docID(id)); err != nil {
		return fmt.Errorf("search: delete message %d: %w", id, err)
	}
	return nil
}

// Search runs a query and returns matching message ids ranked by relevance.
func (i *Index) Search(q Query) ([]Hit, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = defaultLimit
	}

	req := bleve.NewSearchRequestOptions(i.build(q), limit, 0, false)
	res, err := i.idx.Search(req)
	if err != nil {
		return nil, fmt.Errorf("search: query %q: %w", q.Text, err)
	}

	hits := make([]Hit, 0, len(res.Hits))
	for _, h := range res.Hits {
		id, convErr := strconv.ParseInt(h.ID, 10, 64)
		if convErr != nil {
			continue
		}
		hits = append(hits, Hit{ID: id, Score: h.Score})
	}
	return hits, nil
}

// build assembles the Bleve query from the request: a text part (fuzzy, multi
// field, ranked) conjoined with a date-range part when bounds are set.
func (i *Index) build(q Query) query.Query {
	var parts []query.Query

	if q.Text != "" {
		parts = append(parts, textQuery(q.Text))
	}
	if !q.After.IsZero() || !q.Before.IsZero() {
		parts = append(parts, dateQuery(q.After, q.Before))
	}

	switch len(parts) {
	case 0:
		return bleve.NewMatchAllQuery()
	case 1:
		return parts[0]
	default:
		return bleve.NewConjunctionQuery(parts...)
	}
}

// textQuery matches the free text across all fields with per-field boosts and
// fuzzy matching, so typos still hit and the best field wins the score. Each hit
// must match the text somewhere (the per-field alternatives form a disjunction).
func textQuery(text string) query.Query {
	fields := []struct {
		name  string
		boost float64
	}{
		{"subject", boostSubject},
		{"from", boostFrom},
		{"to", boostRecip},
		{"cc", boostRecip},
		{"body", boostBody},
	}

	alts := make([]query.Query, 0, len(fields))
	for _, f := range fields {
		mq := bleve.NewMatchQuery(text)
		mq.SetField(f.name)
		mq.SetFuzziness(fuzziness)
		mq.SetBoost(f.boost)
		alts = append(alts, mq)
	}
	return bleve.NewDisjunctionQuery(alts...)
}

// dateQuery bounds the message date. Bleve treats a zero time as an open side
// when passed through; we pass explicit min/max to keep the meaning clear.
func dateQuery(after, before time.Time) query.Query {
	lo := after
	if lo.IsZero() {
		lo = time.Unix(0, 0)
	}
	hi := before
	if hi.IsZero() {
		hi = time.Now().AddDate(100, 0, 0)
	}
	dq := bleve.NewDateRangeQuery(lo, hi)
	dq.SetField("date")
	return dq
}

// toIndexable projects a Doc into the field map Bleve indexes (keys must match
// the mapping field names).
func toIndexable(d Doc) map[string]any {
	return map[string]any{
		"subject": d.Subject,
		"from":    d.From,
		"to":      d.To,
		"cc":      d.Cc,
		"body":    d.Body,
		"date":    d.Date,
	}
}

// docID renders a message id as the stable Bleve document id.
func docID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// buildMapping describes the message document: standard-analyzed text fields and
// a datetime field for range filtering.
func buildMapping() *mapping.IndexMappingImpl {
	text := bleve.NewTextFieldMapping()
	text.Store = false

	date := bleve.NewDateTimeFieldMapping()
	date.Store = false

	doc := bleve.NewDocumentMapping()
	for _, name := range []string{"subject", "from", "to", "cc", "body"} {
		doc.AddFieldMappingsAt(name, text)
	}
	doc.AddFieldMappingsAt("date", date)

	im := bleve.NewIndexMapping()
	im.DefaultMapping = doc
	return im
}

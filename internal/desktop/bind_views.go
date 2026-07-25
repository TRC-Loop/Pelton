package desktop

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TRC-Loop/Pelton/internal/search"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// maxViewMatches caps how many messages a single view resolves to. It bounds the
// per-view eager-run cost (counts run on startup and after every sync) and the
// warm result set. A view that would match more is truncated newest-first.
const maxViewMatches = 500

// ViewDTO is the settings/sidebar view of a saved search: its definition plus the
// eager-run counts. AccountID is 0 for "all accounts".
type ViewDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`

	QueryText    string `json:"queryText"`
	QueryFrom    string `json:"queryFrom"`
	QueryTo      string `json:"queryTo"`
	QuerySubject string `json:"querySubject"`
	WithinDays   int    `json:"withinDays"`

	UnreadOnly    bool  `json:"unreadOnly"`
	FlaggedOnly   bool  `json:"flaggedOnly"`
	HasAttachment bool  `json:"hasAttachment"`
	AccountID     int64 `json:"accountId"`

	Position    int `json:"position"`
	UnreadCount int `json:"unreadCount"`
	TotalCount  int `json:"totalCount"`
}

// ListViews returns every saved view with its current eager-run counts, ordered
// for the sidebar.
func (a *App) ListViews() ([]ViewDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	views, err := a.store.ListViews(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ViewDTO, 0, len(views))
	for _, v := range views {
		total, unread := a.viewCounts(a.ctx, v)
		dto := viewToDTO(v)
		dto.TotalCount = total
		dto.UnreadCount = unread
		out = append(out, dto)
	}
	return out, nil
}

// SaveView creates a view (id 0) or updates an existing one, then returns the
// saved view with fresh counts. A blank name is rejected.
func (a *App) SaveView(dto ViewDTO) (ViewDTO, error) {
	if err := a.ready(); err != nil {
		return ViewDTO{}, err
	}
	if strings.TrimSpace(dto.Name) == "" {
		return ViewDTO{}, fmt.Errorf("view name is required")
	}
	v := dtoToView(dto)
	if v.ID == 0 {
		if _, err := a.store.CreateView(a.ctx, &v); err != nil {
			return ViewDTO{}, err
		}
	} else if err := a.store.UpdateView(a.ctx, &v); err != nil {
		return ViewDTO{}, err
	}

	a.emit(EventViewsChanged, nil)
	total, unread := a.viewCounts(a.ctx, v)
	out := viewToDTO(v)
	out.TotalCount = total
	out.UnreadCount = unread
	return out, nil
}

// DeleteView removes a saved view.
func (a *App) DeleteView(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	if err := a.store.DeleteView(a.ctx, id); err != nil {
		return err
	}
	a.emit(EventViewsChanged, nil)
	return nil
}

// ReorderViews persists a new sidebar order for the given view ids.
func (a *App) ReorderViews(orderedIDs []int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	if err := a.store.SetViewPositions(a.ctx, orderedIDs); err != nil {
		return err
	}
	a.emit(EventViewsChanged, nil)
	return nil
}

// refreshViewCounts recomputes every view's counts and notifies the ui. It is
// the eager-run entry point, called on startup and after each sync. It runs off
// the caller's hot path (sync) via a goroutine at the call sites.
func (a *App) refreshViewCounts() {
	if err := a.ready(); err != nil {
		return
	}
	views, err := a.store.ListViews(a.ctx)
	if err != nil {
		a.log.Error("refresh view counts", "err", err)
		return
	}
	if len(views) == 0 {
		return
	}
	a.emit(EventViewsChanged, nil)
}

// viewCounts returns the total and unread match counts for a view, swallowing
// errors as zero so one broken view never blanks the whole sidebar.
func (a *App) viewCounts(ctx context.Context, v storage.View) (total, unread int) {
	matches, err := a.runViewMatches(ctx, v)
	if err != nil {
		a.log.Error("run view", "view", v.Name, "err", err)
		return 0, 0
	}
	for _, m := range matches {
		if !m.Flags.Has(storage.FlagSeen) {
			unread++
		}
	}
	return len(matches), unread
}

// runViewMatches resolves a view to its matching messages, newest first. Text
// criteria go through the full-text index; a pure scope view reads the store
// directly. Scope, date and hygiene filters are then applied in Go, since flags
// are not indexed.
func (a *App) runViewMatches(ctx context.Context, v storage.View) ([]storage.Message, error) {
	var after time.Time
	if v.WithinDays > 0 {
		after = time.Now().Add(-time.Duration(v.WithinDays) * 24 * time.Hour)
	}

	hasText := v.QueryText != "" || v.QueryFrom != "" || v.QueryTo != "" || v.QuerySubject != ""

	var candidates []storage.Message
	if hasText {
		if a.index == nil {
			return nil, errSearchUnavailable
		}
		hits, err := a.index.Search(search.Query{
			Text:    v.QueryText,
			From:    v.QueryFrom,
			To:      v.QueryTo,
			Subject: v.QuerySubject,
			After:   after,
			Limit:   maxViewMatches,
		})
		if err != nil {
			return nil, err
		}
		for _, h := range hits {
			m, err := a.store.GetMessage(ctx, h.ID)
			if err != nil {
				continue
			}
			candidates = append(candidates, *m)
		}
	} else {
		folderIDs, err := a.viewScopeFolderIDs(ctx, v.AccountID)
		if err != nil {
			return nil, err
		}
		msgs, err := a.store.QueryMessages(ctx, storage.MessageQuery{FolderIDs: folderIDs, Limit: maxViewMatches})
		if err != nil {
			return nil, err
		}
		candidates = msgs
	}

	out := make([]storage.Message, 0, len(candidates))
	for _, m := range candidates {
		if m.SnoozeHidden {
			continue
		}
		if v.AccountID != 0 && m.AccountID != v.AccountID {
			continue
		}
		if v.UnreadOnly && m.Flags.Has(storage.FlagSeen) {
			continue
		}
		if v.FlaggedOnly && !m.Flags.Has(storage.FlagFlagged) {
			continue
		}
		if v.HasAttachment && !m.HasAttachments {
			continue
		}
		if !after.IsZero() && m.Date.Before(after) {
			continue
		}
		out = append(out, m)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	if len(out) > maxViewMatches {
		out = out[:maxViewMatches]
	}
	return out, nil
}

// viewScopeFolderIDs returns the folder ids a scope-only view spans: one
// account's folders, or every account's when accountID is 0.
func (a *App) viewScopeFolderIDs(ctx context.Context, accountID int64) ([]int64, error) {
	var accountIDs []int64
	if accountID != 0 {
		accountIDs = []int64{accountID}
	} else {
		accounts, err := a.store.ListAccounts(ctx)
		if err != nil {
			return nil, err
		}
		for _, acc := range accounts {
			accountIDs = append(accountIDs, acc.ID)
		}
	}

	var folderIDs []int64
	for _, id := range accountIDs {
		folders, err := a.store.ListFolders(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, f := range folders {
			folderIDs = append(folderIDs, f.ID)
		}
	}
	return folderIDs, nil
}

// listSavedView returns a page of a saved view's resolved messages. It backs the
// "savedView" list kind.
func (a *App) listSavedView(ctx context.Context, viewID int64, limit, offset int) (MessageListDTO, error) {
	v, err := a.store.GetView(ctx, viewID)
	if err != nil {
		return MessageListDTO{}, err
	}
	matches, err := a.runViewMatches(ctx, *v)
	if err != nil {
		return MessageListDTO{}, err
	}

	total := len(matches)
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	page := matches[offset:end]

	summaries, err := a.buildSummaries(ctx, page)
	if err != nil {
		return MessageListDTO{}, err
	}
	return MessageListDTO{Messages: summaries, Total: total}, nil
}

func viewToDTO(v storage.View) ViewDTO {
	return ViewDTO{
		ID:            v.ID,
		Name:          v.Name,
		Icon:          v.Icon,
		Color:         v.Color,
		QueryText:     v.QueryText,
		QueryFrom:     v.QueryFrom,
		QueryTo:       v.QueryTo,
		QuerySubject:  v.QuerySubject,
		WithinDays:    v.WithinDays,
		UnreadOnly:    v.UnreadOnly,
		FlaggedOnly:   v.FlaggedOnly,
		HasAttachment: v.HasAttachment,
		AccountID:     v.AccountID,
		Position:      v.Position,
	}
}

func dtoToView(d ViewDTO) storage.View {
	return storage.View{
		ID:            d.ID,
		Name:          strings.TrimSpace(d.Name),
		Icon:          d.Icon,
		Color:         d.Color,
		QueryText:     strings.TrimSpace(d.QueryText),
		QueryFrom:     strings.TrimSpace(d.QueryFrom),
		QueryTo:       strings.TrimSpace(d.QueryTo),
		QuerySubject:  strings.TrimSpace(d.QuerySubject),
		WithinDays:    d.WithinDays,
		UnreadOnly:    d.UnreadOnly,
		FlaggedOnly:   d.FlaggedOnly,
		HasAttachment: d.HasAttachment,
		AccountID:     d.AccountID,
		Position:      d.Position,
	}
}

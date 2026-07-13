package desktop

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/mailview"
	"github.com/TRC-Loop/Pelton/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ListMessagesRequest selects the page to read. Kind is "folder" or "view".
// FolderID applies to "folder"; View (an inbox/flagged/sent/drafts key) applies
// to "view". Limit and Offset drive pagination.
type ListMessagesRequest struct {
	Kind     string `json:"kind"`
	FolderID int64  `json:"folderId"`
	View     string `json:"view"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// ListMessages returns a page of summaries for a folder or a unified view, plus
// the total count for pagination. This is the single read path the list uses.
func (a *App) ListMessages(req ListMessagesRequest) (MessageListDTO, error) {
	if err := a.ready(); err != nil {
		return MessageListDTO{}, err
	}

	q, err := a.requestQuery(a.ctx, req)
	if err != nil {
		return MessageListDTO{}, err
	}
	q.Limit = req.Limit
	q.Offset = req.Offset

	messages, err := a.store.QueryMessages(a.ctx, q)
	if err != nil {
		return MessageListDTO{}, err
	}
	total, err := a.store.CountMessages(a.ctx, q)
	if err != nil {
		return MessageListDTO{}, err
	}

	summaries, err := a.buildSummaries(a.ctx, messages)
	if err != nil {
		return MessageListDTO{}, err
	}
	return MessageListDTO{Messages: summaries, Total: total}, nil
}

// requestQuery turns a list request into the storage query (the folder id set
// and any flag filter), delegating unified views to viewQuery.
func (a *App) requestQuery(ctx context.Context, req ListMessagesRequest) (storage.MessageQuery, error) {
	if req.Kind == "view" {
		return a.viewQuery(ctx, req.View)
	}
	return storage.MessageQuery{FolderIDs: []int64{req.FolderID}}, nil
}

// GetMessage returns the full message for the reading pane: sanitized html with
// remote content blocked and inline cid images resolved, the plain alternative,
// and the downloadable attachment list.
func (a *App) GetMessage(id int64) (MessageDetailDTO, error) {
	if err := a.ready(); err != nil {
		return MessageDetailDTO{}, err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return MessageDetailDTO{}, err
	}

	email, folderName := a.lookupContext(a.ctx, m.AccountID, m.FolderID)
	summary := toSummaryDTO(*m, email, folderName)

	atts, err := a.store.ListAttachments(a.ctx, id)
	if err != nil {
		return MessageDetailDTO{}, err
	}

	// trusted senders/domains (or the global setting) render remote content with
	// no prompt; everyone else is blocked until the user asks.
	autoAllow := a.remoteAutoAllow(m.FromAddress)

	detail := MessageDetailDTO{
		MessageSummaryDTO: summary,
		ToAddresses:       m.ToAddresses,
		CcAddresses:       m.CcAddresses,
		BodyPlain:         m.BodyPlain,
		IsHTML:            m.BodyHTML != "",
		HasRemoteContent:  mailview.HasRemoteContent(m.BodyHTML),
		RemoteAllowed:     autoAllow,
		RemoteHosts:       mailview.RemoteHosts(m.BodyHTML),
		Attachments:       toAttachmentDTOs(atts, m.BodyHTML),
		Unsubscribe:       a.unsubscribeInfo(m),
	}
	detail.BodyHTMLSafe = a.renderHTML(m.BodyHTML, atts, autoAllow)
	return detail, nil
}

// GetMessageHTML re-renders a message body with the chosen remote policy. The ui
// calls it with allowRemote=true when the user clicks "load remote images".
func (a *App) GetMessageHTML(id int64, allowRemote bool) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return "", err
	}
	atts, err := a.store.ListAttachments(a.ctx, id)
	if err != nil {
		return "", err
	}
	return a.renderHTML(m.BodyHTML, atts, allowRemote), nil
}

// renderHTML resolves inline cid images to data urls then sanitizes with the
// given remote policy. Inlining happens before sanitize so the cid scheme is
// already gone and only trusted data urls remain.
func (a *App) renderHTML(html string, atts []storage.Attachment, allowRemote bool) string {
	if html == "" {
		return ""
	}
	resolved := mailview.ResolveCIDs(html, a.inlineDataURLs(atts))
	return mailview.Sanitize(resolved, allowRemote)
}

// inlineDataURLs builds a content-id to data-url map for inline attachments by
// reading their files off disk. Failures are skipped so one unreadable inline
// image never breaks the whole body.
func (a *App) inlineDataURLs(atts []storage.Attachment) map[string]string {
	out := make(map[string]string)
	for _, att := range atts {
		if att.ContentID == "" {
			continue
		}
		rc, err := a.store.OpenAttachment(att.DiskPath)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		id := trimAngles(att.ContentID)
		out[id] = fmt.Sprintf("data:%s;base64,%s", att.ContentType, base64.StdEncoding.EncodeToString(data))
	}
	return out
}

// SetSeen sets or clears the \Seen flag on a message and queues the change for
// the next sync to push to the server.
func (a *App) SetSeen(id int64, seen bool) error {
	return a.updateFlag(id, storage.FlagSeen, seen)
}

// SetFlagged sets or clears the \Flagged flag and queues the push.
func (a *App) SetFlagged(id int64, flagged bool) error {
	return a.updateFlag(id, storage.FlagFlagged, flagged)
}

// updateFlag toggles one flag bit on a message and marks it pending so sync
// pushes it. It reads the current mask first so other flags are preserved.
func (a *App) updateFlag(id int64, flag storage.Flag, on bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return err
	}
	flags := m.Flags
	if on {
		flags |= flag
	} else {
		flags &^= flag
	}
	return a.store.MarkFlagsPending(a.ctx, id, flags)
}

// DeleteMessage marks a message for deletion. The row is kept and hidden from
// the list until the next sync expunges it on the server, then it is purged
// locally. This is the safe path: nothing is lost if the server rejects it.
func (a *App) DeleteMessage(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.MarkDeletePending(a.ctx, id)
}

// UndoDelete reverses a pending delete while the message is still cached (before
// the next sync expunges it), bringing the row back into the list.
func (a *App) UndoDelete(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.ClearDeletePending(a.ctx, id)
}

// SaveAttachment prompts for a destination and writes the attachment file there,
// returning the chosen path (empty if the user cancelled). messageID scopes the
// lookup so the id cannot reach another message's files.
func (a *App) SaveAttachment(messageID, attachmentID int64) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	atts, err := a.store.ListAttachments(a.ctx, messageID)
	if err != nil {
		return "", err
	}
	var target *storage.Attachment
	for i := range atts {
		if atts[i].ID == attachmentID {
			target = &atts[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("pelton: attachment %d not found", attachmentID)
	}

	dest, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: target.Filename,
		Title:           "Save attachment",
	})
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil
	}
	if err := a.copyAttachment(target.DiskPath, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// copyAttachment streams an attachment file from disk to dest.
func (a *App) copyAttachment(diskPath, dest string) error {
	src, err := a.store.OpenAttachment(diskPath)
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(filepath.Clean(dest))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return err
	}
	return nil
}

// buildSummaries flattens stored messages into list rows, resolving each
// message's account email and folder name through small caches so a page of
// rows from many folders does not re-query the same context repeatedly.
func (a *App) buildSummaries(ctx context.Context, messages []storage.Message) ([]MessageSummaryDTO, error) {
	emailCache := make(map[int64]string)
	folderCache := make(map[int64]string)

	out := make([]MessageSummaryDTO, 0, len(messages))
	for _, m := range messages {
		email, ok := emailCache[m.AccountID]
		if !ok {
			if acc, err := a.store.GetAccount(ctx, m.AccountID); err == nil {
				email = acc.Email
			}
			emailCache[m.AccountID] = email
		}
		folderName, ok := folderCache[m.FolderID]
		if !ok {
			if f, err := a.store.GetFolder(ctx, m.FolderID); err == nil {
				folderName = f.Name
			}
			folderCache[m.FolderID] = folderName
		}
		out = append(out, toSummaryDTO(m, email, folderName))
	}
	return out, nil
}

// lookupContext resolves one message's account email and folder name, tolerating
// missing rows by returning empty strings.
func (a *App) lookupContext(ctx context.Context, accountID, folderID int64) (email, folderName string) {
	if acc, err := a.store.GetAccount(ctx, accountID); err == nil {
		email = acc.Email
	}
	if f, err := a.store.GetFolder(ctx, folderID); err == nil {
		folderName = f.Name
	}
	return email, folderName
}

// toAttachmentDTOs flattens stored attachments. A part is treated as inline (and
// hidden from the download list) only when it carries a content id that the body
// actually references via a cid: url. This keeps real attachments that merely
// carry a content id visible in the list.
func toAttachmentDTOs(atts []storage.Attachment, bodyHTML string) []AttachmentDTO {
	referenced := mailview.ReferencedCIDs(bodyHTML)
	out := make([]AttachmentDTO, 0, len(atts))
	for _, att := range atts {
		inline := att.ContentID != "" && referenced[trimAngles(strings.ToLower(att.ContentID))]
		out = append(out, AttachmentDTO{
			ID:          att.ID,
			Filename:    att.Filename,
			ContentType: att.ContentType,
			SizeBytes:   att.SizeBytes,
			Inline:      inline,
		})
	}
	return out
}

// trimAngles strips the surrounding <> some content ids carry.
func trimAngles(s string) string {
	s = trimPrefixByte(s, '<')
	s = trimSuffixByte(s, '>')
	return s
}

func trimPrefixByte(s string, b byte) string {
	if len(s) > 0 && s[0] == b {
		return s[1:]
	}
	return s
}

func trimSuffixByte(s string, b byte) string {
	if len(s) > 0 && s[len(s)-1] == b {
		return s[:len(s)-1]
	}
	return s
}

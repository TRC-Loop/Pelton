package desktop

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/TRC-Loop/Pelton/internal/crypto"
	"github.com/TRC-Loop/Pelton/internal/smtp"
)

// AddressDTO is one mail address with an optional display name.
type AddressDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ComposeAttachment is a file the user attached. Content is base64 so it crosses
// the bindings boundary as a plain string. Inline marks an image referenced from
// the html body by ContentID.
type ComposeAttachment struct {
	Filename      string `json:"filename"`
	ContentType   string `json:"contentType"`
	ContentBase64 string `json:"contentBase64"`
	Inline        bool   `json:"inline"`
	ContentID     string `json:"contentId"`
}

// ComposeRequest is the full input to send or save a message. The frontend
// produces both Text and HTML: for markdown mode it renders markdown to html
// itself and sends the markdown source as Text; for plaintext mode HTML is
// empty. Threading fields are copied from the message being replied to.
type ComposeRequest struct {
	AccountID   int64               `json:"accountId"`
	To          []AddressDTO        `json:"to"`
	Cc          []AddressDTO        `json:"cc"`
	Bcc         []AddressDTO        `json:"bcc"`
	Subject     string              `json:"subject"`
	Text        string              `json:"text"`
	HTML        string              `json:"html"`
	InReplyTo   string              `json:"inReplyTo"`
	References  []string            `json:"references"`
	Attachments []ComposeAttachment `json:"attachments"`
}

// SendMessage builds the mime message and enqueues it in the durable outbox. The
// background worker transmits it when smtp credentials are configured. This call
// returns once the message is safely queued, so the ui can confirm immediately.
// Crypto mode is none here; the per-message pgp toggle is a later ui addition
// that would pass a mode and options through to smtp.Enqueue.
func (a *App) SendMessage(req ComposeRequest) (int64, error) {
	if err := a.ready(); err != nil {
		return 0, err
	}

	msg, err := a.buildMessage(req)
	if err != nil {
		return 0, err
	}

	// apply the configured send delay so the worker holds the message and the user
	// can undo within the window. zero (or negative) sends as soon as possible.
	var notBefore time.Time
	if delay := a.intSetting(settingSendDelay, 0); delay > 0 {
		notBefore = time.Now().UTC().Add(time.Duration(delay) * time.Second)
	}

	id, err := smtp.Enqueue(a.ctx, a.queue, req.AccountID, msg, nil, crypto.ModeNone, crypto.Options{}, notBefore)
	if err != nil {
		return 0, err
	}
	a.emit(EventOutboxChanged, nil)

	// harvest every recipient into the address book so autocomplete learns from
	// who the user writes to. best effort: a failure here must not fail the send.
	for _, group := range [][]AddressDTO{req.To, req.Cc, req.Bcc} {
		for _, addr := range group {
			if err := a.store.RecordAddress(a.ctx, addr.Email, addr.Name); err != nil {
				a.log.Error("record recipient address", "email", addr.Email, "err", err)
			}
		}
	}
	return id, nil
}

// CancelSend pulls a still-queued message back out of the outbox, returning
// whether it was cancelled. It only succeeds while the message is waiting in its
// undo-send delay window; once the worker has claimed it for sending it cannot be
// recalled. The frontend reopens the original draft after a successful cancel.
func (a *App) CancelSend(id int64) (bool, error) {
	if err := a.ready(); err != nil {
		return false, err
	}
	cancelled, err := a.queue.Cancel(a.ctx, id)
	if err != nil {
		return false, err
	}
	if cancelled {
		a.emit(EventOutboxChanged, nil)
	}
	return cancelled, nil
}

// ClearSentOutbox removes rows already marked sent. The ui calls it after showing
// the brief "sent" confirmation so the queue does not keep completed messages.
func (a *App) ClearSentOutbox() error {
	if err := a.ready(); err != nil {
		return err
	}
	n, err := a.queue.PruneSent(a.ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		a.emit(EventOutboxChanged, nil)
	}
	return nil
}

// buildMessage assembles an smtp.Message from a compose request, resolving the
// from address from the account and decoding attachment bytes.
func (a *App) buildMessage(req ComposeRequest) (*smtp.Message, error) {
	acc, err := a.store.GetAccount(a.ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	atts, err := decodeAttachments(req.Attachments)
	if err != nil {
		return nil, err
	}

	return &smtp.Message{
		From:        smtp.Address{Name: acc.DisplayName, Email: acc.Email},
		To:          toBuilderAddresses(req.To),
		Cc:          toBuilderAddresses(req.Cc),
		Bcc:         toBuilderAddresses(req.Bcc),
		Subject:     req.Subject,
		Text:        req.Text,
		HTML:        req.HTML,
		Attachments: atts,
		InReplyTo:   req.InReplyTo,
		References:  req.References,
	}, nil
}

// DraftDTO is a locally saved unsent draft. Drafts are stored in the settings
// table as json for now. Appending drafts to the server Drafts folder needs a
// live imap connection and credentials, which arrive with the account-setup and
// keyring step; until then drafts are local only.
// TODO(backend): once credentials exist, also AppendToDrafts on the imap client.
type DraftDTO struct {
	ID      int64          `json:"id"`
	SavedAt string         `json:"savedAt"`
	Request ComposeRequest `json:"request"`
}

// draftsKey is the settings key holding the json array of local drafts.
const draftsKey = "local_drafts"

// SaveDraft stores a compose request as a local draft and returns its id. An id
// of 0 in the request creates a new draft; a non zero id replaces that draft.
func (a *App) SaveDraft(id int64, req ComposeRequest) (int64, error) {
	if err := a.ready(); err != nil {
		return 0, err
	}
	drafts, err := a.loadDrafts()
	if err != nil {
		return 0, err
	}

	if id == 0 {
		id = time.Now().UnixNano()
		drafts = append(drafts, DraftDTO{ID: id})
	}
	for i := range drafts {
		if drafts[i].ID == id {
			drafts[i].Request = req
			drafts[i].SavedAt = time.Now().UTC().Format(time.RFC3339)
		}
	}
	if err := a.store.SetJSON(a.ctx, draftsKey, drafts); err != nil {
		return 0, err
	}
	return id, nil
}

// ListDrafts returns the locally saved drafts, newest first by save time.
func (a *App) ListDrafts() ([]DraftDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.loadDrafts()
}

// DeleteDraft removes a local draft by id.
func (a *App) DeleteDraft(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	drafts, err := a.loadDrafts()
	if err != nil {
		return err
	}
	kept := drafts[:0]
	for _, d := range drafts {
		if d.ID != id {
			kept = append(kept, d)
		}
	}
	return a.store.SetJSON(a.ctx, draftsKey, kept)
}

// loadDrafts reads the drafts json, treating an unset key as an empty list.
func (a *App) loadDrafts() ([]DraftDTO, error) {
	var drafts []DraftDTO
	err := a.store.GetJSON(a.ctx, draftsKey, &drafts)
	if err != nil {
		if isSettingMissing(err) {
			return nil, nil
		}
		return nil, err
	}
	return drafts, nil
}

// OutboxRowDTO is one queued or failed message for the outbox view.
type OutboxRowDTO struct {
	ID            int64    `json:"id"`
	AccountID     int64    `json:"accountId"`
	Recipients    []string `json:"recipients"`
	State         string   `json:"state"`
	Attempts      int      `json:"attempts"`
	LastError     string   `json:"lastError"`
	NextAttemptAt string   `json:"nextAttemptAt"`
	CreatedAt     string   `json:"createdAt"`
}

// ListOutbox returns the current outbox, so the ui can show queued, sending and
// failed messages and surface send failures explicitly.
func (a *App) ListOutbox() ([]OutboxRowDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	rows, err := a.queue.List(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]OutboxRowDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, OutboxRowDTO{
			ID:            r.ID,
			AccountID:     r.AccountID,
			Recipients:    r.Recipients,
			State:         r.State,
			Attempts:      r.Attempts,
			LastError:     r.LastError,
			NextAttemptAt: formatDate(r.NextAttemptAt),
			CreatedAt:     formatDate(r.CreatedAt),
		})
	}
	return out, nil
}

// decodeAttachments base64-decodes compose attachments into builder attachments.
func decodeAttachments(in []ComposeAttachment) ([]smtp.Attachment, error) {
	out := make([]smtp.Attachment, 0, len(in))
	for _, att := range in {
		data, err := base64.StdEncoding.DecodeString(att.ContentBase64)
		if err != nil {
			return nil, fmt.Errorf("pelton: decode attachment %q: %w", att.Filename, err)
		}
		out = append(out, smtp.Attachment{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			Content:     data,
			Inline:      att.Inline,
			ContentID:   att.ContentID,
		})
	}
	return out, nil
}

// toBuilderAddresses converts address dtos to builder addresses.
func toBuilderAddresses(in []AddressDTO) []smtp.Address {
	out := make([]smtp.Address, 0, len(in))
	for _, a := range in {
		out = append(out, smtp.Address{Name: a.Name, Email: a.Email})
	}
	return out
}

package desktop

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/TRC-Loop/Pelton/internal/smtp"
)

// ErrSendAtInvalid is returned when a ComposeRequest.SendAt string does not
// parse as RFC3339.
var ErrSendAtInvalid = fmt.Errorf("pelton: sendAt is not a valid RFC3339 timestamp")

// ErrSendAtPast is returned when a ComposeRequest.SendAt is not in the future.
var ErrSendAtPast = fmt.Errorf("pelton: sendAt must be in the future")

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
	// SendAt is an optional RFC3339 timestamp for a user-scheduled send ("send
	// later"). Empty means send immediately, subject to the undo-send delay
	// below. When set it must be in the future.
	SendAt string `json:"sendAt"`
	// Protection is the pgp treatment for this one message: "none", "sign",
	// "encrypt" or "signencrypt". Empty means none. The compose window resolves
	// it from the account default and what the keys actually allow, and the
	// send refuses rather than falling back to plaintext.
	Protection string `json:"protection"`
}

// SendMessage builds the mime message and enqueues it in the durable outbox. The
// background worker transmits it when smtp credentials are configured. This call
// returns once the message is safely queued, so the ui can confirm immediately.
//
// Signing and encryption happen here, at enqueue time, not when the worker
// transmits: the queue holds the finished bytes, so a message sits in the
// outbox already protected and a passphrase is only ever needed while the user
// is still in front of the send button. Anything that would stop the protection
// from being applied fails the send outright, because the alternative is
// quietly putting a message the user marked as protected on the wire in
// plaintext.
func (a *App) SendMessage(req ComposeRequest) (int64, error) {
	if err := a.ready(); err != nil {
		return 0, err
	}

	msg, err := a.buildMessage(req)
	if err != nil {
		return 0, err
	}

	notBefore, err := resolveNotBefore(req.SendAt, a.intSetting(settingSendDelay, 0), time.Now().UTC())
	if err != nil {
		return 0, err
	}

	account, err := a.store.GetAccount(a.ctx, req.AccountID)
	if err != nil {
		return 0, err
	}
	mode, opts, engine, err := a.protectionOptions(*account, req)
	if err != nil {
		return 0, err
	}

	id, err := smtp.Enqueue(a.ctx, a.queue, req.AccountID, msg, engine, mode, opts, notBefore)
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

// RetrySend puts a failed message back in the send queue, returning whether it
// was requeued. False means it is no longer failed, which the ui treats as
// "already handled" rather than an error.
func (a *App) RetrySend(id int64) (bool, error) {
	if err := a.ready(); err != nil {
		return false, err
	}
	retried, err := a.queue.Retry(a.ctx, id)
	if err != nil {
		return false, err
	}
	if retried {
		a.emit(EventOutboxChanged, nil)
	}
	return retried, nil
}

// DiscardFailedSend removes a failed message from the outbox, returning whether
// it was removed. The message is not recoverable afterwards, so the ui confirms
// before calling this.
func (a *App) DiscardFailedSend(id int64) (bool, error) {
	if err := a.ready(); err != nil {
		return false, err
	}
	discarded, err := a.queue.Discard(a.ctx, id)
	if err != nil {
		return false, err
	}
	if discarded {
		a.emit(EventOutboxChanged, nil)
	}
	return discarded, nil
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

// resolveNotBefore decides when an outbox message becomes eligible to send. An
// explicit sendAt (RFC3339) takes precedence over the undo-send delay: a
// message the user deliberately scheduled shouldn't also sit through the short
// undo window on top of its scheduled time. sendAt must parse and be strictly
// after now, or an error is returned. With sendAt empty, the undo-send delay
// (delaySeconds, 0 meaning none) is applied as before.
func resolveNotBefore(sendAt string, delaySeconds int, now time.Time) (time.Time, error) {
	if sendAt == "" {
		if delaySeconds > 0 {
			return now.Add(time.Duration(delaySeconds) * time.Second), nil
		}
		return time.Time{}, nil
	}
	when, err := time.Parse(time.RFC3339, sendAt)
	if err != nil {
		return time.Time{}, ErrSendAtInvalid
	}
	if !when.After(now) {
		return time.Time{}, ErrSendAtPast
	}
	return when.UTC(), nil
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
// A draft of a message the user chose to encrypt is not stored as it stands:
// the point of encrypting it was that its contents are not to sit around
// readable, and an unsent draft is no different from a sent one in that
// respect. Sealed holds the whole request encrypted to the sender's own key,
// Request is left empty, and Locked says the passphrase is needed to read it
// back. Ordinary drafts are unchanged.
type DraftDTO struct {
	ID      int64          `json:"id"`
	SavedAt string         `json:"savedAt"`
	Request ComposeRequest `json:"request"`
	// Locked is true when this draft is sealed and could not be opened with a
	// passphrase Pelton currently holds.
	Locked bool `json:"locked"`
	// AccountID and Protection survive sealing so the drafts list can still say
	// which mailbox a locked draft belongs to and that it is protected.
	AccountID  int64  `json:"accountId"`
	Protection string `json:"protection"`
}

// storedDraft is a draft as it sits in the settings table. It is kept separate
// from DraftDTO so the sealed ciphertext has somewhere to live that the
// frontend never sees: the ui gets Locked and the metadata, never the armor.
type storedDraft struct {
	ID      int64          `json:"id"`
	SavedAt string         `json:"savedAt"`
	Request ComposeRequest `json:"request"`
	// Sealed is the armored, encrypted request, set only for a draft of a
	// message the user chose to encrypt. Request is empty when it is set.
	Sealed     string `json:"sealed,omitempty"`
	AccountID  int64  `json:"accountId"`
	Protection string `json:"protection"`
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

	entry := storedDraft{
		ID:         id,
		SavedAt:    time.Now().UTC().Format(time.RFC3339),
		Request:    req,
		AccountID:  req.AccountID,
		Protection: req.Protection,
	}
	// a draft of a message set to encrypt is sealed to the sender's own key
	// rather than written out as it stands. Failing to seal fails the save: the
	// alternative is quietly storing in the clear the one thing the user asked
	// not to be.
	if encrypts(req.Protection) {
		sealed, err := a.sealDraft(req)
		if err != nil {
			return 0, err
		}
		entry.Sealed = sealed
		entry.Request = ComposeRequest{AccountID: req.AccountID, Protection: req.Protection}
	}

	if id == 0 {
		entry.ID = time.Now().UnixNano()
		id = entry.ID
		drafts = append(drafts, entry)
	}
	for i := range drafts {
		if drafts[i].ID == id {
			entry.ID = id
			drafts[i] = entry
		}
	}
	if err := a.store.SetJSON(a.ctx, draftsKey, drafts); err != nil {
		return 0, err
	}
	return id, nil
}

// ListDrafts returns the locally saved drafts, newest first by save time. A
// sealed draft is opened when a passphrase for its key is already held, and
// otherwise comes back marked Locked with no content, for UnsealDraft to open.
func (a *App) ListDrafts() ([]DraftDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	drafts, err := a.loadDrafts()
	if err != nil {
		return nil, err
	}
	out := make([]DraftDTO, 0, len(drafts))
	for _, draft := range drafts {
		out = append(out, a.toDraftDTO(draft))
	}
	return out, nil
}

// toDraftDTO opens a sealed draft when a passphrase for it is already held, and
// otherwise hands back an empty one marked Locked. Nothing prompts from here:
// listing drafts must not put a passphrase dialog in front of anybody.
func (a *App) toDraftDTO(draft storedDraft) DraftDTO {
	dto := DraftDTO{
		ID:         draft.ID,
		SavedAt:    draft.SavedAt,
		Request:    draft.Request,
		AccountID:  draft.AccountID,
		Protection: draft.Protection,
	}
	if draft.Sealed == "" {
		return dto
	}
	req, err := a.unsealDraft(draft, nil)
	if err != nil {
		dto.Locked = true
		return dto
	}
	dto.Request = req
	return dto
}

// UnsealDraft opens a locked draft with a passphrase and returns it. The
// passphrase is held for the rest of the session, so a user working through
// several protected drafts is asked once.
func (a *App) UnsealDraft(id int64, passphrase string) (DraftDTO, error) {
	if err := a.ready(); err != nil {
		return DraftDTO{}, err
	}
	drafts, err := a.loadDrafts()
	if err != nil {
		return DraftDTO{}, err
	}
	for _, draft := range drafts {
		if draft.ID != id {
			continue
		}
		if draft.Sealed == "" {
			return a.toDraftDTO(draft), nil
		}
		req, err := a.unsealDraft(draft, []byte(passphrase))
		if err != nil {
			return DraftDTO{}, err
		}
		dto := a.toDraftDTO(draft)
		dto.Request = req
		dto.Locked = false
		return dto, nil
	}
	return DraftDTO{}, fmt.Errorf("pelton: draft %d not found", id)
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
func (a *App) loadDrafts() ([]storedDraft, error) {
	var drafts []storedDraft
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

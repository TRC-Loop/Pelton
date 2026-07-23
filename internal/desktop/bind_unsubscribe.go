package desktop

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/TRC-Loop/Pelton/internal/crypto"
	"github.com/TRC-Loop/Pelton/internal/smtp"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// Unsubscribe support (#71). Preference order per the List-Unsubscribe headers
// stored at sync time:
//   - RFC 8058 one-click (https target + List-Unsubscribe-Post): a background
//     POST, fired here, only ever on explicit user action.
//   - mailto target: an unsubscribe mail sent through the account's own SMTP
//     via the outbox.
//   - plain http(s) target: the frontend opens it in the system browser
//     (auto-POSTing arbitrary URLs without RFC 8058 would be unsafe).

// unsubscribe kinds as exposed to the frontend.
const (
	unsubKindOneClick = "oneclick"
	unsubKindMailto   = "mailto"
	unsubKindLink     = "link"
)

// settingUnsubscribed holds a json map of sender address -> RFC3339 time of a
// carried-out unsubscribe, so the button can show "already unsubscribed".
const settingUnsubscribed = "unsubscribed_senders"

// UnsubscribeDTO describes the unsubscribe mechanism a message advertises.
type UnsubscribeDTO struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
	// Done is true when this sender was already unsubscribed from.
	Done bool `json:"done"`
}

// unsubEntry matches one <...> entry of a List-Unsubscribe header.
var unsubEntry = regexp.MustCompile(`<([^>]+)>`)

// parseListUnsubscribe picks the preferred mechanism from a raw header value.
// It returns an empty kind when the header offers nothing usable.
func parseListUnsubscribe(header string, oneClickPost bool) (kind, target string) {
	var mailto, httpsURL, httpURL string
	for _, m := range unsubEntry.FindAllStringSubmatch(header, -1) {
		entry := strings.TrimSpace(m[1])
		switch {
		case strings.HasPrefix(strings.ToLower(entry), "mailto:") && mailto == "":
			mailto = entry
		case strings.HasPrefix(strings.ToLower(entry), "https://") && httpsURL == "":
			httpsURL = entry
		case strings.HasPrefix(strings.ToLower(entry), "http://") && httpURL == "":
			httpURL = entry
		}
	}
	// one-click requires https per RFC 8058.
	if oneClickPost && httpsURL != "" {
		return unsubKindOneClick, httpsURL
	}
	if mailto != "" {
		return unsubKindMailto, mailto
	}
	if httpsURL != "" {
		return unsubKindLink, httpsURL
	}
	if httpURL != "" {
		return unsubKindLink, httpURL
	}
	return "", ""
}

// unsubscribeInfo builds the DTO for a message, or nil when it advertises no
// unsubscribe mechanism (including rows cached before the header column
// existed; the frontend still falls back to scanning the body for a link).
func (a *App) unsubscribeInfo(m *storage.Message) *UnsubscribeDTO {
	kind, target := parseListUnsubscribe(m.ListUnsubscribe, m.ListUnsubscribePost)
	if kind == "" {
		return nil
	}
	return &UnsubscribeDTO{Kind: kind, Target: target, Done: a.unsubscribedAt(m.FromAddress) != ""}
}

// Unsubscribe carries out a message's unsubscribe mechanism: the one-click
// POST or the mailto send. Plain-link targets are opened by the frontend in
// the system browser instead, so calling this for one is an error. On success
// the sender is remembered so the button relabels.
func (a *App) Unsubscribe(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return err
	}
	kind, target := parseListUnsubscribe(m.ListUnsubscribe, m.ListUnsubscribePost)
	switch kind {
	case unsubKindOneClick:
		if err := postOneClick(a.httpClient(15*time.Second), target); err != nil {
			return err
		}
	case unsubKindMailto:
		if err := a.sendUnsubscribeMail(m.AccountID, target); err != nil {
			return err
		}
	default:
		return fmt.Errorf("pelton: message offers no unsubscribe action")
	}
	a.markUnsubscribed(m.FromAddress)
	return nil
}

// postOneClick fires the RFC 8058 one-click POST. The fixed form body is the
// whole protocol; a 2xx after redirects counts as done.
func postOneClick(client *http.Client, target string) error {
	resp, err := client.Post(target, "application/x-www-form-urlencoded",
		strings.NewReader("List-Unsubscribe=One-Click"))
	if err != nil {
		return fmt.Errorf("pelton: unsubscribe request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("pelton: unsubscribe endpoint answered %s", resp.Status)
	}
	return nil
}

// sendUnsubscribeMail queues the unsubscribe mail through the account's own
// smtp via the outbox, so it gets the same durability and retries as any send.
func (a *App) sendUnsubscribeMail(accountID int64, target string) error {
	u, err := url.Parse(target)
	if err != nil || u.Opaque == "" {
		return fmt.Errorf("pelton: invalid unsubscribe mailto target")
	}
	acc, err := a.store.GetAccount(a.ctx, accountID)
	if err != nil {
		return err
	}
	q := u.Query()
	subject := q.Get("subject")
	if subject == "" {
		subject = "unsubscribe"
	}
	body := q.Get("body")
	if body == "" {
		body = "unsubscribe"
	}
	msg := &smtp.Message{
		From:    smtp.Address{Name: acc.DisplayName, Email: acc.Email},
		To:      []smtp.Address{{Email: u.Opaque}},
		Subject: subject,
		Text:    body,
	}
	_, err = smtp.Enqueue(a.ctx, a.queue, accountID, msg, nil, crypto.ModeNone, crypto.Options{}, time.Time{})
	return err
}

// unsubscribedAt returns when the sender was unsubscribed from, or "".
func (a *App) unsubscribedAt(sender string) string {
	var done map[string]string
	if err := a.store.GetJSON(a.ctx, settingUnsubscribed, &done); err != nil {
		return ""
	}
	return done[strings.ToLower(sender)]
}

// markUnsubscribed records a carried-out unsubscribe for the sender. Failures
// only cost the "already unsubscribed" label, so they are logged, not fatal.
func (a *App) markUnsubscribed(sender string) {
	var done map[string]string
	if err := a.store.GetJSON(a.ctx, settingUnsubscribed, &done); err != nil || done == nil {
		done = map[string]string{}
	}
	done[strings.ToLower(sender)] = time.Now().UTC().Format(time.RFC3339)
	if err := a.store.SetJSON(a.ctx, settingUnsubscribed, done); err != nil {
		a.log.Error("record unsubscribe", "sender", sender, "err", err)
	}
}

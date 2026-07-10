package desktop

import (
	"fmt"
	"strings"
)

// remote image allowlist. mail remote content (images) is blocked by default to
// stop tracking pixels. the user can permanently allow it per sender or per
// sender-domain, or globally. those choices persist in the settings table and
// are consulted when a message is rendered so trusted senders load immediately
// with no banner.
const (
	settingRemoteSenders = "remote_allow_senders"
	settingRemoteDomains = "remote_allow_domains"
)

// remoteSenders returns the lowercased from-addresses the user trusts for remote
// content.
func (a *App) remoteSenders() []string {
	var out []string
	_ = a.store.GetJSON(a.ctx, settingRemoteSenders, &out)
	return out
}

// remoteDomains returns the lowercased sender domains the user trusts.
func (a *App) remoteDomains() []string {
	var out []string
	_ = a.store.GetJSON(a.ctx, settingRemoteDomains, &out)
	return out
}

// remoteAutoAllow reports whether a message from fromAddress should render remote
// content without prompting, because of the global setting, a trusted sender, or
// a trusted sender domain.
func (a *App) remoteAutoAllow(fromAddress string) bool {
	if a.boolSetting(settingRemoteAlways, false) {
		return true
	}
	addr := strings.ToLower(strings.TrimSpace(fromAddress))
	if addr == "" {
		return false
	}
	for _, s := range a.remoteSenders() {
		if s == addr {
			return true
		}
	}
	domain := emailDomain(addr)
	if domain == "" {
		return false
	}
	for _, d := range a.remoteDomains() {
		if d == domain {
			return true
		}
	}
	return false
}

// TrustSenderImages permanently allows remote content from a message's sender.
func (a *App) TrustSenderImages(messageID int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	m, err := a.store.GetMessage(a.ctx, messageID)
	if err != nil {
		return err
	}
	addr := strings.ToLower(strings.TrimSpace(m.FromAddress))
	if addr == "" {
		return nil
	}
	senders := appendUnique(a.remoteSenders(), addr)
	return a.store.SetJSON(a.ctx, settingRemoteSenders, senders)
}

// AllowDomainImages permanently allows remote content from a message sender's
// whole domain (for example all mail from example.com).
func (a *App) AllowDomainImages(messageID int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	m, err := a.store.GetMessage(a.ctx, messageID)
	if err != nil {
		return err
	}
	domain := emailDomain(strings.ToLower(strings.TrimSpace(m.FromAddress)))
	if domain == "" {
		return nil
	}
	domains := appendUnique(a.remoteDomains(), domain)
	return a.store.SetJSON(a.ctx, settingRemoteDomains, domains)
}

// ImageAllowEntryDTO is one entry in the remote-image allowlist: a trusted
// sender address or sender domain, with an example cached message (if any) so
// the ui can show the user which mail the entry came from.
type ImageAllowEntryDTO struct {
	Value            string `json:"value"`
	Kind             string `json:"kind"` // "sender" or "domain"
	ExampleMessageID int64  `json:"exampleMessageId"`
	ExampleSubject   string `json:"exampleSubject"`
	ExampleFrom      string `json:"exampleFrom"`
}

// ListImageAllowlist returns every trusted sender and domain the user has
// allowed remote content for, each paired with an example cached message.
func (a *App) ListImageAllowlist() ([]ImageAllowEntryDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	out := []ImageAllowEntryDTO{}
	for _, s := range a.remoteSenders() {
		out = append(out, a.allowEntry(s, false))
	}
	for _, d := range a.remoteDomains() {
		out = append(out, a.allowEntry(d, true))
	}
	return out, nil
}

// allowEntry builds one allowlist entry, resolving an example message for the
// sender or domain when one is cached locally.
func (a *App) allowEntry(value string, isDomain bool) ImageAllowEntryDTO {
	kind := "sender"
	if isDomain {
		kind = "domain"
	}
	entry := ImageAllowEntryDTO{Value: value, Kind: kind}
	if m, err := a.store.LatestMessageFrom(a.ctx, value, isDomain); err == nil && m != nil {
		entry.ExampleMessageID = m.ID
		entry.ExampleSubject = m.Subject
		entry.ExampleFrom = m.FromAddress
	}
	return entry
}

// RemoveImageAllow drops a trusted sender ("sender") or domain ("domain") from
// the remote-image allowlist, so its future mail is blocked again by default.
func (a *App) RemoveImageAllow(kind, value string) error {
	if err := a.ready(); err != nil {
		return err
	}
	value = strings.ToLower(strings.TrimSpace(value))
	switch kind {
	case "sender":
		return a.store.SetJSON(a.ctx, settingRemoteSenders, removeValue(a.remoteSenders(), value))
	case "domain":
		return a.store.SetJSON(a.ctx, settingRemoteDomains, removeValue(a.remoteDomains(), value))
	default:
		return fmt.Errorf("pelton: unknown allowlist kind %q", kind)
	}
}

// removeValue returns list without value, preserving order.
func removeValue(list []string, value string) []string {
	out := make([]string, 0, len(list))
	for _, v := range list {
		if v != value {
			out = append(out, v)
		}
	}
	return out
}

// emailDomain returns the domain part of an address, or empty if malformed.
func emailDomain(addr string) string {
	at := strings.LastIndex(addr, "@")
	if at < 0 || at == len(addr)-1 {
		return ""
	}
	return addr[at+1:]
}

// appendUnique adds value to the slice if it is not already present.
func appendUnique(list []string, value string) []string {
	for _, v := range list {
		if v == value {
			return list
		}
	}
	return append(list, value)
}

package desktop

import (
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

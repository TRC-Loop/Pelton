package phishing

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Kind names one thing that was found. The ui maps these to explanations, so
// they are stable identifiers rather than sentences.
type Kind string

const (
	// KindDMARCFail means the message failed the policy its own From domain
	// publishes. It is the strongest signal available.
	KindDMARCFail Kind = "dmarc_fail"
	// KindSPFFail and KindDKIMFail are the individual method failures.
	KindSPFFail  Kind = "spf_fail"
	KindDKIMFail Kind = "dkim_fail"
	// KindUnaligned means authentication passed, but for a different domain
	// than the one in the visible From. That is how a mail sent through a
	// provider the sender never used still passes spf.
	KindUnaligned Kind = "unaligned"
	// KindDisplayNameSpoof means the display name is someone the user knows,
	// or is itself an address, and the real address is a different one.
	KindDisplayNameSpoof Kind = "display_name_spoof"
	// KindReplyToMismatch means a reply would go to another organisation.
	KindReplyToMismatch Kind = "reply_to_mismatch"
	// KindLookalikeDomain means the sender domain is a near-copy of one the
	// user actually corresponds with.
	KindLookalikeDomain Kind = "lookalike_domain"
	// KindPunycodeSender means the sender domain is written in punycode.
	KindPunycodeSender Kind = "punycode_sender"
	// KindLinkTextMismatch means a link's visible text names one site and the
	// link goes to another.
	KindLinkTextMismatch Kind = "link_text_mismatch"
	// KindPunycodeLink means a link points at a punycode host.
	KindPunycodeLink Kind = "punycode_link"
	// KindShortenedLink means a link hides its destination behind a shortener.
	KindShortenedLink Kind = "shortened_link"
	// KindCredentialLink means a link looks like a sign-in page on a domain
	// that has nothing to do with the apparent sender.
	KindCredentialLink Kind = "credential_link"
)

// weight is how much a signal counts. Nothing here is a verdict on its own
// except the strong ones, and even those are worded as what was observed.
type weight int

const (
	weak weight = iota + 1
	medium
	strong
)

// Signal is one finding. Detail carries the specific value involved (a domain,
// a url) so the ui can show what it is talking about without re-deriving it.
type Signal struct {
	Kind   Kind   `json:"kind"`
	Detail string `json:"detail,omitempty"`
	// Target is the link a link-signal is about, empty for sender signals. The
	// ui marks these urls inline in the message body.
	Target string `json:"target,omitempty"`

	weight weight
}

// Level is the overall reading. The three values are deliberately not "safe":
// nothing here can establish that a message is genuine, only that nothing
// looked wrong.
const (
	// LevelNone means nothing was found.
	LevelNone = "none"
	// LevelCaution means something is odd but has ordinary explanations.
	LevelCaution = "caution"
	// LevelWarning means the message is claiming to be from someone it
	// demonstrably was not sent by.
	LevelWarning = "warning"
)

// Report is the outcome for one message.
type Report struct {
	Level   string   `json:"level"`
	Signals []Signal `json:"signals,omitempty"`
	// Links lists the urls any link signal was about, deduplicated, so the body
	// renderer can mark them without walking the signals.
	Links []string `json:"links,omitempty"`
}

// Message is everything the checks read. Anything unknown can be left empty:
// a check with nothing to work from produces no signal rather than a guess.
type Message struct {
	// From is the visible sender address; FromName its display name.
	From     string
	FromName string
	// ReplyTo is the Reply-To address, empty when the message has none.
	ReplyTo string
	// Auth is what the receiving server reported.
	Auth Auth
	// HTML and Text are the message body as stored. HTML is preferred when both
	// are present, since that is what the reader sees.
	HTML string
	Text string
	// Correspondents is who the user actually exchanges mail with: address to
	// display name, both lowercased. It is what makes the lookalike and
	// display-name checks specific to this user rather than a list of brands
	// somebody else chose.
	Correspondents map[string]string
}

// maxLinks caps the link analysis. A message with thousands of links is a
// newsletter, and the first few hundred already say whatever there is to say.
const maxLinks = 300

// shorteners are the link-hiding services common enough in mail to be worth
// naming. The list is short on purpose: it is a hint, weighted as one, not a
// blocklist, and a shortener nobody named still gets caught by the other checks
// when it is doing anything else odd.
var shorteners = map[string]bool{
	"bit.ly": true, "tinyurl.com": true, "goo.gl": true, "t.co": true,
	"ow.ly": true, "buff.ly": true, "is.gd": true, "cutt.ly": true,
	"rebrand.ly": true, "shorturl.at": true, "rb.gy": true, "tiny.cc": true,
	"lnkd.in": true, "t.ly": true, "s.id": true, "short.io": true,
}

// credentialPath matches the paths a sign-in page lives at. On the sender's own
// domain that is unremarkable, which is why the check also asks whose domain it
// is.
var credentialPath = regexp.MustCompile(`(?i)(^|[/._-])(login|signin|sign-in|log-in|verify|verification|authenticate|auth|password|reset|secure|account|confirm|validate|unlock|billing|payment)([/._-]|$)`)

// Analyse runs every check and weighs what comes back.
func Analyse(msg Message) Report {
	var signals []Signal
	signals = append(signals, senderSignals(msg)...)
	signals = append(signals, linkSignals(msg)...)

	report := Report{Level: LevelNone, Signals: signals}
	if len(signals) == 0 {
		return Report{Level: LevelNone}
	}

	seen := make(map[string]bool)
	var strongs, mediums, weaks int
	for _, s := range signals {
		switch s.weight {
		case strong:
			strongs++
		case medium:
			mediums++
		default:
			weaks++
		}
		if s.Target != "" && !seen[s.Target] {
			seen[s.Target] = true
			report.Links = append(report.Links, s.Target)
		}
	}

	switch {
	case strongs > 0 || mediums >= 2:
		report.Level = LevelWarning
	case mediums == 1 || weaks >= 2:
		report.Level = LevelCaution
	default:
		// a single weak signal is ordinary. Newsletters set a Reply-To on
		// another domain constantly, and warning about that teaches the reader
		// to click past the banner, which costs more than it saves.
		report.Level = LevelNone
	}
	return report
}

// senderSignals answers "is the sender who they claim to be".
func senderSignals(msg Message) []Signal {
	var out []Signal
	from := domainOf(msg.From)
	if from == "" {
		return nil
	}

	// authentication, but only when the server said something. Mail with no
	// Authentication-Results header, and everything cached before Pelton stored
	// them, has nothing to fail.
	if msg.Auth.Stated() {
		switch msg.Auth.DMARC {
		case ResultFail:
			out = append(out, Signal{Kind: KindDMARCFail, Detail: from, weight: strong})
		case ResultPass:
			// dmarc passing means spf or dkim passed and was aligned, so the
			// individual results below cannot say anything it has not settled.
			return append(out, nameSignals(msg, from)...)
		}
		if msg.Auth.SPF == ResultFail {
			out = append(out, Signal{Kind: KindSPFFail, Detail: msg.Auth.SPFDomain, weight: medium})
		}
		if msg.Auth.DKIM == ResultFail {
			out = append(out, Signal{Kind: KindDKIMFail, Detail: msg.Auth.DKIMDomain, weight: medium})
		}
		// alignment: something passed, but for a domain that is not the one in
		// the From line.
		if unaligned := unalignedDomain(msg.Auth, from); unaligned != "" {
			out = append(out, Signal{Kind: KindUnaligned, Detail: unaligned, weight: medium})
		}
	}

	return append(out, nameSignals(msg, from)...)
}

// nameSignals are the checks about the visible identity, which hold whatever
// authentication said: a domain the sender genuinely owns can still be a
// lookalike of somebody else's.
func nameSignals(msg Message, from string) []Signal {
	var out []Signal

	if spoofed := displayNameSpoof(msg, from); spoofed != "" {
		out = append(out, Signal{Kind: KindDisplayNameSpoof, Detail: spoofed, weight: strong})
	}
	if reply := domainOf(msg.ReplyTo); reply != "" && !sameOrg(reply, from) {
		out = append(out, Signal{Kind: KindReplyToMismatch, Detail: reply, weight: weak})
	}
	if punycode(from) {
		out = append(out, Signal{Kind: KindPunycodeSender, Detail: unicodeDomain(from), weight: medium})
	}
	if known, strength := lookalikeOf(from, msg.Correspondents); known != "" {
		out = append(out, Signal{Kind: KindLookalikeDomain, Detail: known, weight: strength})
	}
	return out
}

// unalignedDomain returns the authenticated domain when it belongs to another
// organisation than the visible From, or empty when things line up or nothing
// passed to compare.
func unalignedDomain(auth Auth, from string) string {
	if auth.DKIM == ResultPass && auth.DKIMDomain != "" {
		if sameOrg(auth.DKIMDomain, from) {
			return ""
		}
		// a passing signature from elsewhere is only worth mentioning when spf
		// does not vouch for the From domain either.
		if auth.SPF == ResultPass && sameOrg(auth.SPFDomain, from) {
			return ""
		}
		return auth.DKIMDomain
	}
	if auth.SPF == ResultPass && auth.SPFDomain != "" && !sameOrg(auth.SPFDomain, from) {
		return auth.SPFDomain
	}
	return ""
}

// displayNameSpoof returns the address being impersonated, or empty. Two
// shapes: the display name is itself an address that is not the sender, or it
// is the name of somebody the user corresponds with at a different address.
func displayNameSpoof(msg Message, from string) string {
	name := strings.TrimSpace(msg.FromName)
	if name == "" {
		return ""
	}
	sender := strings.ToLower(strings.TrimSpace(msg.From))

	if inner := addressIn(name); inner != "" && inner != sender && !sameOrg(domainOf(inner), from) {
		return inner
	}

	folded := foldName(name)
	if folded == "" {
		return ""
	}
	for address, known := range msg.Correspondents {
		if foldName(known) != folded {
			continue
		}
		if strings.EqualFold(address, sender) {
			return ""
		}
		// somebody the user knows, writing from a different organisation. A
		// colleague moving jobs looks like this too, which is why the ui says
		// what it saw rather than what it thinks.
		if !sameOrg(domainOf(address), from) {
			return address
		}
	}
	return ""
}

// addressPattern finds an email address embedded in a display name.
var addressPattern = regexp.MustCompile(`[\w.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[\w-]+(\.[\w-]+)+`)

// addressIn returns the address a display name contains, lowercased, or empty.
func addressIn(name string) string {
	return strings.ToLower(addressPattern.FindString(name))
}

// foldName reduces a display name to something comparable: lowercase, no
// punctuation, single spaces. "Arne K." and "arne k" are the same person.
func foldName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r > 127:
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// lookalikeOf returns the correspondent domain a sender is imitating and how
// closely, or empty. The closest resemblance found wins, so one weak match
// among the user's contacts cannot mask a real copy of another.
func lookalikeOf(from string, correspondents map[string]string) (string, weight) {
	var bestDomain string
	var best weight
	for address := range correspondents {
		known := domainOf(address)
		if strength := lookalikeStrength(from, known); strength > best {
			bestDomain, best = registrable(known), strength
		}
	}
	return bestDomain, best
}

// linkSignals answers "are the links genuine".
func linkSignals(msg Message) []Signal {
	body := msg.HTML
	links := parseAnchors(body)
	if body == "" {
		links = plainLinks(msg.Text)
	}
	if len(links) == 0 {
		return nil
	}
	from := domainOf(msg.From)

	var out []Signal
	for _, link := range links {
		host := hostOf(link.href)
		if host == "" {
			continue
		}
		// the visible text names a site, and it is not where the link goes.
		if textHost := hostOf(link.text); textHost != "" && !sameOrg(textHost, host) {
			out = append(out, Signal{
				Kind:   KindLinkTextMismatch,
				Detail: fmt.Sprintf("%s -> %s", textHost, host),
				Target: link.href,
				weight: medium,
			})
			continue
		}
		if punycode(host) {
			out = append(out, Signal{
				Kind: KindPunycodeLink, Detail: unicodeDomain(host), Target: link.href, weight: medium,
			})
			continue
		}
		if shorteners[registrable(host)] {
			out = append(out, Signal{
				Kind: KindShortenedLink, Detail: host, Target: link.href, weight: weak,
			})
			continue
		}
		// a sign-in page somewhere unrelated to whoever this claims to be from.
		if credentialPath.MatchString(pathOf(link.href)) && !sameOrg(host, from) {
			out = append(out, Signal{
				Kind: KindCredentialLink, Detail: host, Target: link.href, weight: medium,
			})
		}
	}
	return out
}

// anchor is one link with the text the reader sees on it.
type anchor struct {
	href string
	text string
}

// anchorPattern captures an anchor's href and its inner text. The body has
// already been through the sanitizer by the time this runs, so the markup is
// well formed enough for this to be the right amount of parsing: a full tree
// walk would cost more and find the same hrefs.
var anchorPattern = regexp.MustCompile(`(?is)<a\s[^>]*href\s*=\s*["']?([^"'\s>]+)["']?[^>]*>(.*?)</a>`)

// tagPattern strips markup from an anchor's inner text.
var tagPattern = regexp.MustCompile(`(?s)<[^>]*>`)

// parseAnchors pulls the links out of an html body, newest markup first, capped.
func parseAnchors(html string) []anchor {
	matches := anchorPattern.FindAllStringSubmatch(html, maxLinks)
	out := make([]anchor, 0, len(matches))
	for _, m := range matches {
		href := strings.TrimSpace(m[1])
		if !isHTTP(href) {
			continue
		}
		out = append(out, anchor{href: href, text: strings.TrimSpace(tagPattern.ReplaceAllString(m[2], ""))})
	}
	return out
}

// urlPattern finds bare urls in plain text.
var urlPattern = regexp.MustCompile(`https?://[^\s<>"')]+`)

// plainLinks reads the links out of a plain-text body. There is no anchor text
// to disagree with the target, so only the host checks apply.
func plainLinks(text string) []anchor {
	found := urlPattern.FindAllString(text, maxLinks)
	out := make([]anchor, 0, len(found))
	for _, href := range found {
		out = append(out, anchor{href: href})
	}
	return out
}

// isHTTP reports whether a href is a web link, as opposed to mailto, cid, tel
// or a fragment, none of which can mislead about a destination.
func isHTTP(href string) bool {
	lower := strings.ToLower(href)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// hostOf returns the host of a url, or of a bare domain written as link text.
// Empty when the string is not a location at all, which is the common case for
// anchor text and is not itself suspicious.
func hostOf(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if isHTTP(s) {
		u, err := url.Parse(s)
		if err != nil {
			return ""
		}
		return strings.ToLower(u.Hostname())
	}
	// link text like "paypal.com" or "www.paypal.com/account" names a host
	// without a scheme. Anything with a space in it is prose, not a location.
	if strings.ContainsAny(s, " \t\n") {
		return ""
	}
	host, _, _ := strings.Cut(s, "/")
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if !strings.Contains(host, ".") || strings.Contains(host, "@") {
		return ""
	}
	if _, err := publicSuffixCheck(host); err != nil {
		return ""
	}
	return host
}

// pathOf returns the path and query of a url, where a sign-in page shows.
func pathOf(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return u.EscapedPath() + "?" + u.RawQuery
}

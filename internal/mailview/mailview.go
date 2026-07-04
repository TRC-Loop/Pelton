// Package mailview prepares stored mail for safe display in the ui. It owns the
// html sanitization policy, inline image (cid) resolution, a best effort pgp
// status probe and the list snippet. Keeping this here means app.go only
// orchestrates and never embeds rendering or security logic.
//
// Security note: the sanitizer is the single trusted boundary between untrusted
// mail html and the renderer. By default it strips remote resource urls so a
// message cannot phone home (tracking pixels). The caller asks for the remote
// variant explicitly when the user clicks "load remote images".
package mailview

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// PGPStatus is the detected protection state of a received message. It is a best
// effort read of the body, not a verification: it reports what the message
// claims, which the ui shows as an indicator.
type PGPStatus string

const (
	// PGPNone means no pgp markers were found.
	PGPNone PGPStatus = "none"
	// PGPSigned means an inline pgp signed block was found.
	PGPSigned PGPStatus = "signed"
	// PGPEncrypted means an inline pgp encrypted message block was found.
	PGPEncrypted PGPStatus = "encrypted"
)

// snippetLen is how many characters of plain text the list preview keeps.
const snippetLen = 140

// remoteURLPattern finds http(s) resource urls in src/href/style positions. It
// is used only to decide whether a "load remote content" affordance is needed,
// not for sanitization (bluemonday does the actual stripping by url scheme).
var remoteURLPattern = regexp.MustCompile(`(?i)(src|background|href)\s*=\s*["']?\s*https?://|url\(\s*["']?\s*https?://`)

// cidRefPattern matches a cid: url referencing an inline attachment by its
// content id, with or without surrounding quotes.
var cidRefPattern = regexp.MustCompile(`(?i)cid:([^"'>\s)]+)`)

// blockRemotePolicy and allowRemotePolicy are built once. They share the ugc
// base; the only difference is which url schemes images and links may use. With
// remote blocked, only data: and cid: image sources survive, so http(s) images
// (and tracking pixels) are dropped entirely.
var (
	blockRemotePolicy = buildPolicy(false)
	allowRemotePolicy = buildPolicy(true)
)

// buildPolicy returns a sanitizer policy. allowRemote decides whether remote
// http(s) resources are permitted; when false only inline data: and cid: image
// sources are kept.
func buildPolicy(allowRemote bool) *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	// allow images with sizing/alt but constrain their source scheme below.
	p.AllowImages()
	p.AllowAttrs("width", "height", "alt", "title").OnElements("img")

	// keep simple inline styling that mail relies on, without url() vectors.
	p.AllowAttrs("style").Globally()
	p.AllowStyles("color", "background-color", "text-align", "font-weight",
		"font-style", "text-decoration", "font-size", "margin", "padding").Globally()

	// tables are heavily used by html mail.
	p.AllowTables()

	if allowRemote {
		p.AllowURLSchemes("http", "https", "mailto", "data", "cid")
	} else {
		// no http(s): remote images and remote links by url get stripped. cid and
		// data stay so inline images we resolve ourselves still render.
		p.AllowURLSchemes("mailto", "data", "cid")
	}

	// open links in a new context and never leak the referrer.
	p.RequireNoReferrerOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	return p
}

// Sanitize cleans untrusted mail html for display. When allowRemote is false
// (the default for first view) remote resources are removed. The result is safe
// to inject into a sandboxed renderer.
func Sanitize(html string, allowRemote bool) string {
	if allowRemote {
		return allowRemotePolicy.Sanitize(html)
	}
	return blockRemotePolicy.Sanitize(html)
}

// HasRemoteContent reports whether the raw html references any remote http(s)
// resource, so the ui can show the "load remote content" affordance only when it
// would actually do something.
func HasRemoteContent(html string) bool {
	return remoteURLPattern.MatchString(html)
}

// remoteHostPattern captures the host of any http(s) url in the body.
var remoteHostPattern = regexp.MustCompile(`(?i)https?://([^/"'\s)>]+)`)

// RemoteHosts returns the unique hosts referenced by remote http(s) urls in the
// body, so the blocked-images banner can show the user exactly where the content
// would be loaded from. The list is capped to keep the banner readable.
func RemoteHosts(html string) []string {
	matches := remoteHostPattern.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool)
	out := make([]string, 0, 4)
	for _, m := range matches {
		host := strings.ToLower(strings.TrimSpace(m[1]))
		// drop any leading userinfo and trailing port for a clean host.
		if at := strings.LastIndex(host, "@"); at >= 0 {
			host = host[at+1:]
		}
		if colon := strings.IndexByte(host, ':'); colon >= 0 {
			host = host[:colon]
		}
		if host == "" || seen[host] {
			continue
		}
		seen[host] = true
		out = append(out, host)
		if len(out) >= 8 {
			break
		}
	}
	return out
}

// ResolveCIDs rewrites cid: references to the data urls of inline attachments.
// byContentID maps a bare content id (no angle brackets) to a full data url. A
// cid with no matching attachment is left untouched.
func ResolveCIDs(html string, byContentID map[string]string) string {
	if len(byContentID) == 0 {
		return html
	}
	return cidRefPattern.ReplaceAllStringFunc(html, func(match string) string {
		id := strings.TrimPrefix(strings.ToLower(match), "cid:")
		if url, ok := byContentID[id]; ok {
			return url
		}
		return match
	})
}

// ReferencedCIDs returns the set of content ids actually referenced by cid: urls
// in the html, lowercased and without angle brackets. The ui uses it to decide
// which attachments are truly inline (shown in the body) versus real downloadable
// attachments that merely happen to carry a content id. Without this check a
// normal pdf with a content id would be wrongly hidden from the attachment list.
func ReferencedCIDs(html string) map[string]bool {
	out := make(map[string]bool)
	for _, match := range cidRefPattern.FindAllString(html, -1) {
		id := strings.TrimPrefix(strings.ToLower(match), "cid:")
		out[id] = true
	}
	return out
}

// DetectPGP probes the stored bodies for inline pgp markers. It does not verify
// signatures or decrypt; it only reports what is present so the ui shows an
// honest indicator. Proper verification belongs to the crypto layer in a later
// step.
func DetectPGP(plain, html string) PGPStatus {
	hay := plain + "\n" + html
	switch {
	case strings.Contains(hay, "-----BEGIN PGP MESSAGE-----"):
		return PGPEncrypted
	case strings.Contains(hay, "-----BEGIN PGP SIGNED MESSAGE-----"),
		strings.Contains(hay, "-----BEGIN PGP SIGNATURE-----"):
		return PGPSigned
	default:
		return PGPNone
	}
}

// Snippet returns a short plain text preview for the message list. It prefers
// the plain body and falls back to stripping tags from the html body.
func Snippet(plain, html string) string {
	text := strings.TrimSpace(plain)
	if text == "" && html != "" {
		text = strings.TrimSpace(bluemonday.StrictPolicy().Sanitize(html))
	}
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > snippetLen {
		cut := snippetLen
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		return strings.TrimSpace(text[:cut]) + "…"
	}
	return text
}

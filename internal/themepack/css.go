package themepack

import (
	"regexp"
	"strings"
)

// CSSFile is one stylesheet from a container, with the remote references
// (url() to a network scheme, or any @import) found in it. RemoteRefs drive
// the import warning; blocking an import strips them (see StripRemote).
type CSSFile struct {
	Path       string   `json:"path"`
	Content    string   `json:"content"`
	RemoteRefs []string `json:"remoteRefs"`
}

var (
	urlRefPattern = regexp.MustCompile(`(?i)url\(\s*(?:"([^"]*)"|'([^']*)'|([^)"']*))\s*\)`)
	// any @import is treated as remote: the engine concatenates the declared
	// css files itself, so even a relative @import would not resolve.
	importPattern = regexp.MustCompile(`(?i)@import\b[^;]*;?`)
)

// scanCSS finds every remote reference in a stylesheet. Relative paths (the
// bundled-assets mechanism) and data: uris are not remote.
func scanCSS(content string) []string {
	var refs []string
	for _, m := range urlRefPattern.FindAllStringSubmatch(content, -1) {
		if ref := m[1] + m[2] + m[3]; isRemoteRef(ref) {
			refs = append(refs, strings.TrimSpace(ref))
		}
	}
	for _, imp := range importPattern.FindAllString(content, -1) {
		refs = append(refs, strings.TrimSpace(imp))
	}
	return refs
}

// StripRemote removes every remote reference from a stylesheet: remote url()
// values become the empty data: uri (never fetched) and @import statements
// are dropped entirely. Local asset references and data: uris pass through.
func StripRemote(content string) string {
	content = urlRefPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := urlRefPattern.FindStringSubmatch(match)
		if isRemoteRef(sub[1] + sub[2] + sub[3]) {
			return `url("data:,")`
		}
		return match
	})
	return importPattern.ReplaceAllString(content, "")
}

// isRemoteRef reports whether a url() value would cause a network request:
// any scheme other than data:, or a protocol-relative // prefix.
func isRemoteRef(ref string) bool {
	ref = strings.ToLower(strings.TrimSpace(ref))
	if ref == "" || strings.HasPrefix(ref, "data:") {
		return false
	}
	if strings.HasPrefix(ref, "//") {
		return true
	}
	if i := strings.Index(ref, ":"); i > 0 {
		scheme := ref[:i]
		for _, r := range scheme {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '+' && r != '-' && r != '.' {
				return false
			}
		}
		return true
	}
	return false
}

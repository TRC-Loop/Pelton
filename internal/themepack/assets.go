package themepack

import (
	"encoding/base64"
	"path"
	"strings"
)

// assetMIME maps bundled-asset extensions to the mime type used when inlining
// them as data: uris. Unknown extensions inline as octet-stream, which
// browsers handle fine for fonts and images alike.
var assetMIME = map[string]string{
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".webp":  "image/webp",
	".svg":   "image/svg+xml",
}

// AppliedCSS concatenates the package's stylesheets in manifest order and
// inlines every relative url() reference as a data: uri read from the
// package's own files. This is how bundled fonts and images work without the
// webview ever touching disk paths or the network. Remote references (which
// only exist if the user allowed them at import) pass through untouched;
// references to files missing from the package become empty data: uris.
func (p *Package) AppliedCSS() string {
	var b strings.Builder
	for _, f := range p.CSSFiles {
		b.WriteString(inlineAssets(f.Content, p.Files))
		b.WriteString("\n")
	}
	return b.String()
}

// inlineAssets rewrites relative url() references against the given file map.
func inlineAssets(css string, files map[string][]byte) string {
	return urlRefPattern.ReplaceAllStringFunc(css, func(match string) string {
		sub := urlRefPattern.FindStringSubmatch(match)
		ref := strings.TrimSpace(sub[1] + sub[2] + sub[3])
		if ref == "" || isRemoteRef(ref) || strings.HasPrefix(strings.ToLower(ref), "data:") {
			return match
		}
		content, ok := files[normalizePath(ref)]
		if !ok || len(content) > maxAssetInline {
			return `url("data:,")`
		}
		mime := assetMIME[strings.ToLower(path.Ext(ref))]
		if mime == "" {
			mime = "application/octet-stream"
		}
		return `url("data:` + mime + `;base64,` + base64.StdEncoding.EncodeToString(content) + `")`
	})
}

// PreviewDataURI returns the theme's preview image as a data: uri, or ""
// when the manifest declares none or the file is missing/oversized.
func (p *Package) PreviewDataURI() string {
	if p.Manifest.Preview == "" {
		return ""
	}
	content, ok := p.Files[normalizePath(p.Manifest.Preview)]
	if !ok || len(content) > maxAssetInline {
		return ""
	}
	mime := assetMIME[strings.ToLower(path.Ext(p.Manifest.Preview))]
	if mime == "" {
		return ""
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(content)
}

package themepack

import (
	"fmt"
)

// Package is a fully parsed and validated .peltontheme container, still in
// memory. Files holds every safe entry by its normalized zip path; the typed
// fields hold the validated views of the manifest-referenced parts.
type Package struct {
	Manifest Manifest
	Tokens   map[string]string
	CSSFiles []CSSFile
	Icons    map[string]string
	Files    map[string][]byte
}

// RemoteRefs collects every remote reference across all css files, for the
// import warning.
func (p *Package) RemoteRefs() []string {
	var refs []string
	for _, f := range p.CSSFiles {
		refs = append(refs, f.RemoteRefs...)
	}
	return refs
}

// ReadContainer parses a .peltontheme zip and validates everything the
// manifest references. The returned package has had all security checks
// applied except the remote-css decision, which is the user's (see the
// blockRemote parameter of Install).
func ReadContainer(data []byte) (*Package, error) {
	if len(data) > maxContainerBytes {
		return nil, fmt.Errorf("theme file larger than %d MB", maxContainerBytes>>20)
	}
	files, err := readZipEntries(data)
	if err != nil {
		return nil, err
	}
	manifestData, ok := files["manifest.json"]
	if !ok {
		return nil, fmt.Errorf("container has no manifest.json")
	}
	m, err := parseManifest(manifestData)
	if err != nil {
		return nil, err
	}
	p := &Package{Manifest: m, Files: files}
	if err := p.loadTokens(); err != nil {
		return nil, err
	}
	if err := p.loadCSS(); err != nil {
		return nil, err
	}
	if err := p.loadIcons(); err != nil {
		return nil, err
	}
	if m.Preview != "" {
		if _, ok := files[normalizePath(m.Preview)]; !ok {
			return nil, fmt.Errorf("preview %q missing from container", m.Preview)
		}
	}
	return p, nil
}

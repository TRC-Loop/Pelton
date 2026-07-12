package themepack

import (
	"encoding/json"
	"fmt"
	"maps"
)

// file returns a manifest-referenced file's content or an error naming it.
func (p *Package) file(ref string) ([]byte, error) {
	content, ok := p.Files[normalizePath(ref)]
	if !ok {
		return nil, fmt.Errorf("%s missing from container", ref)
	}
	return content, nil
}

// loadTokens merges the manifest's token files (or inline object) in order
// and validates the result against the allowlist.
func (p *Package) loadTokens() error {
	paths, inline, err := p.Manifest.tokenPaths()
	if err != nil {
		return err
	}
	merged := make(map[string]string)
	for _, ref := range paths {
		content, err := p.file(ref)
		if err != nil {
			return err
		}
		var tokens map[string]string
		if err := json.Unmarshal(content, &tokens); err != nil {
			return fmt.Errorf("%s: %w", ref, err)
		}
		maps.Copy(merged, tokens)
	}
	maps.Copy(merged, inline)
	p.Tokens, err = validateTokens(merged)
	return err
}

// loadCSS reads the manifest's stylesheets in order, enforcing the total css
// cap and recording each file's remote references for the import warning.
func (p *Package) loadCSS() error {
	var total int
	for _, ref := range p.Manifest.CSS {
		content, err := p.file(ref)
		if err != nil {
			return err
		}
		total += len(content)
		if total > maxCSSTotalBytes {
			return fmt.Errorf("css files exceed %d KB together", maxCSSTotalBytes>>10)
		}
		p.CSSFiles = append(p.CSSFiles, CSSFile{
			Path:       normalizePath(ref),
			Content:    string(content),
			RemoteRefs: scanCSS(string(content)),
		})
	}
	return nil
}

// loadIcons validates every icon override: well-formed name, safe svg.
func (p *Package) loadIcons() error {
	if len(p.Manifest.Icons) == 0 {
		return nil
	}
	p.Icons = make(map[string]string, len(p.Manifest.Icons))
	for name, ref := range p.Manifest.Icons {
		if err := checkIconName(name); err != nil {
			return err
		}
		content, err := p.file(ref)
		if err != nil {
			return err
		}
		if err := checkSVG(content); err != nil {
			return fmt.Errorf("icon %s (%s): %w", name, ref, err)
		}
		p.Icons[name] = string(content)
	}
	return nil
}

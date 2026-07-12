package themepack

import (
	"fmt"
	"strings"
)

// tokenAllowlist is the themeable token surface: the tokens.css contract
// minus spacing/density (density is its own setting; themes overriding row
// and pane spacing would fight it and break layouts).
var tokenAllowlist = map[string]bool{
	// surfaces
	"surface-base": true, "surface-raised": true, "surface-overlay": true,
	"surface-sunken": true, "surface-hover": true,
	"selection-bg": true, "selection-bg-strong": true,
	// text
	"text-primary": true, "text-secondary": true, "text-tertiary": true,
	"text-inverse": true, "link": true,
	// borders
	"border-subtle": true, "border-default": true, "border-strong": true,
	"hairline": true,
	// accent
	"accent": true, "accent-fg": true,
	// semantic
	"success": true, "success-bg": true, "warning": true, "warning-bg": true,
	"danger": true, "danger-bg": true,
	// radii
	"radius-control": true, "radius-card": true, "radius-none": true,
	// fonts and type
	"font-ui": true, "font-mono": true,
	"fz-meta": true, "fz-label": true, "fz-list": true, "fz-body": true,
	"fz-heading": true, "fz-title": true,
	"fw-regular": true, "fw-medium": true, "fw-semibold": true, "fw-bold": true,
	// elevation
	"shadow-overlay": true,
}

// validateTokens checks a token map against the allowlist and the safe value
// charset, returning a cleaned copy. Unknown token names are rejected rather
// than ignored so a typo in a theme fails loudly at import, not silently.
func validateTokens(tokens map[string]string) (map[string]string, error) {
	clean := make(map[string]string, len(tokens))
	for name, value := range tokens {
		name = strings.TrimPrefix(strings.TrimSpace(name), "--")
		if !tokenAllowlist[name] {
			return nil, fmt.Errorf("token %q is not themeable", name)
		}
		if err := checkTokenValue(value); err != nil {
			return nil, fmt.Errorf("token %q: %w", name, err)
		}
		clean[name] = strings.TrimSpace(value)
	}
	return clean, nil
}

// checkTokenValue rejects values that could escape a css declaration or smuggle
// in a fetch. Allows everything colors, font stacks, sizes and shadows need
// (color-mix(), quotes, commas, percentages).
func checkTokenValue(value string) error {
	if len(value) > maxTokenValueLen {
		return fmt.Errorf("value too long")
	}
	if strings.ContainsAny(value, ";{}<>@\\") {
		return fmt.Errorf("value contains forbidden characters")
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("value contains control characters")
		}
	}
	if strings.Contains(strings.ToLower(value), "url(") {
		return fmt.Errorf("url() is not allowed in token values")
	}
	return nil
}

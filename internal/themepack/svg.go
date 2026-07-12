package themepack

import (
	"fmt"
	"regexp"
	"strings"
)

// svg icon overrides are injected inline into the dom (so currentColor works),
// which makes them script-capable in principle. Icons are small and fully
// author-controlled, so anything dangerous is rejected with a clear error
// instead of silently stripped - a theme author should fix the file.
var svgForbidden = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<\s*script`),
	regexp.MustCompile(`(?i)\son[a-z]+\s*=`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)<\s*(foreignobject|iframe|embed|object|use|animate)`),
	// href/xlink:href with any scheme or protocol-relative target loads
	// external content; fragment and relative refs would be meaningless in an
	// inlined icon anyway, so all href attributes are refused.
	regexp.MustCompile(`(?i)(xlink:)?href\s*=`),
	regexp.MustCompile(`(?i)url\s*\(`),
	regexp.MustCompile(`(?i)@import`),
}

var svgRootPattern = regexp.MustCompile(`(?is)\A(?:\s|<\?xml[^>]*\?>|<!--.*?-->)*<svg[\s>]`)

// checkSVG validates an icon override file for inline injection. It returns a
// descriptive error naming what was found so the import preview can show it.
func checkSVG(content []byte) error {
	if len(content) > maxSVGBytes {
		return fmt.Errorf("svg larger than %d bytes", maxSVGBytes)
	}
	s := string(content)
	if !svgRootPattern.MatchString(s) {
		return fmt.Errorf("not an svg document")
	}
	if strings.Contains(s, "<!DOCTYPE") || strings.Contains(s, "<!ENTITY") {
		return fmt.Errorf("doctype/entity declarations are not allowed")
	}
	for _, re := range svgForbidden {
		if loc := re.FindString(s); loc != "" {
			return fmt.Errorf("forbidden content %q", strings.TrimSpace(loc))
		}
	}
	return nil
}

// iconNamePattern is the shape of an icon override name: the app's icon names
// (tabler names without the Icon prefix), lowercase kebab.
var iconNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// checkIconName validates an icons map key. Unknown-but-well-formed names are
// accepted (forward compatibility with icons the app does not use yet);
// malformed ones are not.
func checkIconName(name string) error {
	if !iconNamePattern.MatchString(name) {
		return fmt.Errorf("icon name %q must be a lowercase slug", name)
	}
	return nil
}

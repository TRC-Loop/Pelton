package themepack

import (
	"fmt"
	"strconv"
	"strings"
)

// CompatWarning reports why a theme may not display correctly on the running
// app version, or "" when it is fine. It never blocks anything: outside-range
// themes still install and apply. Unparsable app versions (dev and test
// builds) always pass so development never trips the warning.
func CompatWarning(r *VersionRange, appVersion string) string {
	if r == nil || (r.Min == "" && r.Max == "") {
		return ""
	}
	app, ok := parseVersion(appVersion)
	if !ok {
		return ""
	}
	if r.Min != "" {
		if min, ok := parseVersion(r.Min); ok && compareVersions(app, min) < 0 {
			return fmt.Sprintf("made for Pelton %s or newer", r.Min)
		}
	}
	if r.Max != "" {
		// a max like "1.1" means any 1.1.x: compare only the segments given.
		if max, ok := parseVersion(r.Max); ok && compareVersions(app[:minInt(len(app), len(max))], max) > 0 {
			return fmt.Sprintf("made for Pelton up to %s", r.Max)
		}
	}
	return ""
}

// parseVersion reads a dotted numeric version, tolerating a leading "v".
// Anything else (dev builds, test-<hash>) reports ok=false.
func parseVersion(s string) ([]int, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if s == "" {
		return nil, false
	}
	parts := strings.Split(s, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}
		nums = append(nums, n)
	}
	return nums, true
}

// compareVersions orders two parsed versions; missing segments count as 0.
func compareVersions(a, b []int) int {
	for i := range maxInt(len(a), len(b)) {
		av, bv := 0, 0
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av != bv {
			if av < bv {
				return -1
			}
			return 1
		}
	}
	return 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

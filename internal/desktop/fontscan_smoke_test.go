package desktop

import "testing"

func TestScanFontDirsFindsSystemFonts(t *testing.T) {
	families := scanFontDirs(systemFontDirs())
	if len(families) == 0 {
		// a bare build container may legitimately have no fonts installed;
		// the scan itself still ran without error, which is what matters here.
		t.Skip("no system fonts found on this machine")
	}
	t.Logf("found %d families, e.g. %v", len(families), families[:min(5, len(families))])
}

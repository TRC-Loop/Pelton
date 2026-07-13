package desktop

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"
)

// System font enumeration for the reader/compose font dropdowns. Family names
// are read directly from the font files' own name tables, so nothing shells
// out (no fc-list) and nothing touches the network.

// fontScanOnce caches the scan for the process lifetime: installed fonts do
// not change mid-session, and the first scan reads every font file once.
var (
	fontScanOnce sync.Once
	fontFamilies []string
)

// maxFontFileBytes skips pathological font files; CJK mega-collections top
// out well under this.
const maxFontFileBytes = 64 << 20

// ListSystemFonts returns the family names of the fonts installed on this
// machine, sorted and deduplicated.
func (a *App) ListSystemFonts() ([]string, error) {
	fontScanOnce.Do(func() {
		fontFamilies = scanFontDirs(systemFontDirs())
	})
	return fontFamilies, nil
}

// systemFontDirs lists the standard font locations per platform, user dirs
// included.
func systemFontDirs() []string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			filepath.Join(home, "Library", "Fonts"),
		}
	case "windows":
		dirs := []string{filepath.Join(os.Getenv("WINDIR"), "Fonts")}
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			dirs = append(dirs, filepath.Join(local, "Microsoft", "Windows", "Fonts"))
		}
		return dirs
	default: // linux and friends
		return []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(home, ".local", "share", "fonts"),
			filepath.Join(home, ".fonts"),
		}
	}
}

// scanFontDirs walks the given directories and collects every distinct font
// family. Unreadable files and non-sfnt formats are simply skipped; a font
// list is best-effort by nature.
func scanFontDirs(dirs []string) []string {
	seen := make(map[string]bool)
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".ttf", ".otf", ".ttc", ".otc":
			default:
				return nil
			}
			if info, err := d.Info(); err != nil || info.Size() > maxFontFileBytes {
				return nil
			}
			for _, family := range fontFileFamilies(path) {
				seen[family] = true
			}
			return nil
		})
	}
	families := make([]string, 0, len(seen))
	for f := range seen {
		families = append(families, f)
	}
	sort.Slice(families, func(i, j int) bool {
		return strings.ToLower(families[i]) < strings.ToLower(families[j])
	})
	return families
}

// fontFileFamilies extracts the family name(s) from one font file, handling
// both single fonts and collections.
func fontFileFamilies(path string) []string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil
	}
	collection, err := sfnt.ParseCollection(data)
	if err != nil {
		return nil
	}
	var families []string
	var buf sfnt.Buffer
	for i := 0; i < collection.NumFonts(); i++ {
		f, err := collection.Font(i)
		if err != nil {
			continue
		}
		// the typographic family (name id 16) groups styles under one name
		// ("Helvetica Neue" instead of "Helvetica Neue Light"); fall back to
		// the plain family when it is absent.
		name, err := f.Name(&buf, sfnt.NameIDTypographicFamily)
		if err != nil || name == "" {
			name, err = f.Name(&buf, sfnt.NameIDFamily)
		}
		if err == nil && name != "" && !strings.HasPrefix(name, ".") {
			families = append(families, name)
		}
	}
	return families
}

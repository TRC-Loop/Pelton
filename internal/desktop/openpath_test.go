package desktop

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestOpenPathRejectsAnEmptyPath(t *testing.T) {
	if err := openPath(""); !errors.Is(err, errNothingToOpen) {
		t.Fatalf("openPath(%q) = %v, want errNothingToOpen", "", err)
	}
}

// A path that is not there has to come back as an error the ui can show. The
// call this replaced could not fail at all: wails dropped the url and the
// button simply went dead.
func TestOpenPathReportsAPathThatIsNotThere(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	err := openPath(missing)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("openPath(%q) = %v, want a not-exist error", missing, err)
	}
}

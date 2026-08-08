package desktop

import (
	"errors"
	"testing"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

func TestRenamedPath(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		to    string
		delim string
		want  string
	}{
		{name: "root level", path: "Projects", to: "Work", delim: "/", want: "Work"},
		{name: "nested keeps its parent", path: "Work/2026/Q1", to: "Q2", delim: "/", want: "Work/2026/Q2"},
		{name: "dot delimiter", path: "INBOX.Archive.Old", to: "Older", delim: ".", want: "INBOX.Archive.Older"},
		// a flat server has no hierarchy, so the whole name is the path.
		{name: "flat server", path: "Projects", to: "Work", delim: "", want: "Work"},
		// the delimiter appearing inside a name segment must not be treated as a
		// level, which is why only the last one splits.
		{name: "name containing the delimiter elsewhere", path: "a/b/c", to: "d", delim: "/", want: "a/b/d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renamedPath(tt.path, tt.to, tt.delim); got != tt.want {
				t.Errorf("renamedPath(%q, %q, %q) = %q, want %q",
					tt.path, tt.to, tt.delim, got, tt.want)
			}
		})
	}
}

func TestProtectSpecialFolder(t *testing.T) {
	tests := []struct {
		name      string
		folder    storage.Folder
		protected bool
	}{
		{
			name:      "inbox by name",
			folder:    storage.Folder{Name: "INBOX", IMAPPath: "INBOX"},
			protected: true,
		},
		{
			name:      "sent by special-use attribute",
			folder:    storage.Folder{Name: "Verzonden", IMAPPath: "Verzonden", Attributes: []string{"\\Sent"}},
			protected: true,
		},
		{
			name:      "trash by name",
			folder:    storage.Folder{Name: "Trash", IMAPPath: "Trash"},
			protected: true,
		},
		{
			name:   "an ordinary folder is fair game",
			folder: storage.Folder{Name: "Projects", IMAPPath: "Projects"},
		},
		{
			// the role fallback matches names exactly, so this is not detected as
			// an archive and stays renameable. Tracked in #186; asserted here so
			// the behavior is deliberate rather than a surprise.
			name:   "localized archive is not detected as special",
			folder: storage.Folder{Name: "Archiv", IMAPPath: "Archiv"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := protectSpecialFolder(tt.folder)
			if tt.protected && !errors.Is(err, errFolderProtected) {
				t.Errorf("protectSpecialFolder(%q) = %v, want errFolderProtected", tt.folder.Name, err)
			}
			if !tt.protected && err != nil {
				t.Errorf("protectSpecialFolder(%q) = %v, want nil", tt.folder.Name, err)
			}
		})
	}
}

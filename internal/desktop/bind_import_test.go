package desktop

import (
	"path/filepath"
	"testing"
)

// an mbox keeps its own name so the imported tree reads like the source, while
// single messages share one folder because a .eml has no folder to name.
func TestFolderNameFor(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{filepath.Join("x", "Archive.mbox"), "Archive"},
		{filepath.Join("x", "Archive.mbx"), "Archive"},
		{filepath.Join("x", "receipt.eml"), importedFolderName},
		{filepath.Join("x", "RECEIPT.EML"), importedFolderName},
		// Thunderbird's own folder files carry no extension.
		{filepath.Join("x", "Sent"), "Sent"},
		// a name that is nothing but an extension leaves no folder name behind.
		{filepath.Join("x", ".mbox"), importedFolderName},
	}
	for _, tc := range tests {
		if got := folderNameFor(tc.path); got != tc.want {
			t.Errorf("folderNameFor(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestSourcesForFilesGroupsSingleMessages(t *testing.T) {
	sources := sourcesForFiles([]string{"a.eml", "b.eml", "Old.mbox"})
	if len(sources) != 3 {
		t.Fatalf("got %d sources, want 3", len(sources))
	}
	if sources[0].Folder != importedFolderName || sources[1].Folder != importedFolderName {
		t.Fatalf("single messages did not share a folder: %+v", sources)
	}
	if sources[2].Folder != "Old" {
		t.Fatalf("archive folder = %q, want Old", sources[2].Folder)
	}
}

// a nested Thunderbird folder arrives as a path inside a .sbd directory; the
// leaf name is what it should keep.
func TestSourcesForFoldersUsesLeafName(t *testing.T) {
	path := filepath.Join("Local Folders", "Archive.sbd", "2019")
	sources := sourcesForFolders([]string{path})
	if len(sources) != 1 || sources[0].Folder != "2019" {
		t.Fatalf("sources = %+v, want one folder named 2019", sources)
	}
}

package desktop

import (
	"testing"

	"github.com/TRC-Loop/Pelton/internal/storage"
	psync "github.com/TRC-Loop/Pelton/internal/sync"
)

func folder(id int64, name string) storage.Folder {
	return storage.Folder{ID: id, Name: name}
}

// the run total is the sum of what each folder's plan asked for, and it grows
// as folders open rather than being guessed at the start.
func TestTallySumsFoldersAsTheyOpen(t *testing.T) {
	var tally syncTally
	tally.begin(3)

	if c := tally.counts(); c.Total != 0 || c.Done != 0 {
		t.Fatalf("a fresh run reports %d of %d, want 0 of 0", c.Done, c.Total)
	}

	tally.enterFolder(0, "INBOX")
	c := tally.record(psync.FolderProgress{Folder: folder(1, "INBOX"), Done: 0, Total: 200})
	if c.Total != 200 || c.Done != 0 {
		t.Errorf("after the first plan: %d of %d, want 0 of 200", c.Done, c.Total)
	}

	c = tally.record(psync.FolderProgress{Folder: folder(1, "INBOX"), Done: 50, Total: 200})
	if c.Done != 50 || c.Total != 200 {
		t.Errorf("mid folder: %d of %d, want 50 of 200", c.Done, c.Total)
	}
	if c.FolderDone != 50 || c.FolderTotal != 200 {
		t.Errorf("folder counts %d of %d, want 50 of 200", c.FolderDone, c.FolderTotal)
	}

	// a second folder adds to the total rather than replacing it, and repeated
	// reports for the same folder do not double count.
	tally.enterFolder(1, "Archive")
	tally.record(psync.FolderProgress{Folder: folder(1, "INBOX"), Done: 200, Total: 200})
	c = tally.record(psync.FolderProgress{Folder: folder(2, "Archive"), Done: 10, Total: 40})
	if c.Total != 240 || c.Done != 210 {
		t.Errorf("across two folders: %d of %d, want 210 of 240", c.Done, c.Total)
	}
	if c.Folder != "Archive" || c.FoldersDone != 1 || c.FoldersTotal != 3 {
		t.Errorf("folder line = %q %d of %d, want Archive 1 of 3", c.Folder, c.FoldersDone, c.FoldersTotal)
	}
}

// a folder with nothing to fetch reports 0 of 0 rather than leaving the last
// folder's numbers standing, which would show a bar that never finishes.
func TestTallyClearsFolderCountsOnEntry(t *testing.T) {
	var tally syncTally
	tally.begin(2)
	tally.enterFolder(0, "INBOX")
	tally.record(psync.FolderProgress{Folder: folder(1, "INBOX"), Done: 30, Total: 30})

	tally.enterFolder(1, "Junk")
	c := tally.counts()
	if c.FolderDone != 0 || c.FolderTotal != 0 {
		t.Errorf("new folder starts at %d of %d, want 0 of 0", c.FolderDone, c.FolderTotal)
	}
	// the run total still holds what the first folder did.
	if c.Done != 30 || c.Total != 30 {
		t.Errorf("run counts %d of %d, want 30 of 30", c.Done, c.Total)
	}
}

// a second run starts from nothing, so yesterday's numbers cannot leak into it.
func TestTallyResetsBetweenRuns(t *testing.T) {
	var tally syncTally
	tally.begin(1)
	tally.enterFolder(0, "INBOX")
	tally.record(psync.FolderProgress{Folder: folder(1, "INBOX"), Done: 100, Total: 100})

	tally.begin(1)
	if c := tally.counts(); c.Done != 0 || c.Total != 0 || c.Folder != "" {
		t.Errorf("new run reports %q %d of %d, want empty 0 of 0", c.Folder, c.Done, c.Total)
	}
}

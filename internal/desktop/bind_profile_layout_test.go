package desktop

import (
	"testing"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// layoutTestFolder creates an account with two folders and returns their ids.
func layoutTestFolder(t *testing.T, a *App) (int64, int64) {
	t.Helper()
	accountID, err := a.store.CreateAccount(a.ctx, &storage.Account{Email: "me@example.com"})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	inbox := storage.Folder{AccountID: accountID, Name: "INBOX", IMAPPath: "INBOX"}
	work := storage.Folder{AccountID: accountID, Name: "Work", IMAPPath: "Work"}
	for _, f := range []*storage.Folder{&inbox, &work} {
		if _, err := a.store.CreateFolder(a.ctx, f); err != nil {
			t.Fatalf("create folder %q: %v", f.Name, err)
		}
	}
	return inbox.ID, work.ID
}

// TestSwitchProfileSwitchesTheLayout is the whole of #325 from the outside: the
// pins and the order follow the profile you are in.
func TestSwitchProfileSwitchesTheLayout(t *testing.T) {
	a := newProfileApp(t)
	inbox, work := layoutTestFolder(t, a)

	if err := a.store.SetFolderPinned(a.ctx, work, true); err != nil {
		t.Fatalf("pin Work: %v", err)
	}
	if err := a.store.SetFolderPositions(a.ctx, []int64{work, inbox}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	created, err := a.CreateProfile(ProfileRequest{Name: "work", StartLayout: startFresh})
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if created.ShareLayout {
		t.Errorf("a fresh layout profile reports shareLayout %v, want false", created.ShareLayout)
	}
	switchTo(t, a, created.ID)

	pinned, err := a.store.ListPinnedFolders(a.ctx)
	if err != nil {
		t.Fatalf("list pinned: %v", err)
	}
	if len(pinned) != 0 {
		t.Errorf("the new profile has %d pinned folders, want none", len(pinned))
	}
	folders, err := a.store.ListFolders(a.ctx, pinnedAccountOf(t, a, inbox))
	if err != nil {
		t.Fatalf("list folders: %v", err)
	}
	if len(folders) != 2 || folders[0].ID != inbox {
		t.Errorf("the new profile lists %v first, want the untouched discovery order", folders[0].Name)
	}
}

// TestCreateProfileCopiesTheLayout covers the default a new profile is made
// with: its own layout, starting from a copy of main's.
func TestCreateProfileCopiesTheLayout(t *testing.T) {
	a := newProfileApp(t)
	inbox, work := layoutTestFolder(t, a)
	if err := a.store.SetFolderPositions(a.ctx, []int64{work, inbox}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	created, err := a.CreateProfile(ProfileRequest{Name: "work", StartLayout: startCopy})
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	switchTo(t, a, created.ID)

	folders, err := a.store.ListFolders(a.ctx, pinnedAccountOf(t, a, inbox))
	if err != nil {
		t.Fatalf("list folders: %v", err)
	}
	if folders[0].ID != work {
		t.Errorf("copied layout lists %q first, want Work", folders[0].Name)
	}

	// a copy is a snapshot: rearranging it leaves main where it was.
	if err := a.store.SetFolderPositions(a.ctx, []int64{inbox, work}); err != nil {
		t.Fatalf("rearrange copy: %v", err)
	}
	main, err := a.store.MainProfile(a.ctx)
	if err != nil {
		t.Fatalf("main profile: %v", err)
	}
	switchTo(t, a, main.ID)
	folders, err = a.store.ListFolders(a.ctx, pinnedAccountOf(t, a, inbox))
	if err != nil {
		t.Fatalf("list folders in main: %v", err)
	}
	if folders[0].ID != work {
		t.Errorf("main lists %q first after the copy was rearranged, want Work", folders[0].Name)
	}
}

// TestSharingTheLayoutKeepsTheProfilesOwn: ticking the box in the editor makes
// the profile follow main, and unticking brings its own arrangement back.
func TestSharingTheLayoutKeepsTheProfilesOwn(t *testing.T) {
	a := newProfileApp(t)
	inbox, work := layoutTestFolder(t, a)
	if err := a.store.SetFolderPositions(a.ctx, []int64{work, inbox}); err != nil {
		t.Fatalf("reorder main: %v", err)
	}

	created, err := a.CreateProfile(ProfileRequest{Name: "work", StartLayout: startFresh})
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	switchTo(t, a, created.ID)
	if err := a.store.SetFolderPositions(a.ctx, []int64{inbox, work}); err != nil {
		t.Fatalf("reorder in profile: %v", err)
	}

	shared, err := a.UpdateProfile(ProfileRequest{ID: created.ID, Name: "work", StartLayout: startShare})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if !shared.ShareLayout {
		t.Fatal("UpdateProfile did not record the share")
	}
	folders, err := a.store.ListFolders(a.ctx, pinnedAccountOf(t, a, inbox))
	if err != nil {
		t.Fatalf("list folders while sharing: %v", err)
	}
	if folders[0].ID != work {
		t.Errorf("while sharing, first folder is %q, want main's Work", folders[0].Name)
	}

	if _, err := a.UpdateProfile(ProfileRequest{ID: created.ID, Name: "work", StartLayout: startFresh}); err != nil {
		t.Fatalf("UpdateProfile unshare: %v", err)
	}
	folders, err = a.store.ListFolders(a.ctx, pinnedAccountOf(t, a, inbox))
	if err != nil {
		t.Fatalf("list folders after unshare: %v", err)
	}
	if folders[0].ID != inbox {
		t.Errorf("after unsharing, first folder is %q, want the profile's own INBOX", folders[0].Name)
	}
}

// switchTo switches profile and stops the session the switch started. These
// tests only care where the store is pointed, and a background sync outliving
// the test leaves wal files behind that t.TempDir then cannot remove.
func switchTo(t *testing.T, a *App, id int64) {
	t.Helper()
	if err := a.SwitchProfile(id); err != nil {
		t.Fatalf("SwitchProfile: %v", err)
	}
	a.endProfileSession()
}

// pinnedAccountOf returns the account a folder belongs to.
func pinnedAccountOf(t *testing.T, a *App, folderID int64) int64 {
	t.Helper()
	folder, err := a.store.GetFolder(a.ctx, folderID)
	if err != nil {
		t.Fatalf("get folder: %v", err)
	}
	return folder.AccountID
}

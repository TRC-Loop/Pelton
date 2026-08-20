package mailimport

import (
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Importing must never touch the source. The user is pointing Pelton at another
// program's live mail store, often one they still use, so reading it is the
// whole contract: no rewrite, no reindex, no deletion, not even a "harmless"
// touch of a modification time. These tests fingerprint the tree before the
// import and require it to come back identical.

// snapshot records every file under root by path, content hash, size and
// modification time.
type snapshot map[string]fileState

type fileState struct {
	hash    [32]byte
	size    int64
	modTime time.Time
	mode    fs.FileMode
}

func snap(t *testing.T, root string) snapshot {
	t.Helper()
	out := make(snapshot)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		state := fileState{size: info.Size(), modTime: info.ModTime(), mode: info.Mode()}
		if !d.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			state.hash = sha256.Sum256(content)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[rel] = state
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

func assertUnchanged(t *testing.T, before, after snapshot) {
	t.Helper()
	for path, was := range before {
		now, ok := after[path]
		if !ok {
			t.Errorf("%s was deleted by the import", path)
			continue
		}
		if now.hash != was.hash {
			t.Errorf("%s was rewritten by the import", path)
		}
		if now.size != was.size {
			t.Errorf("%s changed size: %d, was %d", path, now.size, was.size)
		}
		if !now.modTime.Equal(was.modTime) {
			t.Errorf("%s had its modification time changed", path)
		}
		if now.mode != was.mode {
			t.Errorf("%s had its permissions changed: %v, was %v", path, now.mode, was.mode)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("the import created %s next to the source files", path)
		}
	}
}

// a Thunderbird-shaped profile: prefs.js, a Local Folders store with a nested
// .sbd folder, and the .msf index sidecars Thunderbird keeps beside its mail.
func writeProfile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"prefs.js": `// Mozilla User Preferences
user_pref("mail.accountmanager.accounts", "account1,account2");
user_pref("mail.account.account1.server", "server1");
user_pref("mail.account.account1.identities", "id1");
user_pref("mail.server.server1.type", "imap");
user_pref("mail.server.server1.hostname", "imap.example.com");
user_pref("mail.server.server1.port", 993);
user_pref("mail.identity.id1.useremail", "me@example.com");
user_pref("mail.account.account2.server", "server2");
user_pref("mail.server.server2.type", "none");
user_pref("mail.server.server2.directory-rel", "[ProfD]Mail/Local Folders");
`,
		filepath.Join("Mail", "Local Folders", "Inbox"): "From a@example.com Mon Jan  6 10:00:00 2020\r\n" +
			"Message-ID: <a@example.com>\r\nSubject: one\r\n\r\nbody one\r\n",
		filepath.Join("Mail", "Local Folders", "Inbox.msf"): "index sidecar, must not be touched",
		filepath.Join("Mail", "Local Folders", "Archive.sbd", "2019"): "From b@example.com Tue Jan  7 10:00:00 2020\r\n" +
			"Message-ID: <b@example.com>\r\nSubject: two\r\n\r\nbody two\r\n",
	}
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestImportLeavesTheSourceProfileUntouched(t *testing.T) {
	profileDir := writeProfile(t)
	db := newTestStore(t)
	ctx := context.Background()

	profile, err := ReadProfile(profileDir)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if len(profile.Accounts) != 1 || len(profile.LocalFolders) != 2 {
		t.Fatalf("profile = %+v, want 1 account and 2 folders", profile)
	}

	before := snap(t, profileDir)

	sources := make([]Source, 0, len(profile.LocalFolders))
	for _, f := range profile.LocalFolders {
		sources = append(sources, Source{Path: f.Path, Folder: f.Name})
	}
	result, err := New(db, nil).Import(ctx, sources)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if result.Imported != 2 {
		t.Fatalf("result = %+v, want 2 imported", result)
	}

	assertUnchanged(t, before, snap(t, profileDir))
}

// discovery walks directories looking for profiles, which must also be a pure
// read: no cache file, no marker, nothing left behind.
func TestReadProfileLeavesNothingBehind(t *testing.T) {
	profileDir := writeProfile(t)
	before := snap(t, profileDir)

	for range 3 {
		if _, err := ReadProfile(profileDir); err != nil {
			t.Fatalf("read profile: %v", err)
		}
	}
	assertUnchanged(t, before, snap(t, profileDir))
}

// a re-import skips messages it already has, and must not "tidy" the source
// while deciding that.
func TestReimportLeavesTheSourceUntouched(t *testing.T) {
	db := newTestStore(t)
	ctx := context.Background()
	path := writeFile(t, t.TempDir(), "Archive",
		"From a@example.com Mon Jan  6 10:00:00 2020\r\nMessage-ID: <a@example.com>\r\nSubject: one\r\n\r\nbody\r\n")
	dir := filepath.Dir(path)

	importer := New(db, nil)
	if _, err := importer.Import(ctx, []Source{{Path: path, Folder: "Archive"}}); err != nil {
		t.Fatalf("first import: %v", err)
	}
	before := snap(t, dir)

	result, err := importer.Import(ctx, []Source{{Path: path, Folder: "Archive"}})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if result.Skipped != 1 {
		t.Fatalf("result = %+v, want 1 skipped", result)
	}
	assertUnchanged(t, before, snap(t, dir))
}

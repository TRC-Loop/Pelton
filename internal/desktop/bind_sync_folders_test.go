package desktop

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// The bug this guards: a folder that failed to sync was logged and skipped, and
// the account was then recorded as a clean sync. The mailbox stopped receiving
// mail, no marker appeared, and the refresh button reported success, so there
// was nothing anywhere to say what had gone wrong.
func TestFolderSyncErrorCarriesTheFailureOut(t *testing.T) {
	if err := folderSyncError(nil, 4, nil); err != nil {
		t.Fatalf("folderSyncError with no failures = %v, want nil", err)
	}

	one := folderSyncError([]string{"INBOX"}, 4, errors.New("no mailbox selected"))
	if one == nil {
		t.Fatal("a failed folder produced no error, so the account would be recorded as synced")
	}

	many := folderSyncError([]string{"INBOX", "Sent"}, 4, errors.New("no mailbox selected"))
	if many == nil {
		t.Fatal("two failed folders produced no error")
	}
}

// The reason has to survive being wrapped, since the ui classifies a failure by
// what the error is and shows the server's own words in the detail dialog.
func TestFolderSyncErrorKeepsTheReason(t *testing.T) {
	for _, tt := range []struct {
		name   string
		failed []string
		err    error
		want   string
	}{
		{"one folder", []string{"INBOX"}, pimap.ErrAuthFailed, syncFailAuth},
		{"several folders", []string{"INBOX", "Sent"}, pimap.ErrAuthFailed, syncFailAuth},
		{"wrapped further down", []string{"INBOX"}, fmt.Errorf("select: %w", pimap.ErrAuthFailed), syncFailAuth},
		{"anything else", []string{"INBOX"}, errors.New("no mailbox selected"), syncFailOther},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := folderSyncError(tt.failed, 4, tt.err)
			if got := syncFailureReason(err); got != tt.want {
				t.Fatalf("syncFailureReason(%v) = %q, want %q", err, got, tt.want)
			}
		})
	}
}

// The folder that failed has to be named, otherwise the detail dialog says a
// sync failed without saying where.
func TestFolderSyncErrorNamesTheFolder(t *testing.T) {
	err := folderSyncError([]string{"Archive"}, 4, errors.New("select refused"))
	got := err.Error()
	if !strings.Contains(got, "Archive") || !strings.Contains(got, "select refused") {
		t.Fatalf("error = %q, want it to name the folder and the reason", got)
	}
}

// A \Noselect folder is a container in the hierarchy, not a mailbox. Syncing
// one fails every time, so with failures now reported it would mark a perfectly
// healthy account as broken on every run.
func TestFolderSelectable(t *testing.T) {
	for _, tt := range []struct {
		name  string
		attrs []string
		want  bool
	}{
		{"an ordinary folder", nil, true},
		{"a folder with a role", []string{"\\Sent"}, true},
		{"a container", []string{"\\Noselect"}, false},
		{"a container, other case", []string{"\\NoSelect"}, false},
		{"gone from the server", []string{"\\NonExistent"}, false},
		{"a container that also has children", []string{"\\HasChildren", "\\Noselect"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := storage.Folder{Name: "x", Attributes: tt.attrs}
			if got := folderSelectable(f); got != tt.want {
				t.Fatalf("folderSelectable(%v) = %v, want %v", tt.attrs, got, tt.want)
			}
		})
	}
}

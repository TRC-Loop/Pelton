package desktop

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/TRC-Loop/Pelton/internal/credentials"
	"github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

func newAccountTestApp(t *testing.T) *App {
	t.Helper()
	ctx := context.Background()
	store, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &App{ctx: ctx, store: store, log: slog.New(slog.DiscardHandler)}
}

// needsPassword is tested rather than AccountsNeedingPassword because the
// latter reads the real OS keyring, where account ids collide with whatever the
// developer running the tests actually has installed.
func TestNeedsPassword(t *testing.T) {
	a := newAccountTestApp(t)
	acct := storage.Account{ID: 1, Email: "imported@example.test", IMAPHost: "imap.example.test", IMAPPort: 993}

	for _, tt := range []struct {
		name      string
		secretErr error
		want      bool
	}{
		// an account imported from another mail client: the import creates the
		// row but cannot take the password with it.
		{"nothing stored", credentials.ErrNotFound, true},
		{"a secret exists", nil, false},
		// a locked or broken keyring is not the same as no password, and
		// prompting would ask for one the user already gave.
		{"keyring unreadable", errors.New("keyring: access denied"), false},
		{"wrapped not found", fmt.Errorf("load: %w", credentials.ErrNotFound), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.needsPassword(acct, tt.secretErr); got != tt.want {
				t.Errorf("needsPassword(%v) = %v, want %v", tt.secretErr, got, tt.want)
			}
		})
	}
}

// The legacy cli account keeps its password in the environment, so it is not
// missing one even with an empty keyring.
func TestEnvBackedAccountDoesNotNeedAPassword(t *testing.T) {
	a := newAccountTestApp(t)
	acct := storage.Account{ID: 1, Email: "cli@example.test"}
	t.Setenv("IMAP_USER", "cli@example.test")
	t.Setenv("IMAP_PASSWORD", "from-the-environment")

	if a.needsPassword(acct, credentials.ErrNotFound) {
		t.Error("an env-backed account was reported as needing a password")
	}
}

// Local Folders holds imported mail and has no server, so an empty keyring is
// its normal state rather than a missing password.
func TestLocalAccountDoesNotNeedAPassword(t *testing.T) {
	a := newAccountTestApp(t)
	acct := storage.Account{ID: 1, Email: storage.LocalAccountEmail, Local: true}

	if a.needsPassword(acct, credentials.ErrNotFound) {
		t.Error("the Local Folders account was reported as needing a password")
	}
}

func TestEmptyPasswordIsRefused(t *testing.T) {
	a := newAccountTestApp(t)
	if err := a.SetAccountPassword(1, ""); !errors.Is(err, errEmptyPassword) {
		t.Errorf("SetAccountPassword(empty) = %v, want errEmptyPassword", err)
	}
}

// imapFromEnv is what the env check leans on; pin its contract so a change
// there cannot silently start prompting for every account.
func TestImapFromEnvNeedsAMatchingUser(t *testing.T) {
	a := newAccountTestApp(t)
	t.Setenv("IMAP_USER", "someone@example.test")
	t.Setenv("IMAP_PASSWORD", "secret")

	if _, err := a.imapFromEnv(imap.Config{Username: "other@example.test"}); err == nil {
		t.Error("a mismatched IMAP_USER was accepted")
	}
	if _, err := a.imapFromEnv(imap.Config{Username: "someone@example.test"}); err != nil {
		t.Errorf("a matching IMAP_USER was rejected: %v", err)
	}
}

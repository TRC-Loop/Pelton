package desktop

import (
	"errors"
	"fmt"
	"net"
	"testing"
)

// The bug this guards: a missing password ended the idle loop for good. The
// user typed the password, the mailbox still received nothing, and only the
// next launch brought it back. Every error has to leave a wait behind, because
// a password can arrive at any moment.
func TestIdleRetryWaitAlwaysWaits(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want string
	}{
		{"no password", errNoCredentials, "long"},
		{"no password, wrapped", fmt.Errorf("idle: %w", errNoCredentials), "long"},
		{"connection dropped", &net.OpError{Op: "read", Err: errors.New("reset by peer")}, "short"},
		{"anything else", errors.New("imap: unexpected response"), "short"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := idleRetryWait(tt.err)
			if got <= 0 {
				t.Fatalf("idleRetryWait(%v) = %v, want a wait rather than giving up", tt.err, got)
			}
			// a missing password waits longer, since nothing changes until the
			// user types one and each try reads the keyring.
			if tt.want == "long" && got != idleNoCredentialsRetry {
				t.Fatalf("idleRetryWait(%v) = %v, want %v", tt.err, got, idleNoCredentialsRetry)
			}
			if tt.want == "short" && got != idleRetry {
				t.Fatalf("idleRetryWait(%v) = %v, want %v", tt.err, got, idleRetry)
			}
		})
	}
}

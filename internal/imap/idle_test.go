package imap

import (
	"errors"
	"testing"
)

type fakeIdleCmd struct{ err error }

func (f fakeIdleCmd) Close() error { return f.err }

func TestStopIdle(t *testing.T) {
	closeErr := errors.New("close boom")
	waitErr := errors.New("wait boom")

	tests := []struct {
		name    string
		close   error
		wait    error
		wantErr bool
	}{
		{"clean", nil, nil, false},
		{"close fails", closeErr, nil, true},
		{"wait fails", nil, waitErr, true},
		{"both fail reports close", closeErr, waitErr, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			done := make(chan error, 1)
			done <- tt.wait
			err := c.stopIdle(fakeIdleCmd{err: tt.close}, done)
			if tt.wantErr != (err != nil) {
				t.Fatalf("stopIdle err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDrainUpdates(t *testing.T) {
	c := &Client{updates: make(chan MailboxUpdate, updateBuffer)}
	for range 5 {
		c.updates <- MailboxUpdate{}
	}
	c.drainUpdates()
	if len(c.updates) != 0 {
		t.Fatalf("drainUpdates left %d updates", len(c.updates))
	}
	// safe to call on an empty channel.
	c.drainUpdates()
}

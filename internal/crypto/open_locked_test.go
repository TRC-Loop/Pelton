package crypto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// TestOpenLockedKeyWithTheRightPassphrase is the other half of
// TestOpenLockedKeyAsksForPassphrase, which only ever checked that the wrong
// passphrase fails. Reading a message with a locked key is the ordinary case
// for anyone who imported a key from gpg.
func TestOpenLockedKeyWithTheRightPassphrase(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})
	part, err := p.Wrap([]byte(sampleEntity), ModeEncrypt, Options{
		SenderEmail: "alice@example.test",
		Recipients:  []string{"alice@example.test"},
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	raw := message("alice@example.test", part)

	locked := newTestEntity(t, "alice@example.test")
	*locked = *alice
	passphrase := []byte("correct horse")
	lockEntity(t, locked, passphrase)
	reader := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{locked})

	opened, err := reader.Open(raw, passphrase)
	if err != nil {
		t.Fatalf("Open() with the right passphrase: %v", err)
	}
	if !opened.Encrypted {
		t.Error("the message is not reported as encrypted")
	}
	if !bytes.Contains(opened.Body, []byte("the eagle lands at dawn")) {
		t.Errorf("decrypted body = %q", opened.Body)
	}
}

// TestOpenLockedKeyAmongUnlockedOnes: a ring can hold several keys, and one
// belonging to a passphrase the user did not type must not stop a message
// addressed to an unlocked one from opening.
func TestOpenLockedKeyAmongUnlockedOnes(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	bob := newTestEntity(t, "bob@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice, bob}, []*openpgp.Entity{alice, bob})
	part, err := p.Wrap([]byte(sampleEntity), ModeEncrypt, Options{
		SenderEmail: "alice@example.test",
		Recipients:  []string{"alice@example.test"},
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	raw := message("alice@example.test", part)

	lockedBob := newTestEntity(t, "bob@example.test")
	*lockedBob = *bob
	lockEntity(t, lockedBob, []byte("bob's passphrase"))
	reader := newOpenTestPGP(t, []*openpgp.Entity{alice, bob}, []*openpgp.Entity{alice, lockedBob})

	if _, err := reader.Open(raw, nil); err != nil {
		t.Fatalf("Open() with an unlocked key present: %v", err)
	}
}

func TestUnlockRing(t *testing.T) {
	passphrase := []byte("correct horse")
	locked := func(t *testing.T) *openpgp.Entity {
		ent := newTestEntity(t, "alice@example.test")
		lockEntity(t, ent, passphrase)
		return ent
	}

	tests := []struct {
		name       string
		ring       func(t *testing.T) openpgp.EntityList
		passphrase []byte
		wantErr    error
	}{
		{
			name:    "no keys at all",
			ring:    func(*testing.T) openpgp.EntityList { return nil },
			wantErr: ErrNoDecryptionKey,
		},
		{
			name: "an unlocked key needs nothing",
			ring: func(t *testing.T) openpgp.EntityList {
				return openpgp.EntityList{newTestEntity(t, "alice@example.test")}
			},
			wantErr: nil,
		},
		{
			name:    "locked with no passphrase asks for one",
			ring:    func(t *testing.T) openpgp.EntityList { return openpgp.EntityList{locked(t)} },
			wantErr: ErrPassphraseRequired,
		},
		{
			name:       "locked with the wrong passphrase says so",
			ring:       func(t *testing.T) openpgp.EntityList { return openpgp.EntityList{locked(t)} },
			passphrase: []byte("nope"),
			wantErr:    ErrWrongPassphrase,
		},
		{
			name:       "locked with the right passphrase opens",
			ring:       func(t *testing.T) openpgp.EntityList { return openpgp.EntityList{locked(t)} },
			passphrase: passphrase,
			wantErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unlockRing(tt.ring(t), tt.passphrase); !errors.Is(err, tt.wantErr) {
				t.Errorf("unlockRing() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

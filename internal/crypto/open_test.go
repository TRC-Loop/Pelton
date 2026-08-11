package crypto

import (
	"errors"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// newOpenTestPGP builds an engine over a keyring holding pub and priv.
func newOpenTestPGP(t *testing.T, pub, priv []*openpgp.Entity) *PGP {
	t.Helper()
	dir := t.TempDir()
	writeKeyrings(t, dir, pub, priv)
	return NewPGP(NewPGPKeyStore(dir))
}

// message wraps a crypto Part into the full RFC 822 message Open expects.
func message(from string, part *Part) []byte {
	return []byte("From: " + from + "\r\n" +
		"Subject: test\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: " + part.ContentType + "\r\n" +
		"\r\n" + string(part.Body))
}

// Anything Pelton can produce, Pelton must be able to read back.
func TestOpenEncryptedRoundTrip(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})

	part, err := p.Wrap([]byte(sampleEntity), ModeEncrypt, Options{
		SenderEmail: "alice@example.test",
		Recipients:  []string{"alice@example.test"},
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	opened, err := p.Open(message("alice@example.test", part), nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !opened.Encrypted {
		t.Error("Encrypted = false, want true")
	}
	if !strings.Contains(string(opened.Body), plaintextMarker) {
		t.Errorf("decrypted body does not contain the plaintext: %q", opened.Body)
	}
}

func TestOpenSignedRoundTrip(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})

	part, err := p.Wrap([]byte(sampleEntity), ModeSign, Options{SenderEmail: "alice@example.test"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	opened, err := p.Open(message("alice@example.test", part), nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened.Encrypted {
		t.Error("Encrypted = true for a signed-only message")
	}
	if opened.Signature.Status != SigValid {
		t.Fatalf("Status = %q (%s), want %q", opened.Signature.Status, opened.Signature.Detail, SigValid)
	}
	if opened.Signature.SignerEmail != "alice@example.test" {
		t.Errorf("SignerEmail = %q", opened.Signature.SignerEmail)
	}
	if opened.Signature.Fingerprint == "" {
		t.Error("Fingerprint is empty, the ui needs it to identify the key")
	}
	if !strings.Contains(string(opened.Body), plaintextMarker) {
		t.Errorf("signed body missing: %q", opened.Body)
	}
}

// A signature riding inside an encrypted message must still be reported.
func TestOpenSignedAndEncrypted(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})

	part, err := p.Wrap([]byte(sampleEntity), ModeSignEncrypt, Options{
		SenderEmail: "alice@example.test",
		Recipients:  []string{"alice@example.test"},
	})
	if err != nil {
		t.Fatalf("sign+encrypt: %v", err)
	}

	opened, err := p.Open(message("alice@example.test", part), nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !opened.Encrypted {
		t.Error("Encrypted = false, want true")
	}
	if opened.Signature.Status != SigValid {
		t.Errorf("Status = %q (%s), want %q", opened.Signature.Status, opened.Signature.Detail, SigValid)
	}
}

// Mail encrypted to somebody else must fail plainly rather than half-open.
func TestOpenWithNoDecryptionKey(t *testing.T) {
	bob := newTestEntity(t, "bob@example.test")
	sender := newOpenTestPGP(t, []*openpgp.Entity{bob}, []*openpgp.Entity{bob})
	part, err := sender.Wrap([]byte(sampleEntity), ModeEncrypt, Options{
		SenderEmail: "bob@example.test",
		Recipients:  []string{"bob@example.test"},
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// a store holding no private keys at all.
	reader := newOpenTestPGP(t, nil, nil)
	if _, err := reader.Open(message("bob@example.test", part), nil); !errors.Is(err, ErrNoDecryptionKey) {
		t.Errorf("Open() = %v, want ErrNoDecryptionKey", err)
	}
}

// A signature from a key we do not hold is unverified, which is neither good
// nor bad and must not be reported as either.
func TestOpenSignedByUnknownKey(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	signer := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})
	part, err := signer.Wrap([]byte(sampleEntity), ModeSign, Options{SenderEmail: "alice@example.test"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// a reader who has never imported alice's key.
	stranger := newTestEntity(t, "carol@example.test")
	reader := newOpenTestPGP(t, []*openpgp.Entity{stranger}, []*openpgp.Entity{stranger})

	opened, err := reader.Open(message("alice@example.test", part), nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened.Signature.Status != SigUntrusted {
		t.Fatalf("Status = %q, want %q", opened.Signature.Status, SigUntrusted)
	}
	if !strings.Contains(opened.Signature.Detail, "keyring") {
		t.Errorf("Detail = %q, want it to say the key is missing", opened.Signature.Detail)
	}
}

// The point of a signature: altering the signed part has to be caught.
func TestOpenDetectsTamperedSignedContent(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})

	part, err := p.Wrap([]byte(sampleEntity), ModeSign, Options{SenderEmail: "alice@example.test"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	raw := message("alice@example.test", part)
	tampered := strings.Replace(string(raw), "eagle", "seagull", 1)
	if tampered == string(raw) {
		t.Fatal("the test did not alter anything")
	}

	opened, err := p.Open([]byte(tampered), nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened.Signature.Status != SigInvalid {
		t.Errorf("Status = %q (%s), want %q", opened.Signature.Status, opened.Signature.Detail, SigInvalid)
	}
}

// A locked key must ask for the passphrase rather than failing obscurely, so
// the reading pane knows to prompt.
func TestOpenLockedKeyAsksForPassphrase(t *testing.T) {
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
	lockEntity(t, locked, []byte("correct horse"))
	reader := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{locked})

	if _, err := reader.Open(raw, nil); !errors.Is(err, ErrPassphraseRequired) {
		t.Errorf("Open() with no passphrase = %v, want ErrPassphraseRequired", err)
	}
	if _, err := reader.Open(raw, []byte("wrong")); err == nil {
		t.Error("Open() with the wrong passphrase succeeded")
	}
}

func TestOpenUnprotectedMessage(t *testing.T) {
	p := newOpenTestPGP(t, nil, nil)
	raw := []byte("From: bob@example.test\r\nContent-Type: text/plain\r\n\r\nnothing to see\r\n")
	if _, err := p.Open(raw, nil); !errors.Is(err, ErrNotProtected) {
		t.Errorf("Open() = %v, want ErrNotProtected", err)
	}
}

// Malformed input must be refused, never rendered as though it were fine.
func TestOpenRejectsMalformedInput(t *testing.T) {
	alice := newTestEntity(t, "alice@example.test")
	p := newOpenTestPGP(t, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})

	for _, tt := range []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"encrypted with no boundary", "Content-Type: multipart/encrypted; protocol=\"application/pgp-encrypted\"\r\n\r\nbody"},
		{"signed with no parts", "Content-Type: multipart/signed; protocol=\"application/pgp-signature\"; boundary=\"x\"\r\n\r\nnothing"},
		{"armor header but no block", "Content-Type: text/plain\r\n\r\n-----BEGIN PGP MESSAGE-----\r\nnot armor\r\n"},
		{"clearsign header but no block", "Content-Type: text/plain\r\n\r\n-----BEGIN PGP SIGNED MESSAGE-----\r\ntruncated\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opened, err := p.Open([]byte(tt.raw), nil)
			if err == nil && opened != nil && opened.Signature.Status == SigValid {
				t.Error("malformed input reported a valid signature")
			}
		})
	}
}

// lockEntity encrypts an entity's private key material so it needs a
// passphrase, the state a key imported from gpg normally arrives in.
func lockEntity(t *testing.T, ent *openpgp.Entity, passphrase []byte) {
	t.Helper()
	if ent.PrivateKey != nil {
		if err := ent.PrivateKey.Encrypt(passphrase); err != nil {
			t.Fatalf("lock primary key: %v", err)
		}
	}
	for _, sub := range ent.Subkeys {
		if sub.PrivateKey == nil {
			continue
		}
		if err := sub.PrivateKey.Encrypt(passphrase); err != nil {
			t.Fatalf("lock subkey: %v", err)
		}
	}
}

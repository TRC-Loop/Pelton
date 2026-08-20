package crypto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// sealTestStore builds a key store holding one locked key, which is the shape a
// real user's own key has.
func sealTestStore(t *testing.T) (*PGPKeyStore, string, []byte) {
	t.Helper()
	dir := t.TempDir()
	const email = "alice@example.com"
	passphrase := []byte("correct horse battery staple")

	alice := newTestEntity(t, email)
	if err := alice.EncryptPrivateKeys(passphrase, nil); err != nil {
		t.Fatalf("lock private keys: %v", err)
	}
	writeKeyrings(t, dir, []*openpgp.Entity{alice}, []*openpgp.Entity{alice})
	return NewPGPKeyStore(dir), email, passphrase
}

// TestSealRoundTrip: a draft sealed to the user's own key comes back byte for
// byte, and never touches disk as plaintext in between.
func TestSealRoundTrip(t *testing.T) {
	store, email, passphrase := sealTestStore(t)
	engine := NewPGP(store)
	plaintext := []byte(`{"subject":"quarterly numbers","text":"not for the database"}`)

	sealed, err := engine.Seal(email, plaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if bytes.Contains(sealed, []byte("quarterly")) {
		t.Fatal("the sealed form still contains the plaintext")
	}
	if !bytes.HasPrefix(sealed, []byte("-----BEGIN PGP MESSAGE-----")) {
		t.Errorf("sealed data is not an armored pgp message: %.40s", sealed)
	}

	opened, err := engine.Unseal(sealed, passphrase)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Errorf("round trip changed the data: %s", opened)
	}
}

func TestSealWithoutAKey(t *testing.T) {
	store, _, _ := sealTestStore(t)
	if _, err := NewPGP(store).Seal("nobody@example.com", []byte("x")); err == nil {
		t.Error("Seal to an address with no key returned no error")
	}
	if _, err := NewPGP(store).Seal("", []byte("x")); err == nil {
		t.Error("Seal with no address returned no error")
	}
}

// TestUnsealLockedKey: the passphrase is what the ui prompts for, so the error
// has to be the one that means "ask", not a generic failure.
func TestUnsealLockedKey(t *testing.T) {
	store, email, passphrase := sealTestStore(t)
	engine := NewPGP(store)
	sealed, err := engine.Seal(email, []byte("secret"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if len(passphrase) == 0 {
		t.Skip("test key is not passphrase protected")
	}
	if _, err := engine.Unseal(sealed, nil); !errors.Is(err, ErrPassphraseRequired) {
		t.Errorf("Unseal with no passphrase = %v, want ErrPassphraseRequired", err)
	}
}

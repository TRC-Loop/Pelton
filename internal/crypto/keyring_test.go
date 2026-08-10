package crypto

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// exportEntity renders one entity the way gpg would, so imports are tested
// against something shaped like a real export rather than our own writer.
func exportEntity(t *testing.T, ent *openpgp.Entity, private bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	block := publicKeyBlock
	if private {
		block = privateKeyBlock
	}
	w, err := armor.Encode(&buf, block, nil)
	if err != nil {
		t.Fatalf("armor: %v", err)
	}
	if private {
		err = ent.SerializePrivateWithoutSigning(w, nil)
	} else {
		err = ent.Serialize(w)
	}
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close armor: %v", err)
	}
	return buf.Bytes()
}

func newStore(t *testing.T) *PGPKeyStore {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "pgp")
	if err := EnsureKeyDir(dir); err != nil {
		t.Fatalf("EnsureKeyDir: %v", err)
	}
	return NewPGPKeyStore(dir)
}

func TestImportPrivateKeyLandsInBothRings(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")

	infos, err := s.Import(exportEntity(t, ent, true))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("imported %d keys, want 1", len(infos))
	}
	if !infos[0].HasPrivate {
		t.Error("imported private key does not report HasPrivate")
	}

	// signing needs it in the private ring...
	if _, err := s.SenderKey("ada@example.test"); err != nil {
		t.Errorf("SenderKey after private import: %v", err)
	}
	// ...and encrypting to yourself needs the public half too.
	if _, err := s.RecipientKey("ada@example.test"); err != nil {
		t.Errorf("RecipientKey after private import: %v", err)
	}
}

func TestImportPublicKeyStaysOutOfThePrivateRing(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "grace@example.test")

	if _, err := s.Import(exportEntity(t, ent, false)); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if _, err := s.RecipientKey("grace@example.test"); err != nil {
		t.Errorf("RecipientKey: %v", err)
	}
	if _, err := s.SenderKey("grace@example.test"); !errors.Is(err, ErrSenderKeyNotFound) {
		t.Errorf("SenderKey = %v, want ErrSenderKeyNotFound", err)
	}
}

// Re-importing a key the user already has must refresh it rather than stack a
// second copy in the ring.
func TestImportReplacesRatherThanDuplicates(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	armored := exportEntity(t, ent, true)

	for range 3 {
		if _, err := s.Import(armored); err != nil {
			t.Fatalf("Import: %v", err)
		}
	}
	keys, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("listed %d keys after three imports, want 1", len(keys))
	}
}

// A private key's public half is in both rings; the list must show it once.
func TestListShowsAKeypairOnce(t *testing.T) {
	s := newStore(t)
	if _, err := s.Import(exportEntity(t, newTestEntity(t, "ada@example.test"), true)); err != nil {
		t.Fatalf("Import private: %v", err)
	}
	if _, err := s.Import(exportEntity(t, newTestEntity(t, "grace@example.test"), false)); err != nil {
		t.Fatalf("Import public: %v", err)
	}

	keys, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("listed %d keys, want 2", len(keys))
	}
	// private keys sort first, so the pair is the one that can sign.
	if !keys[0].HasPrivate || keys[0].Email != "ada@example.test" {
		t.Errorf("first key = %+v, want ada's private key", keys[0])
	}
	if keys[1].HasPrivate {
		t.Errorf("grace's public key reports HasPrivate")
	}
}

func TestDeleteRemovesFromBothRings(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	if _, err := s.Import(exportEntity(t, ent, true)); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if err := s.Delete(Fingerprint(ent)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.SenderKey("ada@example.test"); !errors.Is(err, ErrSenderKeyNotFound) {
		t.Errorf("private key survived delete: %v", err)
	}
	if _, err := s.RecipientKey("ada@example.test"); !errors.Is(err, ErrRecipientKeyNotFound) {
		t.Errorf("public key survived delete: %v", err)
	}
}

// Deleting one key must not take its neighbours with it, which is the failure
// mode of rewriting a whole ring.
func TestDeleteKeepsTheOtherKeys(t *testing.T) {
	s := newStore(t)
	for _, email := range []string{"ada@example.test", "grace@example.test", "alan@example.test"} {
		if _, err := s.Import(exportEntity(t, newTestEntity(t, email), true)); err != nil {
			t.Fatalf("Import %s: %v", email, err)
		}
	}
	keys, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var target string
	for _, k := range keys {
		if k.Email == "grace@example.test" {
			target = k.Fingerprint
		}
	}
	if target == "" {
		t.Fatal("grace's key was never imported")
	}
	if err := s.Delete(target); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	left, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(left) != 2 {
		t.Fatalf("%d keys left, want 2", len(left))
	}
	for _, k := range left {
		if k.Email == "grace@example.test" {
			t.Error("deleted key is still listed")
		}
		if _, err := s.SenderKey(k.Email); err != nil {
			t.Errorf("surviving key %s is no longer usable: %v", k.Email, err)
		}
	}
}

func TestDeleteUnknownFingerprintReports(t *testing.T) {
	s := newStore(t)
	if err := s.Delete("DEADBEEF"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Delete(unknown) = %v, want ErrKeyNotFound", err)
	}
}

// gpg prints fingerprints in spaced groups; pasting one must still match.
func TestFingerprintLookupIgnoresSpacingAndCase(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	if _, err := s.Import(exportEntity(t, ent, true)); err != nil {
		t.Fatalf("Import: %v", err)
	}
	fp := Fingerprint(ent)
	spaced := strings.ToLower(fp[:4] + " " + fp[4:8] + " " + fp[8:])

	if _, err := s.SenderKeyByFingerprint(spaced); err != nil {
		t.Errorf("SenderKeyByFingerprint(%q): %v", spaced, err)
	}
}

func TestExportRoundTripsThroughImport(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	if _, err := s.Import(exportEntity(t, ent, true)); err != nil {
		t.Fatalf("Import: %v", err)
	}
	fp := Fingerprint(ent)

	for _, private := range []bool{false, true} {
		blob, err := s.Export(fp, private)
		if err != nil {
			t.Fatalf("Export(private=%v): %v", private, err)
		}
		other := NewPGPKeyStore(filepath.Join(t.TempDir(), "pgp"))
		infos, err := other.Import(blob)
		if err != nil {
			t.Fatalf("re-import (private=%v): %v", private, err)
		}
		if len(infos) != 1 || infos[0].Fingerprint != fp {
			t.Fatalf("re-imported %+v, want fingerprint %s", infos, fp)
		}
		if infos[0].HasPrivate != private {
			t.Errorf("re-imported HasPrivate = %v, want %v", infos[0].HasPrivate, private)
		}
	}
}

// A public export must not carry private key material, which is the whole
// point of the distinction.
func TestPublicExportCarriesNoPrivateKey(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	if _, err := s.Import(exportEntity(t, ent, true)); err != nil {
		t.Fatalf("Import: %v", err)
	}

	blob, err := s.Export(Fingerprint(ent), false)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if strings.Contains(string(blob), "PRIVATE KEY BLOCK") {
		t.Fatal("public export is armored as a private key block")
	}
	list, err := readKeys(blob)
	if err != nil {
		t.Fatalf("readKeys: %v", err)
	}
	for _, e := range list {
		if e.PrivateKey != nil {
			t.Error("public export contains private key material")
		}
		for _, sk := range e.Subkeys {
			if sk.PrivateKey != nil {
				t.Error("public export contains a private subkey")
			}
		}
	}
}

// A locked key must survive a round trip through the ring. Serializing it the
// signing way would need the passphrase and fail here.
func TestLockedPrivateKeySurvivesImportAndRewrite(t *testing.T) {
	s := newStore(t)
	ent := newTestEntity(t, "ada@example.test")
	passphrase := []byte("correct horse battery staple")
	if err := ent.PrivateKey.Encrypt(passphrase); err != nil {
		t.Fatalf("lock primary key: %v", err)
	}
	for _, sk := range ent.Subkeys {
		if err := sk.PrivateKey.Encrypt(passphrase); err != nil {
			t.Fatalf("lock subkey: %v", err)
		}
	}

	if _, err := s.Import(exportEntity(t, ent, true)); err != nil {
		t.Fatalf("Import locked key: %v", err)
	}
	keys, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 1 || !keys[0].Locked {
		t.Fatalf("listed %+v, want one locked key", keys)
	}

	// importing a second key rewrites the ring, which is where a naive
	// serializer would drop the locked one.
	if _, err := s.Import(exportEntity(t, newTestEntity(t, "grace@example.test"), true)); err != nil {
		t.Fatalf("Import second key: %v", err)
	}
	back, err := s.SenderKey("ada@example.test")
	if err != nil {
		t.Fatalf("SenderKey after rewrite: %v", err)
	}
	if !entityLocked(back) {
		t.Error("key came back unlocked")
	}
	if err := unlock(back, passphrase); err != nil {
		t.Errorf("locked key no longer unlocks with its passphrase: %v", err)
	}
}

func TestImportRejectsNonKeyData(t *testing.T) {
	s := newStore(t)
	if _, err := s.Import([]byte("this is not a key, it is a shopping list")); !errors.Is(err, ErrNoKeysInFile) {
		t.Errorf("Import(garbage) = %v, want ErrNoKeysInFile", err)
	}
}

// The key directory and the rings hold the user's private keys; anything
// group- or world-readable is a real exposure.
func TestKeyMaterialIsOwnerOnly(t *testing.T) {
	s := newStore(t)
	if _, err := s.Import(exportEntity(t, newTestEntity(t, "ada@example.test"), true)); err != nil {
		t.Fatalf("Import: %v", err)
	}

	dir, err := os.Stat(s.dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if perm := dir.Mode().Perm(); perm != keyDirPerm {
		t.Errorf("key directory mode = %04o, want %04o", perm, keyDirPerm)
	}
	for _, file := range []string{pubringFile, secringFile} {
		st, err := os.Stat(filepath.Join(s.dir, file))
		if err != nil {
			t.Fatalf("stat %s: %v", file, err)
		}
		if perm := st.Mode().Perm(); perm != keyFilePerm {
			t.Errorf("%s mode = %04o, want %04o", file, perm, keyFilePerm)
		}
	}
}

// A rewrite must not leave the temporary file behind, which would be a second
// copy of the private ring with a name nobody cleans up.
func TestWriteLeavesNoTempFile(t *testing.T) {
	s := newStore(t)
	if _, err := s.Import(exportEntity(t, newTestEntity(t, "ada@example.test"), true)); err != nil {
		t.Fatalf("Import: %v", err)
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temporary keyring %q left behind", e.Name())
		}
	}
}

func TestEnsureKeyDirTightensLoosePermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pgp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := EnsureKeyDir(dir); err != nil {
		t.Fatalf("EnsureKeyDir: %v", err)
	}
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := st.Mode().Perm(); perm != keyDirPerm {
		t.Errorf("mode = %04o, want %04o", perm, keyDirPerm)
	}
}

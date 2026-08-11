package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
	pgperrors "github.com/ProtonMail/go-crypto/openpgp/errors"
)

// Opening a received message is the mirror of Wrap, and the safety rule runs the
// other way: a message that cannot be decrypted or whose signature does not
// check out must never be presented as though it were fine. Every failure is
// reported, and the caller shows the protected form rather than pretending.

var (
	// ErrNotProtected means the message carries no OpenPGP content to open.
	ErrNotProtected = errors.New("crypto: message is not pgp protected")
	// ErrNoDecryptionKey means none of the stored private keys can open the
	// message: it was encrypted to somebody else, or to a key not imported yet.
	ErrNoDecryptionKey = errors.New("crypto: no private key can decrypt this message")
	// ErrMalformedArmor means the pgp block could not be read at all.
	ErrMalformedArmor = errors.New("crypto: the pgp data is malformed")
	// ErrWrongPassphrase means a private key was found but the passphrase given
	// does not unlock it.
	ErrWrongPassphrase = errors.New("crypto: the passphrase does not unlock the key")
)

// Opened is a received message after decryption and signature checking.
type Opened struct {
	// Body is the decrypted MIME entity for encrypted mail, the signed content
	// for signed mail, or the plain text of a clearsigned block. It is never
	// written to disk by this package.
	Body []byte
	// Encrypted records that Body had to be decrypted to be read, which decides
	// whether the caller may cache it.
	Encrypted bool
	// Signature is the verdict on any signature found. A message may be
	// encrypted without being signed, in which case Status is SigNone.
	Signature Signature
}

// IsProtected reports whether a raw message carries OpenPGP content worth
// keeping the source for. It is a cheap structural check run over every synced
// message, so it looks at the content type and the armor headers rather than
// attempting to parse anything.
func IsProtected(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	headers, body, found := splitHeaders(raw)
	if !found {
		return false
	}
	mediaType, params := parseContentType(headers)
	switch {
	case mediaType == "multipart/encrypted" &&
		strings.EqualFold(params["protocol"], "application/pgp-encrypted"):
		return true
	case mediaType == "multipart/signed" &&
		strings.EqualFold(params["protocol"], "application/pgp-signature"):
		return true
	}
	return bytes.Contains(body, []byte("-----BEGIN PGP MESSAGE-----")) ||
		bytes.Contains(body, []byte("-----BEGIN PGP SIGNED MESSAGE-----"))
}

// Open decrypts and verifies a received message.
//
// raw is the complete RFC 822 message. passphrase unlocks a private key when
// one is needed and may be nil for mail that is only signed. The four shapes
// real mail uses are all handled: PGP/MIME encrypted and signed (RFC 3156), and
// the inline armored and clearsigned forms that predate it and are still
// common.
func (p *PGP) Open(raw []byte, passphrase []byte) (*Opened, error) {
	headers, body, found := splitHeaders(raw)
	if !found {
		return nil, ErrNotProtected
	}
	mediaType, params := parseContentType(headers)

	switch {
	case mediaType == "multipart/encrypted" &&
		strings.EqualFold(params["protocol"], "application/pgp-encrypted"):
		return p.openMIMEEncrypted(body, params["boundary"], passphrase)
	case mediaType == "multipart/signed" &&
		strings.EqualFold(params["protocol"], "application/pgp-signature"):
		return p.openMIMESigned(body, params["boundary"])
	}
	return p.openInline(headers, body, passphrase)
}

// openMIMEEncrypted reads an RFC 3156 multipart/encrypted message. The first
// part is the version marker, which carries nothing worth checking beyond its
// presence; the second holds the armored ciphertext.
func (p *PGP) openMIMEEncrypted(body []byte, boundary string, passphrase []byte) (*Opened, error) {
	if boundary == "" {
		return nil, ErrMalformedArmor
	}
	sections := splitOnBoundary(body, []byte("--"+boundary))
	if len(sections) < 2 {
		return nil, ErrMalformedArmor
	}
	_, ciphertext, ok := splitHeaders(sections[1])
	if !ok {
		return nil, ErrMalformedArmor
	}
	return p.decrypt(ciphertext, passphrase)
}

// openMIMESigned reads an RFC 3156 multipart/signed message. The signed part is
// taken verbatim: the signature covers the bytes as transmitted, so parsing and
// rebuilding the part would invalidate a signature that is perfectly good.
func (p *PGP) openMIMESigned(body []byte, boundary string) (*Opened, error) {
	if boundary == "" {
		return nil, ErrMalformedArmor
	}
	sections := splitOnBoundary(body, []byte("--"+boundary))
	if len(sections) < 2 {
		return nil, ErrMalformedArmor
	}
	signed := sections[0]
	_, sigArmor, ok := splitHeaders(sections[1])
	if !ok {
		return nil, ErrMalformedArmor
	}
	return &Opened{Body: signed, Signature: p.checkDetached(signed, sigArmor)}, nil
}

// openInline handles the pre-MIME forms: an armored message or a clearsigned
// block sitting in an ordinary text body.
func (p *PGP) openInline(headers, body []byte, passphrase []byte) (*Opened, error) {
	text, err := decodeBody(headers, body)
	if err != nil {
		text = body
	}
	switch {
	case bytes.Contains(text, []byte("-----BEGIN PGP MESSAGE-----")):
		return p.decrypt(text, passphrase)
	case bytes.Contains(text, []byte("-----BEGIN PGP SIGNED MESSAGE-----")):
		return p.openClearsigned(text)
	}
	return nil, ErrNotProtected
}

// openClearsigned verifies an inline clearsigned block and returns the text it
// covers, with the armor removed.
func (p *PGP) openClearsigned(text []byte) (*Opened, error) {
	block, _ := clearsign.Decode(text)
	if block == nil {
		return nil, ErrMalformedArmor
	}
	keyring, err := p.keys.PublicKeys()
	if err != nil {
		return nil, err
	}
	opened := &Opened{Body: block.Bytes}
	if len(keyring) == 0 {
		opened.Signature = unknownSigner()
		return opened, nil
	}
	signer, err := openpgp.CheckDetachedSignature(
		keyring, bytes.NewReader(block.Bytes), block.ArmoredSignature.Body, cryptoConfig())
	opened.Signature = signatureVerdict(signer, err)
	return opened, nil
}

// decrypt opens an armored OpenPGP message with whichever stored private key it
// was encrypted to, reporting any signature carried inside it.
func (p *PGP) decrypt(ciphertext, passphrase []byte) (*Opened, error) {
	private, err := p.keys.PrivateKeys()
	if err != nil {
		return nil, err
	}
	if len(private) == 0 {
		return nil, ErrNoDecryptionKey
	}
	// a signature inside an encrypted message is checked against the public
	// keyring, so both rings go to the reader.
	public, err := p.keys.PublicKeys()
	if err != nil {
		return nil, err
	}
	keyring := append(openpgp.EntityList{}, private...)
	keyring = append(keyring, public...)

	block, err := armor.Decode(bytes.NewReader(ciphertext))
	if err != nil {
		return nil, ErrMalformedArmor
	}

	// prompt is called once per candidate key. Returning an error rather than
	// nil on the second call stops go-crypto retrying the same wrong passphrase
	// against every key in the ring.
	var asked bool
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if asked || len(passphrase) == 0 {
			return nil, ErrPassphraseRequired
		}
		asked = true
		return passphrase, nil
	}

	md, err := openpgp.ReadMessage(block.Body, keyring, prompt, cryptoConfig())
	if err != nil {
		return nil, decryptError(err)
	}
	plaintext, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, decryptError(err)
	}
	// the signature state is only final once the body has been read to the end,
	// which is why this is checked here rather than off md alone.
	opened := &Opened{Body: plaintext, Encrypted: true}
	switch {
	case !md.IsSigned:
		// encrypted but unsigned: confidential, with nothing establishing who
		// wrote it.
	case md.SignedBy == nil:
		opened.Signature = unknownSigner()
	default:
		opened.Signature = signatureVerdict(md.SignedBy.Entity, md.SignatureError)
	}
	return opened, nil
}

// checkDetached verifies an armored detached signature over content.
func (p *PGP) checkDetached(content, sigArmor []byte) Signature {
	keyring, err := p.keys.PublicKeys()
	if err != nil || len(keyring) == 0 {
		return unknownSigner()
	}
	signer, err := openpgp.CheckArmoredDetachedSignature(
		keyring, bytes.NewReader(content), bytes.NewReader(sigArmor), cryptoConfig())
	if err == nil {
		return signatureVerdict(signer, nil)
	}
	// mail that passed through something that rewrote its line endings would
	// otherwise be reported as altered, so the canonical form gets a second try.
	canonical := toCRLF(content)
	if !bytes.Equal(canonical, content) {
		if signer, retry := openpgp.CheckArmoredDetachedSignature(
			keyring, bytes.NewReader(canonical), bytes.NewReader(sigArmor), cryptoConfig()); retry == nil {
			return signatureVerdict(signer, nil)
		}
	}
	return signatureVerdict(nil, err)
}

// signatureVerdict turns a verification outcome into the same three-state
// verdict S/MIME reports, so the reading pane has one way to describe a
// signature whatever produced it.
func signatureVerdict(signer *openpgp.Entity, err error) Signature {
	if err != nil {
		if isUnknownIssuer(err) {
			return unknownSigner()
		}
		return Signature{Status: SigInvalid, Detail: "the message was altered after it was signed"}
	}
	if signer == nil {
		return unknownSigner()
	}
	name, email := identityOf(signer)
	return Signature{
		Status:      SigValid,
		SignerName:  name,
		SignerEmail: email,
		Fingerprint: Fingerprint(signer),
	}
}

// unknownSigner is the verdict for a signature made by a key that is not in the
// keyring. The content may be untouched, but with no key there is no way to
// establish that, so it is reported as unverified rather than good or bad.
func unknownSigner() Signature {
	return Signature{
		Status: SigUntrusted,
		Detail: "the signing key is not in your keyring, so the signature cannot be checked",
	}
}

// isUnknownIssuer reports whether a verification failure was caused by a
// missing key rather than by the content not matching.
func isUnknownIssuer(err error) bool {
	return errors.Is(err, pgperrors.ErrUnknownIssuer) ||
		strings.Contains(err.Error(), "unknown issuer")
}

// identityOf reads a display name and address out of a key's primary user id.
func identityOf(ent *openpgp.Entity) (name, email string) {
	for _, id := range ent.Identities {
		if id.UserId == nil {
			continue
		}
		if email == "" && id.UserId.Email != "" {
			name, email = id.UserId.Name, id.UserId.Email
		}
	}
	return name, email
}

// decryptError maps a go-crypto failure onto the sentinel a caller can act on.
func decryptError(err error) error {
	switch {
	case errors.Is(err, ErrPassphraseRequired):
		return ErrPassphraseRequired
	case errors.Is(err, pgperrors.ErrKeyIncorrect):
		return ErrWrongPassphrase
	case strings.Contains(err.Error(), "no suitable key"),
		strings.Contains(err.Error(), "cannot decrypt"):
		return ErrNoDecryptionKey
	}
	return fmt.Errorf("crypto: decrypt: %w", err)
}

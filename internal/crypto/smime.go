package crypto

import "errors"

// ErrSMIMENotSupported is returned by the S/MIME engine. S/MIME is wired into
// the same Engine interface and the same hard-fail safety contract as PGP, but
// the actual signing and encryption is not implemented yet, so every call fails
// cleanly instead of producing a broken or plaintext message.
var ErrSMIMENotSupported = errors.New("crypto: s/mime is not supported yet")

// SMIME is a placeholder Engine for S/MIME (RFC 8551).
//
// Maintenance reality, stated plainly: there is no adequately maintained pure-Go
// S/MIME library at the time of writing. The commonly cited options are effectively
// abandoned:
//
//   - github.com/InfiniteLoopSpace/go_S-MIME: no meaningful activity for years,
//     no tagged releases, would pull in unreviewed PKCS#7/CMS code on the
//     critical path of an encryption feature.
//   - github.com/mastahyeti/cms and its forks: also stale, same concern.
//
// Shipping signing and encryption on top of an unmaintained CMS implementation
// would be the wrong call for a security feature, so S/MIME is deliberately
// stubbed rather than half-implemented. PGP/MIME is fully implemented and is the
// recommended path today.
//
// When this is revisited, the most credible direction is a vendored, audited
// PKCS#7/CMS layer (or a cgo binding to a maintained library behind a build tag),
// implementing application/pkcs7-mime for encryption and multipart/signed with
// application/pkcs7-signature for signing, x509 certs and keys looked up the same
// way PGP keys are. Until then this stub keeps the interface honest.
type SMIME struct{}

// NewSMIME returns the S/MIME placeholder engine.
func NewSMIME() *SMIME {
	return &SMIME{}
}

// Wrap always fails with ErrSMIMENotSupported. Because it returns a nil Part,
// the smtp layer cannot transmit anything for an S/MIME-protected message, which
// keeps the safety rule intact: a message meant to be protected is never sent in
// the clear.
func (*SMIME) Wrap(entity []byte, mode Mode, opts Options) (*Part, error) {
	return nil, ErrSMIMENotSupported
}

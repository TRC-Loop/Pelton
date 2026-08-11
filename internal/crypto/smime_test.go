package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
)

// signedPart is the message content the tests sign, with the crlf line endings
// mail is transmitted with.
const signedPart = "Content-Type: text/plain\r\n\r\nthe quick brown fox\r\n"

// testCA is a certificate authority the tests install as the trust root.
type testCA struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
	der  []byte
}

func newTestCA(t *testing.T) testCA {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate ca key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Pelton Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create ca cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse ca cert: %v", err)
	}
	return testCA{cert: cert, key: key, der: der}
}

// issue returns a signing certificate for email, issued by ca. A zero notAfter
// means a normally valid certificate.
func (ca testCA) issue(t *testing.T, email string, notAfter time.Time) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signer key: %v", err)
	}
	if notAfter.IsZero() {
		notAfter = time.Now().Add(12 * time.Hour)
	}
	tmpl := &x509.Certificate{
		SerialNumber:   big.NewInt(2),
		Subject:        pkix.Name{CommonName: "Alice Example"},
		EmailAddresses: []string{email},
		NotBefore:      time.Now().Add(-time.Hour),
		NotAfter:       notAfter,
		KeyUsage:       x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatalf("create signer cert: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse signer cert: %v", err)
	}
	return cert, key
}

// trust installs pool as the platform trust store for the duration of a test.
func trust(t *testing.T, certs ...*x509.Certificate) {
	t.Helper()
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	original := systemRoots
	systemRoots = func() (*x509.CertPool, error) { return pool, nil }
	t.Cleanup(func() { systemRoots = original })
}

// detachedSignature signs content and returns the base64 der of the detached
// CMS structure.
func detachedSignature(t *testing.T, content []byte, cert *x509.Certificate, key *rsa.PrivateKey) []byte {
	t.Helper()
	sd, err := pkcs7.NewSignedData(content)
	if err != nil {
		t.Fatalf("new signed data: %v", err)
	}
	if err := sd.AddSigner(cert, key, pkcs7.SignerInfoConfig{}); err != nil {
		t.Fatalf("add signer: %v", err)
	}
	sd.Detach()
	der, err := sd.Finish()
	if err != nil {
		t.Fatalf("finish signed data: %v", err)
	}
	return der
}

// multipartSigned assembles the multipart/signed message real S/MIME mail uses.
func multipartSigned(from, content string, der []byte) []byte {
	const boundary = "pelton-test-boundary"
	b64 := base64.StdEncoding.EncodeToString(der)
	var wrapped strings.Builder
	for i := 0; i < len(b64); i += 64 {
		end := min(i+64, len(b64))
		wrapped.WriteString(b64[i:end])
		wrapped.WriteString("\r\n")
	}
	return []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"Subject: signed\r\n"+
			"Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"; "+
			"micalg=sha-256; boundary=\"%s\"\r\n"+
			"\r\n"+
			"--%s\r\n"+
			"%s"+
			"\r\n--%s\r\n"+
			"Content-Type: application/pkcs7-signature; name=\"smime.p7s\"\r\n"+
			"Content-Transfer-Encoding: base64\r\n"+
			"\r\n"+
			"%s"+
			"\r\n--%s--\r\n",
		from, boundary, boundary, content, boundary, wrapped.String(), boundary))
}

func TestVerifyTrustedSignature(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "alice@example.test", time.Time{})
	trust(t, ca.cert)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	raw := multipartSigned("Alice Example <alice@example.test>", signedPart, der)

	sig := VerifySMIME(raw, "Alice Example <alice@example.test>")
	if sig.Status != SigValid {
		t.Fatalf("Status = %q (%s), want %q", sig.Status, sig.Detail, SigValid)
	}
	if sig.SignerEmail != "alice@example.test" {
		t.Errorf("SignerEmail = %q", sig.SignerEmail)
	}
	if sig.SignerName != "Alice Example" {
		t.Errorf("SignerName = %q", sig.SignerName)
	}
	if sig.Issuer != "Pelton Test CA" {
		t.Errorf("Issuer = %q", sig.Issuer)
	}
	if sig.Detail != "" {
		t.Errorf("Detail = %q, want empty for a valid signature", sig.Detail)
	}
}

// The bytes verify, but nothing vouches for who signed them. That must not read
// as a good signature.
func TestVerifyUnknownAuthority(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "alice@example.test", time.Time{})
	// an empty trust store: the issuing ca is not installed.
	trust(t)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	raw := multipartSigned("alice@example.test", signedPart, der)

	sig := VerifySMIME(raw, "alice@example.test")
	if sig.Status != SigUntrusted {
		t.Fatalf("Status = %q, want %q", sig.Status, SigUntrusted)
	}
	if !strings.Contains(sig.Detail, "does not trust") {
		t.Errorf("Detail = %q, want it to explain the untrusted authority", sig.Detail)
	}
}

// A genuine certificate issued to somebody else does not establish that this
// sender wrote the message.
func TestVerifyCertificateIssuedToSomebodyElse(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "mallory@example.test", time.Time{})
	trust(t, ca.cert)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	raw := multipartSigned("alice@example.test", signedPart, der)

	sig := VerifySMIME(raw, "alice@example.test")
	if sig.Status != SigUntrusted {
		t.Fatalf("Status = %q, want %q", sig.Status, SigUntrusted)
	}
	if !strings.Contains(sig.Detail, "mallory@example.test") {
		t.Errorf("Detail = %q, want it to name the certificate's address", sig.Detail)
	}
}

func TestVerifyExpiredCertificate(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "alice@example.test", time.Now().Add(-time.Minute))
	trust(t, ca.cert)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	raw := multipartSigned("alice@example.test", signedPart, der)

	sig := VerifySMIME(raw, "alice@example.test")
	if sig.Status != SigUntrusted {
		t.Fatalf("Status = %q, want %q", sig.Status, SigUntrusted)
	}
	if !strings.Contains(sig.Detail, "expired") {
		t.Errorf("Detail = %q, want it to say the certificate expired", sig.Detail)
	}
}

// The whole point of a signature: altering the content has to be detected.
func TestVerifyDetectsTamperedContent(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "alice@example.test", time.Time{})
	trust(t, ca.cert)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	tampered := strings.Replace(signedPart, "quick brown fox", "slow brown fox", 1)
	raw := multipartSigned("alice@example.test", tampered, der)

	sig := VerifySMIME(raw, "alice@example.test")
	if sig.Status != SigInvalid {
		t.Fatalf("Status = %q, want %q", sig.Status, SigInvalid)
	}
}

func TestVerifyUnsignedMessage(t *testing.T) {
	raw := []byte("From: bob@example.test\r\nContent-Type: text/plain\r\n\r\nhello\r\n")
	if sig := VerifySMIME(raw, "bob@example.test"); sig.Status != SigNone {
		t.Errorf("Status = %q, want %q", sig.Status, SigNone)
	}
}

// Nothing here may panic or hang on input that is not a message at all.
func TestVerifyRejectsMalformedInput(t *testing.T) {
	for _, tt := range []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"headers only", "From: a@b.test\r\n"},
		{"signed with no boundary", "Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"\r\n\r\nbody"},
		{"boundary with no parts", "Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"; boundary=\"x\"\r\n\r\nnothing here"},
		{"signature is not der", "Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"; boundary=\"x\"\r\n" +
			"\r\n--x\r\nContent-Type: text/plain\r\n\r\nhi\r\n--x\r\n" +
			"Content-Type: application/pkcs7-signature\r\nContent-Transfer-Encoding: base64\r\n\r\nbm90IGRlcg==\r\n--x--\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sig := VerifySMIME([]byte(tt.raw), "a@b.test")
			if sig.Status == SigValid {
				t.Errorf("Status = %q, malformed input must never verify", sig.Status)
			}
		})
	}
}

// A signature is computed over crlf bytes, so content that reached us with bare
// line feeds must still verify rather than being reported as tampered with.
func TestVerifyToleratesRewrittenLineEndings(t *testing.T) {
	ca := newTestCA(t)
	cert, key := ca.issue(t, "alice@example.test", time.Time{})
	trust(t, ca.cert)

	der := detachedSignature(t, []byte(signedPart), cert, key)
	unix := strings.ReplaceAll(signedPart, "\r\n", "\n")
	raw := multipartSigned("alice@example.test", unix, der)

	sig := VerifySMIME(raw, "alice@example.test")
	if sig.Status != SigValid {
		t.Fatalf("Status = %q (%s), want %q", sig.Status, sig.Detail, SigValid)
	}
}

// A certificate that expires after the message was signed is caught by chain
// validation rather than the signing-time check. Constructing that needs a
// certificate to expire mid-test, so the wording is pinned directly.
func TestChainProblemWording(t *testing.T) {
	expired := x509.CertificateInvalidError{Reason: x509.Expired}
	if got := chainProblem(expired); !strings.Contains(got, "expired") {
		t.Errorf("chainProblem(expired) = %q", got)
	}
	unknown := x509.UnknownAuthorityError{}
	if got := chainProblem(unknown); !strings.Contains(got, "does not trust") {
		t.Errorf("chainProblem(unknown) = %q", got)
	}
	if got := chainProblem(fmt.Errorf("something else")); got == "" {
		t.Error("chainProblem() returned no explanation for an unrecognised failure")
	}
}

func TestBareAddress(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"Alice Example <alice@example.test>", "alice@example.test"},
		{"alice@example.test", "alice@example.test"},
		{"  ALICE@Example.TEST  ", "alice@example.test"},
		{"not an address", ""},
		{"", ""},
	} {
		if got := bareAddress(tt.in); got != tt.want {
			t.Errorf("bareAddress(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

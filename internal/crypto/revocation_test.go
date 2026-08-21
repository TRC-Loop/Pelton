package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/ocsp"
)

// testPair builds a throwaway issuer and a certificate it signed, so the tests
// can exercise a real chain without a certificate authority.
func testPair(t *testing.T) (issuer, signer *x509.Certificate, issuerKey *ecdsa.PrivateKey) {
	t.Helper()
	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("issuer key: %v", err)
	}
	issuerTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, issuerTmpl, issuerTmpl, &issuerKey.PublicKey, issuerKey)
	if err != nil {
		t.Fatalf("issuer cert: %v", err)
	}
	if issuer, err = x509.ParseCertificate(der); err != nil {
		t.Fatalf("parse issuer: %v", err)
	}

	signerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("signer key: %v", err)
	}
	signerTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(4242),
		Subject:      pkix.Name{CommonName: "Someone"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err = x509.CreateCertificate(rand.Reader, signerTmpl, issuer, &signerKey.PublicKey, issuerKey)
	if err != nil {
		t.Fatalf("signer cert: %v", err)
	}
	if signer, err = x509.ParseCertificate(der); err != nil {
		t.Fatalf("parse signer: %v", err)
	}
	return issuer, signer, issuerKey
}

// ocspServer answers every request with the given status.
func ocspServer(t *testing.T, issuer, signer *x509.Certificate, key *ecdsa.PrivateKey, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := ocsp.Response{
			Status:       status,
			SerialNumber: signer.SerialNumber,
			ThisUpdate:   time.Now().Add(-time.Minute),
			NextUpdate:   time.Now().Add(time.Hour),
		}
		if status == ocsp.Revoked {
			tmpl.RevokedAt = time.Now().Add(-2 * time.Hour)
			tmpl.RevocationReason = ocsp.KeyCompromise
		}
		body, err := ocsp.CreateResponse(issuer, issuer, tmpl, key)
		if err != nil {
			t.Errorf("create ocsp response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/ocsp-response")
		_, _ = w.Write(body)
	}))
}

func TestRevocationGoodFromOCSP(t *testing.T) {
	issuer, signer, key := testPair(t)
	srv := ocspServer(t, issuer, signer, key, ocsp.Good)
	defer srv.Close()
	signer.OCSPServer = []string{srv.URL}

	got := CheckRevocation(context.Background(), srv.Client(), signer, issuer)
	if got.Status != RevocationGood {
		t.Fatalf("status = %q, want good (detail %q)", got.Status, got.Detail)
	}
	if got.NextUpdate.IsZero() {
		t.Error("next update is zero, so the answer would never cache")
	}
}

func TestRevocationRevokedFromOCSP(t *testing.T) {
	issuer, signer, key := testPair(t)
	srv := ocspServer(t, issuer, signer, key, ocsp.Revoked)
	defer srv.Close()
	signer.OCSPServer = []string{srv.URL}

	got := CheckRevocation(context.Background(), srv.Client(), signer, issuer)
	if got.Status != RevocationRevoked {
		t.Fatalf("status = %q, want revoked", got.Status)
	}
	if got.Detail == "" {
		t.Error("a revoked certificate must say why, so the reader knows what happened")
	}
	if got.RevokedAt.IsZero() {
		t.Error("revoked at is zero, though the responder gave one")
	}
}

// a certificate with no responder and no list is unknown, never good: nothing
// was asked, and saying otherwise would be a claim the check cannot support.
func TestRevocationUnknownWithoutResponder(t *testing.T) {
	issuer, signer, _ := testPair(t)

	got := CheckRevocation(context.Background(), http.DefaultClient, signer, issuer)
	if got.Status != RevocationUnknown {
		t.Fatalf("status = %q, want unknown", got.Status)
	}
	if got.Detail == "" {
		t.Error("unknown must explain itself, or it reads as a failure")
	}
}

// being offline is not evidence about a certificate, so an unreachable
// responder is unknown rather than revoked.
func TestRevocationUnknownWhenResponderIsDown(t *testing.T) {
	issuer, signer, key := testPair(t)
	srv := ocspServer(t, issuer, signer, key, ocsp.Good)
	url := srv.URL
	srv.Close()
	signer.OCSPServer = []string{url}

	got := CheckRevocation(context.Background(), &http.Client{Timeout: time.Second}, signer, issuer)
	if got.Status != RevocationUnknown {
		t.Fatalf("status = %q, want unknown", got.Status)
	}
}

// a crl serves the same question when the certificate names no responder.
func TestRevocationRevokedFromCRL(t *testing.T) {
	issuer, signer, key := testPair(t)
	list, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: time.Now().Add(-time.Minute),
		NextUpdate: time.Now().Add(time.Hour),
		RevokedCertificateEntries: []x509.RevocationListEntry{
			{SerialNumber: signer.SerialNumber, RevocationTime: time.Now().Add(-time.Hour)},
		},
	}, issuer, key)
	if err != nil {
		t.Fatalf("create crl: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(list)
	}))
	defer srv.Close()
	signer.CRLDistributionPoints = []string{srv.URL}

	got := CheckRevocation(context.Background(), srv.Client(), signer, issuer)
	if got.Status != RevocationRevoked {
		t.Fatalf("status = %q, want revoked", got.Status)
	}
}

// a certificate absent from the list is in force.
func TestRevocationGoodFromCRL(t *testing.T) {
	issuer, signer, key := testPair(t)
	list, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: time.Now().Add(-time.Minute),
		NextUpdate: time.Now().Add(time.Hour),
	}, issuer, key)
	if err != nil {
		t.Fatalf("create crl: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(list)
	}))
	defer srv.Close()
	signer.CRLDistributionPoints = []string{srv.URL}

	got := CheckRevocation(context.Background(), srv.Client(), signer, issuer)
	if got.Status != RevocationGood {
		t.Fatalf("status = %q, want good", got.Status)
	}
}

// a list somebody else signed proves nothing, so it must not be believed.
func TestRevocationIgnoresUnsignedCRL(t *testing.T) {
	issuer, signer, _ := testPair(t)
	_, _, otherKey := testPair(t)
	other, _, _ := testPair(t)
	list, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: time.Now().Add(-time.Minute),
		NextUpdate: time.Now().Add(time.Hour),
		RevokedCertificateEntries: []x509.RevocationListEntry{
			{SerialNumber: signer.SerialNumber, RevocationTime: time.Now()},
		},
	}, other, otherKey)
	if err != nil {
		t.Fatalf("create crl: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(list)
	}))
	defer srv.Close()
	signer.CRLDistributionPoints = []string{srv.URL}

	got := CheckRevocation(context.Background(), srv.Client(), signer, issuer)
	if got.Status == RevocationRevoked {
		t.Fatal("a revocation list signed by another authority was believed")
	}
}

func TestCertFingerprintIsStable(t *testing.T) {
	_, signer, _ := testPair(t)
	first := CertFingerprint([][]byte{signer.Raw})
	if first == "" {
		t.Fatal("fingerprint is empty")
	}
	if second := CertFingerprint([][]byte{signer.Raw}); second != first {
		t.Errorf("fingerprint changed between calls: %q then %q", first, second)
	}
	if CertFingerprint(nil) != "" {
		t.Error("no certificate must produce no fingerprint, not a hash of nothing")
	}
}

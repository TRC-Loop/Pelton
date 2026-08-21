package crypto

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
)

// RevocationStatus is what the issuing authority says about a certificate now,
// as opposed to what the certificate itself claims.
//
// A signature can verify perfectly against bytes that were signed with a stolen
// key. The chain checks in VerifySMIME cannot see that, because the certificate
// is still well formed and still within its dates; only the authority that
// issued it knows it has been withdrawn, and only over the network.
type RevocationStatus string

const (
	// RevocationUnchecked means no check was made, because the setting is off
	// or the message carries no signature.
	RevocationUnchecked RevocationStatus = ""
	// RevocationGood means the authority was asked and said the certificate is
	// still in force.
	RevocationGood RevocationStatus = "good"
	// RevocationRevoked means the authority has withdrawn the certificate. The
	// signature still verifies; it just no longer vouches for anybody.
	RevocationRevoked RevocationStatus = "revoked"
	// RevocationUnknown means the question could not be answered: offline, no
	// responder named, or a responder that refused. It is not evidence either
	// way and must not read as one.
	RevocationUnknown RevocationStatus = "unknown"
)

// Revocation is the outcome of one check.
type Revocation struct {
	Status RevocationStatus
	// Detail explains a status that is not good, in a sentence fit to show.
	Detail string
	// RevokedAt is when the authority says the certificate was withdrawn. Zero
	// unless Status is revoked, and zero even then for a responder that does
	// not say.
	RevokedAt time.Time
	// NextUpdate is when the answer stops being current, taken from the
	// responder. Zero means the caller picks its own lifetime.
	NextUpdate time.Time
}

// crlSizeLimit bounds a downloaded revocation list. Real lists run to a few
// megabytes; a responder that keeps sending past this is not worth waiting for,
// and refusing early keeps one bad url from eating memory.
const crlSizeLimit = 8 << 20

// ocspSizeLimit bounds a responder's reply. One is a couple of kilobytes.
const ocspSizeLimit = 256 << 10

// CheckRevocation asks the issuing authority whether signer is still in force.
//
// OCSP is tried first: it asks about one certificate, so the reply is small and
// current. Certificates that name no responder fall back to the issuer's
// revocation list, which is a larger download but covers the same question.
// Neither being available is RevocationUnknown, not a failure: being offline
// says nothing about a certificate.
//
// Every request goes through client, which the caller builds from the app's
// proxy settings, so turning this on cannot become a way around them.
func CheckRevocation(ctx context.Context, client *http.Client, signer, issuer *x509.Certificate) Revocation {
	if signer == nil || issuer == nil {
		return Revocation{
			Status: RevocationUnknown,
			Detail: "the issuing certificate was not included, so its authority could not be asked",
		}
	}

	if len(signer.OCSPServer) > 0 {
		if r, ok := checkOCSP(ctx, client, signer, issuer); ok {
			return r
		}
	}
	if len(signer.CRLDistributionPoints) > 0 {
		if r, ok := checkCRL(ctx, client, signer, issuer); ok {
			return r
		}
	}
	if len(signer.OCSPServer) == 0 && len(signer.CRLDistributionPoints) == 0 {
		return Revocation{
			Status: RevocationUnknown,
			Detail: "the certificate names no way to check whether it has been withdrawn",
		}
	}
	return Revocation{
		Status: RevocationUnknown,
		Detail: "the issuing authority could not be reached",
	}
}

// checkOCSP asks each responder the certificate names, in order, and reports
// the first that answers. ok is false when none did, so the caller can fall
// back rather than treat an unreachable responder as an answer.
func checkOCSP(ctx context.Context, client *http.Client, signer, issuer *x509.Certificate) (Revocation, bool) {
	req, err := ocsp.CreateRequest(signer, issuer, nil)
	if err != nil {
		return Revocation{}, false
	}
	for _, responder := range signer.OCSPServer {
		body, err := post(ctx, client, responder, "application/ocsp-request", req, ocspSizeLimit)
		if err != nil {
			continue
		}
		// ParseResponseForCert checks the responder's own signature against the
		// issuer, so a forged "good" from a hijacked url does not pass.
		res, err := ocsp.ParseResponseForCert(body, signer, issuer)
		if err != nil {
			continue
		}
		switch res.Status {
		case ocsp.Good:
			return Revocation{Status: RevocationGood, NextUpdate: res.NextUpdate}, true
		case ocsp.Revoked:
			return Revocation{
				Status:     RevocationRevoked,
				Detail:     revokedDetail(res.RevocationReason),
				RevokedAt:  res.RevokedAt,
				NextUpdate: res.NextUpdate,
			}, true
		default:
			// The responder knows the issuer but not this certificate. That is
			// an answer, but not one to build on, so let the crl have a go.
			continue
		}
	}
	return Revocation{}, false
}

// checkCRL downloads the issuer's revocation list and looks for the serial.
func checkCRL(ctx context.Context, client *http.Client, signer, issuer *x509.Certificate) (Revocation, bool) {
	for _, point := range signer.CRLDistributionPoints {
		body, err := get(ctx, client, point, crlSizeLimit)
		if err != nil {
			continue
		}
		list, err := parseCRL(body)
		if err != nil {
			continue
		}
		// A list is only worth reading if the issuer actually signed it.
		if err := list.CheckSignatureFrom(issuer); err != nil {
			continue
		}
		for _, entry := range list.RevokedCertificateEntries {
			if entry.SerialNumber.Cmp(signer.SerialNumber) == 0 {
				return Revocation{
					Status:     RevocationRevoked,
					Detail:     revokedDetail(int(entry.ReasonCode)),
					RevokedAt:  entry.RevocationTime,
					NextUpdate: list.NextUpdate,
				}, true
			}
		}
		return Revocation{Status: RevocationGood, NextUpdate: list.NextUpdate}, true
	}
	return Revocation{}, false
}

// parseCRL reads a list in either encoding. Distribution points usually serve
// der, but some serve pem, and the difference is not worth a failed check.
func parseCRL(body []byte) (*x509.RevocationList, error) {
	if list, err := x509.ParseRevocationList(body); err == nil {
		return list, nil
	}
	block, _ := pem.Decode(body)
	if block == nil {
		return nil, errors.New("crypto: the revocation list could not be read")
	}
	return x509.ParseRevocationList(block.Bytes)
}

// revokedDetail turns the authority's reason code into a sentence. The codes
// that say something a reader can act on are named; the rest are not worth
// spelling out, since the fact of revocation is the point either way.
func revokedDetail(reason int) string {
	switch reason {
	case int(ocsp.KeyCompromise):
		return "the issuing authority has withdrawn this certificate: its private key was compromised"
	case int(ocsp.CACompromise):
		return "the issuing authority has withdrawn this certificate: the authority itself was compromised"
	case int(ocsp.Superseded):
		return "the issuing authority has withdrawn this certificate: it was replaced by a newer one"
	case int(ocsp.CessationOfOperation):
		return "the issuing authority has withdrawn this certificate: it is no longer in use"
	case int(ocsp.AffiliationChanged):
		return "the issuing authority has withdrawn this certificate: the details it vouches for changed"
	default:
		return "the issuing authority has withdrawn this certificate"
	}
}

func post(ctx context.Context, client *http.Client, url, contentType string, body []byte, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return send(client, req, limit)
}

func get(ctx context.Context, client *http.Client, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return send(client, req, limit)
}

func send(client *http.Client, req *http.Request, limit int64) ([]byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crypto: %s answered %s", req.URL.Host, res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, limit))
}

// CertFingerprint is the sha-256 of a signing certificate, hex encoded. It
// identifies the certificate across messages, so a thread from one sender is
// one question to its authority rather than one per message.
func CertFingerprint(certs [][]byte) string {
	if len(certs) == 0 {
		return ""
	}
	sum := sha256.Sum256(certs[0])
	return hex.EncodeToString(sum[:])
}

// ParseCertPair reads the signer and its issuer back out of stored DER.
func ParseCertPair(certs [][]byte) (signer, issuer *x509.Certificate, err error) {
	if len(certs) < 2 {
		return nil, nil, errors.New("crypto: the stored certificate chain is incomplete")
	}
	if signer, err = x509.ParseCertificate(certs[0]); err != nil {
		return nil, nil, fmt.Errorf("crypto: signing certificate: %w", err)
	}
	if issuer, err = x509.ParseCertificate(certs[1]); err != nil {
		return nil, nil, fmt.Errorf("crypto: issuing certificate: %w", err)
	}
	return signer, issuer, nil
}

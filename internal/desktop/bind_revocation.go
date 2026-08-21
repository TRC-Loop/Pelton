package desktop

import (
	"context"
	"errors"
	"time"

	"github.com/TRC-Loop/Pelton/internal/crypto"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// settingSMIMERevocation turns on asking each sender's certificate authority
// whether their certificate is still in force. Off by default, and deliberately
// so: the question names the certificate, so the authority learns which of its
// certificates you are validating, and therefore roughly who mails you and
// when you read it. That is a third party in the mail path, which is the one
// thing this client does not do without being asked.
const settingSMIMERevocation = "smime_revocation"

// revocationTimeout bounds one check. An authority that has not answered by
// then is treated as unreachable, which leaves the signature as it was rather
// than condemning it.
const revocationTimeout = 10 * time.Second

// revocationFallback is how long an answer holds when the authority named no
// expiry of its own. A day is short enough that a withdrawal is noticed
// promptly and long enough that reading a thread is one question, not twenty.
const revocationFallback = 24 * time.Hour

// SMIMERevocationDTO is what the reader is told about a certificate's standing.
type SMIMERevocationDTO struct {
	// Status is good, revoked, unknown, or '' when no check was made.
	Status string `json:"status"`
	// Detail explains anything that is not good, in a sentence fit to show.
	Detail string `json:"detail"`
	// RevokedAt is when the authority withdrew the certificate, RFC 3339, or ''.
	RevokedAt string `json:"revokedAt"`
	// Checked is when the answer was obtained, RFC 3339, or '' when none was.
	Checked string `json:"checked"`
}

// SMIMERevocationEnabled reports whether revocation checking is on.
func (a *App) SMIMERevocationEnabled() bool {
	return a.boolSetting(settingSMIMERevocation, false)
}

// SetSMIMERevocation turns revocation checking on or off.
//
// Turning it off empties the cache. A stored answer is a record of having asked
// an authority about a correspondent, and leaving those behind after the reader
// said stop would keep the very thing the setting exists to avoid.
func (a *App) SetSMIMERevocation(on bool) error {
	if err := a.store.SetBool(a.ctx, settingSMIMERevocation, on); err != nil {
		return err
	}
	if !on {
		return a.store.ClearRevocations(a.ctx)
	}
	return nil
}

// CheckSMIMERevocation reports whether the certificate that signed a message
// has since been withdrawn by the authority that issued it.
//
// The frontend calls this when a signed message is opened rather than at sync:
// a certificate valid when the mail arrived can be withdrawn the week after,
// which is exactly the case a stored verdict would miss. Answers are cached per
// certificate for as long as the authority says they hold, so a thread from one
// sender is one question.
func (a *App) CheckSMIMERevocation(messageID int64) (SMIMERevocationDTO, error) {
	if !a.SMIMERevocationEnabled() {
		return SMIMERevocationDTO{}, nil
	}
	msg, err := a.store.GetMessage(a.ctx, messageID)
	if err != nil {
		return SMIMERevocationDTO{}, err
	}
	// unsigned mail, and a signature whose chain never validated, have no
	// certificate to ask about.
	if msg.SMIME.Fingerprint == "" {
		return SMIMERevocationDTO{}, nil
	}

	if cached, err := a.store.RevocationFor(a.ctx, msg.SMIME.Fingerprint); err == nil {
		if cached.Fresh(time.Now(), revocationFallback) {
			return revocationDTO(*cached), nil
		}
	} else if !errors.Is(err, storage.ErrRevocationNotCached) {
		return SMIMERevocationDTO{}, err
	}

	record, err := a.askAuthority(msg.ID, msg.SMIME.Fingerprint)
	if err != nil {
		return SMIMERevocationDTO{}, err
	}
	// a failure to cache costs an extra request later, not an answer.
	if err := a.store.SaveRevocation(a.ctx, record); err != nil {
		a.log.Error("cache revocation answer", "message", messageID, "err", err)
	}
	return revocationDTO(record), nil
}

// askAuthority runs the network check for one message's signing certificate.
func (a *App) askAuthority(messageID int64, fingerprint string) (storage.RevocationRecord, error) {
	certs, err := a.store.MessageSMIMECerts(a.ctx, messageID)
	if err != nil {
		return storage.RevocationRecord{}, err
	}
	now := time.Now()
	signer, issuer, err := crypto.ParseCertPair(certs)
	if err != nil {
		return storage.RevocationRecord{
			Fingerprint: fingerprint,
			Status:      string(crypto.RevocationUnknown),
			Detail:      "the stored certificate could not be read, so its authority could not be asked",
			CheckedAt:   now,
		}, nil
	}

	ctx, cancel := context.WithTimeout(a.ctx, revocationTimeout)
	defer cancel()
	result := crypto.CheckRevocation(ctx, a.httpClient(revocationTimeout), signer, issuer)
	return storage.RevocationRecord{
		Fingerprint: fingerprint,
		Status:      string(result.Status),
		Detail:      result.Detail,
		RevokedAt:   result.RevokedAt,
		CheckedAt:   now,
		NextUpdate:  result.NextUpdate,
	}, nil
}

func revocationDTO(r storage.RevocationRecord) SMIMERevocationDTO {
	return SMIMERevocationDTO{
		Status:    r.Status,
		Detail:    r.Detail,
		RevokedAt: formatDate(r.RevokedAt),
		Checked:   formatDate(r.CheckedAt),
	}
}

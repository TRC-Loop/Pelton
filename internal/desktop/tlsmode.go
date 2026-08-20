package desktop

import (
	"errors"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	psmtp "github.com/TRC-Loop/Pelton/internal/smtp"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// errUnknownTLSMode rejects a security value the frontend should never send.
var errUnknownTLSMode = errors.New("unknown connection security")

// Connection security as stored on an account and exchanged with the frontend.
// Empty means derive it from the port, which is how every account created
// before it was storable behaves.
const (
	tlsModeAuto     = ""
	tlsModeSSL      = "ssl"
	tlsModeStartTLS = "starttls"
)

// validTLSMode reports whether a value is one the frontend may send.
func validTLSMode(mode string) bool {
	switch mode {
	case tlsModeAuto, tlsModeSSL, tlsModeStartTLS:
		return true
	}
	return false
}

// imapTLSMode maps the stored value onto the imap layer's own enum. Anything
// unrecognised falls back to deriving from the port rather than guessing, so a
// bad value cannot silently downgrade a connection to cleartext.
func imapTLSMode(mode string) pimap.TLSMode {
	switch mode {
	case tlsModeSSL:
		return pimap.TLSImplicit
	case tlsModeStartTLS:
		return pimap.TLSStartTLS
	default:
		return pimap.TLSAuto
	}
}

// effectiveIMAPTLS and effectiveSMTPTLS report the security an account actually
// connects with: the stored value, or what the port resolves to when nothing is
// pinned. The imap and smtp layers apply exactly this rule when handed TLSAuto,
// so reporting it here lets the ui show the truth for an account that predates
// the setting without repeating the rule, and drifting from it later.
func effectiveIMAPTLS(a storage.Account) string {
	if a.IMAPTLS != tlsModeAuto {
		return a.IMAPTLS
	}
	if a.IMAPPort == pimap.PortStartTLS {
		return tlsModeStartTLS
	}
	return tlsModeSSL
}

func effectiveSMTPTLS(a storage.Account) string {
	if a.SMTPTLS != tlsModeAuto {
		return a.SMTPTLS
	}
	if a.SMTPPort == psmtp.PortStartTLS {
		return tlsModeStartTLS
	}
	return tlsModeSSL
}

// smtpTLSMode is imapTLSMode for the smtp layer.
func smtpTLSMode(mode string) psmtp.TLSMode {
	switch mode {
	case tlsModeSSL:
		return psmtp.TLSImplicit
	case tlsModeStartTLS:
		return psmtp.TLSStartTLS
	default:
		return psmtp.TLSAuto
	}
}

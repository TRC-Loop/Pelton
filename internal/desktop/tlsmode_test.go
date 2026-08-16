package desktop

import (
	"testing"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	psmtp "github.com/TRC-Loop/Pelton/internal/smtp"
)

func TestTLSModeMapping(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		wantIMAP pimap.TLSMode
		wantSMTP psmtp.TLSMode
		wantOK   bool
	}{
		{"empty derives from the port", "", pimap.TLSAuto, psmtp.TLSAuto, true},
		{"ssl is implicit tls", "ssl", pimap.TLSImplicit, psmtp.TLSImplicit, true},
		{"starttls upgrades", "starttls", pimap.TLSStartTLS, psmtp.TLSStartTLS, true},
		// an unrecognised value must not fall through to cleartext; deriving
		// from the port is the safe reading.
		{"garbage falls back to auto", "plaintext", pimap.TLSAuto, psmtp.TLSAuto, false},
		{"case is not normalised", "SSL", pimap.TLSAuto, psmtp.TLSAuto, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := imapTLSMode(tt.mode); got != tt.wantIMAP {
				t.Errorf("imapTLSMode(%q) = %v, want %v", tt.mode, got, tt.wantIMAP)
			}
			if got := smtpTLSMode(tt.mode); got != tt.wantSMTP {
				t.Errorf("smtpTLSMode(%q) = %v, want %v", tt.mode, got, tt.wantSMTP)
			}
			if got := validTLSMode(tt.mode); got != tt.wantOK {
				t.Errorf("validTLSMode(%q) = %v, want %v", tt.mode, got, tt.wantOK)
			}
		})
	}
}

// A STARTTLS account on a custom port is the case #237 reported: the port says
// nothing useful, so only the stored mode can get the transport right.
func TestTLSModeCustomPortStillStartTLS(t *testing.T) {
	cfg := pimap.Config{Host: "127.0.0.1", Port: 1143, TLS: imapTLSMode(tlsModeStartTLS)}
	if cfg.TLS != pimap.TLSStartTLS {
		t.Fatalf("imap TLS = %v, want TLSStartTLS", cfg.TLS)
	}

	send := psmtp.Config{Host: "127.0.0.1", Port: 1025, TLS: smtpTLSMode(tlsModeStartTLS)}
	if send.TLS != psmtp.TLSStartTLS {
		t.Fatalf("smtp TLS = %v, want TLSStartTLS", send.TLS)
	}
}

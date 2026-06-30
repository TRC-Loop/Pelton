// Package imap is Pelton's IMAP client over github.com/emersion/go-imap/v2:
// connect, auth, list folders, fetch headers and messages, flags, and IDLE.
package imap

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

const (
	// DefaultPort is IMAP over implicit TLS (RFC 8314).
	DefaultPort = 993
	// PortStartTLS is the cleartext IMAP port upgraded to TLS with STARTTLS.
	PortStartTLS = 143

	dialTimeout  = 30 * time.Second
	updateBuffer = 16
)

// TLSMode selects how the connection is secured.
type TLSMode int

const (
	// TLSAuto picks implicit TLS or STARTTLS from the port (143 = STARTTLS,
	// anything else = implicit TLS). This matches how the smtp layer behaves.
	TLSAuto TLSMode = iota
	// TLSImplicit dials straight into TLS (port 993).
	TLSImplicit
	// TLSStartTLS connects in cleartext then issues STARTTLS (port 143).
	TLSStartTLS
)

// Config holds connection and auth parameters. Gmail and iCloud need an
// app-specific password, not the account password.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string

	// TLS selects implicit TLS vs STARTTLS. The zero value (TLSAuto) derives the
	// mode from the port, so callers that only set a port get the right transport.
	TLS TLSMode

	// OAuth2Token, when set, authenticates with XOAUTH2 (gmail, outlook) instead
	// of a password. The caller obtains and refreshes it; this layer only uses it.
	OAuth2Token string

	// InsecureSkipVerify disables TLS verification. Debugging only.
	InsecureSkipVerify bool
	// DebugWriter receives the raw protocol stream, including credentials.
	DebugWriter io.Writer
}

// tlsMode resolves the effective TLS mode, deriving it from the port when set to
// TLSAuto.
func (c Config) tlsMode() TLSMode {
	if c.TLS != TLSAuto {
		return c.TLS
	}
	if c.Port == PortStartTLS {
		return TLSStartTLS
	}
	return TLSImplicit
}

// MailboxUpdate is a server push delivered while idling; one field is set.
type MailboxUpdate struct {
	NumMessages    *uint32 // new mail (EXISTS)
	ExpungedSeqNum *uint32 // message removed (EXPUNGE)
}

// Client is a stateful IMAP client. Not safe for concurrent use.
type Client struct {
	raw     *imapclient.Client
	cfg     Config
	updates chan MailboxUpdate
}

// Connect opens a TLS connection but does not authenticate; call Login next.
func Connect(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("imap: host is required")
	}
	if cfg.Username == "" || (cfg.Password == "" && cfg.OAuth2Token == "") {
		return nil, fmt.Errorf("imap: username and a password or oauth token are required")
	}

	port := cfg.Port
	if port == 0 {
		port = DefaultPort
	}

	updates := make(chan MailboxUpdate, updateBuffer)

	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			ServerName:         cfg.Host, // needed for hostname verification
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		},
		DebugWriter: cfg.DebugWriter,
		Dialer:      &net.Dialer{Timeout: dialTimeout},
		// only way go-imap surfaces unsolicited EXISTS/EXPUNGE during IDLE
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{
			Mailbox: func(data *imapclient.UnilateralDataMailbox) {
				if data.NumMessages != nil {
					sendUpdate(updates, MailboxUpdate{NumMessages: data.NumMessages})
				}
			},
			Expunge: func(seqNum uint32) {
				n := seqNum
				sendUpdate(updates, MailboxUpdate{ExpungedSeqNum: &n})
			},
		},
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	// implicit TLS dials straight into TLS; STARTTLS connects in cleartext and
	// upgrades. the live connection test in the wizard validates the choice.
	dial := imapclient.DialTLS
	if cfg.tlsMode() == TLSStartTLS {
		dial = imapclient.DialStartTLS
	}
	raw, err := dial(addr, options)
	if err != nil {
		return nil, fmt.Errorf("imap: dial %s: %w", addr, err)
	}

	return &Client{raw: raw, cfg: cfg, updates: updates}, nil
}

// sendUpdate never blocks the read loop; drops if the consumer is behind.
func sendUpdate(ch chan MailboxUpdate, u MailboxUpdate) {
	select {
	case ch <- u:
	default:
	}
}

// Login authenticates with the credentials from Config: XOAUTH2 when an oauth
// token is present, otherwise a password LOGIN.
func (c *Client) Login() error {
	if c.cfg.OAuth2Token != "" {
		if err := c.raw.Authenticate(newXOAuth2Client(c.cfg.Username, c.cfg.OAuth2Token)); err != nil {
			return fmt.Errorf("imap: xoauth2 auth as %q: %w", c.cfg.Username, err)
		}
		return nil
	}
	if err := c.raw.Login(c.cfg.Username, c.cfg.Password).Wait(); err != nil {
		return fmt.Errorf("imap: login as %q: %w", c.cfg.Username, err)
	}
	return nil
}

// Updates returns the channel of server pushes delivered while idling.
func (c *Client) Updates() <-chan MailboxUpdate {
	return c.updates
}

// SupportsIdle reports whether the server advertises IDLE.
func (c *Client) SupportsIdle() bool {
	return c.raw.Caps().Has(imap.CapIdle)
}

// Logout ends the session; still call Close afterwards.
func (c *Client) Logout() error {
	if err := c.raw.Logout().Wait(); err != nil {
		return fmt.Errorf("imap: logout: %w", err)
	}
	return nil
}

// Close terminates the connection.
func (c *Client) Close() error {
	if err := c.raw.Close(); err != nil {
		return fmt.Errorf("imap: close connection: %w", err)
	}
	return nil
}

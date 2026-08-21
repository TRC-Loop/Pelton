package imap

import (
	"errors"
	"net"
	"testing"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"
)

// loginServer starts an in-memory imap server with one user and returns an
// unauthenticated client pointed at it, so Login can be exercised for real.
func loginServer(t *testing.T, cfg Config) *Client {
	t.Helper()

	mem := imapmemserver.New()
	user := imapmemserver.NewUser("user", "pass")
	if err := user.Create("INBOX", nil); err != nil {
		t.Fatalf("create INBOX: %v", err)
	}
	mem.AddUser(user)

	srv := imapserver.New(&imapserver.Options{
		NewSession: func(*imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return mem.NewSession(), nil, nil
		},
		InsecureAuth: true,
		Logger:       discardLogger{},
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	raw := imapclient.New(conn, nil)
	t.Cleanup(func() { _ = raw.Close() })
	return &Client{raw: raw, cfg: cfg}
}

// A rejected password has to be distinguishable from an unreachable server:
// the app marks the mailbox and asks for a new password on the first, and just
// retries later on the second.
func TestLoginWrongPasswordIsAnAuthFailure(t *testing.T) {
	c := loginServer(t, Config{Username: "user", Password: "not-the-password"})

	err := c.Login()
	if err == nil {
		t.Fatal("login with a wrong password succeeded")
	}
	if !errors.Is(err, ErrAuthFailed) {
		t.Errorf("err = %v, want it to wrap ErrAuthFailed", err)
	}
}

func TestLoginSucceeds(t *testing.T) {
	c := loginServer(t, Config{Username: "user", Password: "pass"})

	if err := c.Login(); err != nil {
		t.Fatalf("login: %v", err)
	}
}

// A connection that never reaches a server says nothing about the password, so
// it must not be reported as one being wrong.
func TestLoginNetworkFailureIsNotAnAuthFailure(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// nothing is ever served on it, and both ends are shut, so the command
	// fails on the connection rather than with a status response.
	_ = ln.Close()
	_ = conn.Close()

	c := &Client{raw: imapclient.New(conn, nil), cfg: Config{Username: "user", Password: "pass"}}
	if err := c.Login(); err == nil {
		t.Fatal("login against a closed connection succeeded")
	} else if errors.Is(err, ErrAuthFailed) {
		t.Errorf("a connection failure was reported as an auth failure: %v", err)
	}
}

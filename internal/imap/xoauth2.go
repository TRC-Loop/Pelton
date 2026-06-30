package imap

import (
	"fmt"

	"github.com/emersion/go-sasl"
)

// xoauth2Mech is the SASL mechanism name for the XOAUTH2 scheme gmail and
// outlook use. go-sasl does not ship it, so we implement the tiny client here.
const xoauth2Mech = "XOAUTH2"

// xoauth2Client implements the XOAUTH2 SASL mechanism. The initial response is
// "user=<user>^Aauth=Bearer <token>^A^A". This mirrors the smtp layer's client
// so imap and smtp authenticate oauth the same way.
type xoauth2Client struct {
	username string
	token    string
}

func newXOAuth2Client(username, token string) sasl.Client {
	return &xoauth2Client{username: username, token: token}
}

func (a *xoauth2Client) Start() (mech string, ir []byte, err error) {
	ir = []byte("user=" + a.username + "\x01auth=Bearer " + a.token + "\x01\x01")
	return xoauth2Mech, ir, nil
}

func (a *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	// a challenge means the server rejected the token and is returning an error
	// payload; a valid xoauth2 exchange has no second step.
	return nil, fmt.Errorf("imap: xoauth2 rejected: %s", challenge)
}

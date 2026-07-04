// Package oauth runs the per-user PKCE OAuth2 flow for providers that require it
// (gmail, outlook) and refreshes access tokens. It uses the loopback redirect
// approach Thunderbird uses: a short-lived local http server catches the
// callback. There is no client secret; the user supplies their own client id
// (registered as a desktop/installed app), so PKCE is the security mechanism.
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/oauth2"
)

// Provider is the endpoint and scope set for one oauth provider.
type Provider struct {
	Label    string
	AuthURL  string
	TokenURL string
	Scopes   []string
}

// providers are the built-in oauth providers. scopes request full imap/smtp mail
// access plus a refresh token (offline access).
var providers = map[string]Provider{
	"google": {
		Label:    "Google",
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
		Scopes:   []string{"https://mail.google.com/", "email"},
	},
	"microsoft": {
		Label:    "Microsoft",
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		Scopes: []string{
			"offline_access",
			"https://outlook.office.com/IMAP.AccessAsUser.All",
			"https://outlook.office.com/SMTP.Send",
		},
	},
}

// Providers returns the known provider keys and labels for the ui.
func Providers() map[string]string {
	out := make(map[string]string, len(providers))
	for key, p := range providers {
		out[key] = p.Label
	}
	return out
}

// Authorize runs the interactive PKCE flow: it stands up a loopback callback
// server, calls open with the consent url (the app opens the system browser),
// waits for the redirect, and exchanges the code for tokens. The returned token
// carries the refresh token to persist. Cancel ctx (or let it time out) to abort.
func Authorize(ctx context.Context, providerKey, clientID, clientSecret, loginHint string, open func(string)) (*oauth2.Token, error) {
	p, ok := providers[providerKey]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", providerKey)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("oauth: open loopback listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	conf := config(p, clientID, clientSecret, redirect)

	verifier := oauth2.GenerateVerifier()
	state, err := randomState()
	if err != nil {
		return nil, err
	}

	opts := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(verifier),
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce, // force a refresh token on repeat consent (google)
	}
	if loginHint != "" {
		opts = append(opts, oauth2.SetAuthURLParam("login_hint", loginHint))
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	server := &http.Server{Handler: callbackHandler(state, codeCh, errCh)}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	defer server.Close()

	open(conf.AuthCodeURL(state, opts...))

	select {
	case code := <-codeCh:
		token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("oauth: token exchange: %w", err)
		}
		return token, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// FreshToken returns a valid access token for an account, refreshing it from the
// refresh token when needed. The returned token may carry a rotated refresh
// token the caller should persist.
func FreshToken(ctx context.Context, providerKey, clientID, clientSecret, refreshToken string) (*oauth2.Token, error) {
	p, ok := providers[providerKey]
	if !ok {
		return nil, fmt.Errorf("oauth: unknown provider %q", providerKey)
	}
	conf := config(p, clientID, clientSecret, "")
	source := conf.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	token, err := source.Token()
	if err != nil {
		return nil, fmt.Errorf("oauth: refresh token: %w", err)
	}
	return token, nil
}

// config builds the oauth2 config for a provider and client id. clientSecret is
// empty for the default public-client PKCE flow and set only for providers
// registered as confidential clients (some Microsoft Entra app registrations).
func config(p Provider, clientID, clientSecret, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes:       p.Scopes,
		Endpoint:     oauth2.Endpoint{AuthURL: p.AuthURL, TokenURL: p.TokenURL},
	}
}

// callbackHandler serves the loopback redirect, validating state and capturing
// the authorization code, then showing the user a short return-to-app page.
func callbackHandler(state string, codeCh chan<- string, errCh chan<- error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			trySend(errCh, fmt.Errorf("oauth: provider error: %s", e))
			writeClosePage(w, "Authorization failed. You can close this window.")
			return
		}
		if q.Get("state") != state {
			trySend(errCh, fmt.Errorf("oauth: state mismatch"))
			writeClosePage(w, "Authorization failed. You can close this window.")
			return
		}
		code := q.Get("code")
		if code == "" {
			trySend(errCh, fmt.Errorf("oauth: no code in callback"))
			writeClosePage(w, "Authorization failed. You can close this window.")
			return
		}
		writeClosePage(w, "Signed in. You can close this window and return to Pelton.")
		trySend(codeCh, code)
	})
	return mux
}

// trySend delivers v on ch without blocking, so a second callback request
// (browser prefetch, reload, antivirus scan) never wedges the handler.
func trySend[T any](ch chan<- T, v T) {
	select {
	case ch <- v:
	default:
	}
}

// writeClosePage renders a minimal confirmation page in the user's browser.
func writeClosePage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!doctype html><meta charset=utf-8><body style=\"font-family:sans-serif;padding:40px;text-align:center\"><h2>Pelton</h2><p>%s</p></body>", message)
}

// randomState returns a random oauth state value.
func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("oauth: generate state: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

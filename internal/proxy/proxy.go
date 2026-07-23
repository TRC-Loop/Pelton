// Package proxy turns Pelton's saved proxy preference into the two things the
// rest of the app needs: a TCP dialer for the mail connections (IMAP and SMTP)
// and an *http.Client for the app's outbound web calls (autoconfig, update
// check, unsubscribe). SOCKS5 and HTTP CONNECT proxies are supported, as well
// as a "system" mode that honours the standard proxy environment variables
// (HTTP_PROXY / HTTPS_PROXY / ALL_PROXY / NO_PROXY).
//
// The password is never stored here; it is supplied by the caller from the OS
// keyring. When no proxy is configured every helper falls back to a direct
// connection, so the default path is unchanged.
package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/http/httpproxy"
	xproxy "golang.org/x/net/proxy"
)

// Mode selects where proxy settings come from.
const (
	// ModeOff makes every connection direct.
	ModeOff = "off"
	// ModeManual uses the Scheme/Host/Port/credentials in the config.
	ModeManual = "manual"
	// ModeSystem follows the standard proxy environment variables.
	ModeSystem = "system"
)

// Supported manual proxy schemes.
const (
	SchemeSOCKS5 = "socks5"
	SchemeHTTP   = "http"
)

// dialTimeout bounds establishing the tcp connection to the proxy (or origin).
const dialTimeout = 30 * time.Second

// Config is the resolved proxy preference. Password is filled from the keyring
// by the caller, not persisted with the rest of the fields.
type Config struct {
	Mode     string `json:"mode"`
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"`
}

// Enabled reports whether any proxying is requested.
func (c Config) Enabled() bool {
	return c.Mode == ModeManual || c.Mode == ModeSystem
}

// Validate rejects a manual config that could not dial, so the settings UI can
// refuse to save something unusable. Off and system need no fields.
func (c Config) Validate() error {
	if c.Mode == ModeManual {
		if c.Scheme != SchemeSOCKS5 && c.Scheme != SchemeHTTP {
			return fmt.Errorf("proxy: unknown scheme %q", c.Scheme)
		}
		if c.Host == "" {
			return fmt.Errorf("proxy: host is required")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return fmt.Errorf("proxy: port %d out of range", c.Port)
		}
	}
	return nil
}

// address is the host:port of the manual proxy.
func (c Config) address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// proxyURL builds the URL form (with any credentials) used for the http client
// and for CONNECT dialing.
func (c Config) proxyURL() *url.URL {
	u := &url.URL{Scheme: c.Scheme, Host: c.address()}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u
}

// DialContext returns a function that opens a tcp connection through the proxy,
// suitable for the IMAP and SMTP configs' Dialer hook. A nil result (returned
// when the config is disabled) tells the caller to use its normal direct dial.
func (c Config) DialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	if !c.Enabled() {
		return nil
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		u, err := c.resolveForAddr(addr)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return (&net.Dialer{Timeout: dialTimeout}).DialContext(ctx, network, addr)
		}
		return dialThrough(ctx, u, network, addr)
	}
}

// resolveForAddr returns the proxy URL to use for a target address, or nil for a
// direct connection. In system mode this consults the environment (honouring
// NO_PROXY for addr); in manual mode it is always the configured proxy.
func (c Config) resolveForAddr(addr string) (*url.URL, error) {
	if c.Mode == ModeManual {
		return c.proxyURL(), nil
	}
	// system mode: reuse the standard library's env parsing so NO_PROXY, the
	// scheme-specific vars and lower/upper-case spellings all behave as users
	// expect. httpproxy keys off the request URL's scheme and host.
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	cfg := httpproxy.FromEnvironment()
	// mail is not http; probe with https so HTTPS_PROXY (the common "secure"
	// var) and ALL_PROXY are both considered, and NO_PROXY is applied to host.
	return cfg.ProxyFunc()(&url.URL{Scheme: "https", Host: host})
}

// dialThrough opens a tcp connection to addr via the proxy described by u.
func dialThrough(ctx context.Context, u *url.URL, network, addr string) (net.Conn, error) {
	switch u.Scheme {
	case SchemeSOCKS5, "socks5h":
		var auth *xproxy.Auth
		if pw, ok := u.User.Password(); ok {
			auth = &xproxy.Auth{User: u.User.Username(), Password: pw}
		}
		d, err := xproxy.SOCKS5(network, u.Host, auth, &net.Dialer{Timeout: dialTimeout})
		if err != nil {
			return nil, fmt.Errorf("proxy: socks5 %s: %w", u.Host, err)
		}
		if cd, ok := d.(xproxy.ContextDialer); ok {
			return cd.DialContext(ctx, network, addr)
		}
		return d.Dial(network, addr)
	case SchemeHTTP, "https":
		return httpConnect(ctx, u, addr)
	default:
		return nil, fmt.Errorf("proxy: unsupported scheme %q", u.Scheme)
	}
}

// HTTPClient returns an *http.Client that routes through the proxy, or a plain
// direct client when proxying is off. timeout bounds the whole request.
func (c Config) HTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	switch c.Mode {
	case ModeManual:
		// the transport understands both http(s):// and socks5:// proxy urls.
		transport.Proxy = http.ProxyURL(c.proxyURL())
	case ModeSystem:
		transport.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

// httpConnect tunnels a tcp connection to addr through an HTTP proxy using the
// CONNECT method (RFC 7231 section 4.3.6). This is what lets the mail
// connections (which are raw tcp, not http) traverse an http proxy; the
// standard library only tunnels CONNECT for its own http client.
func httpConnect(ctx context.Context, proxyURL *url.URL, addr string) (net.Conn, error) {
	d := &net.Dialer{Timeout: dialTimeout}
	conn, err := d.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		return nil, fmt.Errorf("proxy: connect to %s: %w", proxyURL.Host, err)
	}
	// an https proxy speaks TLS on the hop to the proxy itself.
	if proxyURL.Scheme == "https" {
		host, _, splitErr := net.SplitHostPort(proxyURL.Host)
		if splitErr != nil {
			host = proxyURL.Host
		}
		tlsConn := tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if hErr := tlsConn.HandshakeContext(ctx); hErr != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("proxy: tls handshake to %s: %w", proxyURL.Host, hErr)
		}
		conn = tlsConn
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	if proxyURL.User != nil {
		if pw, ok := proxyURL.User.Password(); ok {
			cred := proxyURL.User.Username() + ":" + pw
			req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(cred)))
		}
	}

	if err := req.Write(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy: write CONNECT: %w", err)
	}

	// the response line is all we need; the tunnel is raw after a 200.
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy: read CONNECT response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy: CONNECT to %s refused: %s", addr, resp.Status)
	}
	// a well-behaved proxy sends nothing after the blank line before the tunnel
	// opens; if it buffered ahead we would lose those bytes, so guard against it.
	if br.Buffered() > 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy: unexpected buffered data after CONNECT to %s", addr)
	}
	return conn, nil
}

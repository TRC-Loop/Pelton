package proxy

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestConfigEnabled(t *testing.T) {
	cases := []struct {
		mode string
		want bool
	}{
		{ModeOff, false},
		{"", false},
		{ModeManual, true},
		{ModeSystem, true},
	}
	for _, c := range cases {
		if got := (Config{Mode: c.mode}).Enabled(); got != c.want {
			t.Errorf("Enabled(%q) = %v, want %v", c.mode, got, c.want)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"off needs nothing", Config{Mode: ModeOff}, false},
		{"system needs nothing", Config{Mode: ModeSystem}, false},
		{"manual ok", Config{Mode: ModeManual, Scheme: SchemeSOCKS5, Host: "h", Port: 1080}, false},
		{"manual bad scheme", Config{Mode: ModeManual, Scheme: "ftp", Host: "h", Port: 1080}, true},
		{"manual no host", Config{Mode: ModeManual, Scheme: SchemeHTTP, Port: 8080}, true},
		{"manual bad port", Config{Mode: ModeManual, Scheme: SchemeHTTP, Host: "h", Port: 0}, true},
		{"manual port too high", Config{Mode: ModeManual, Scheme: SchemeHTTP, Host: "h", Port: 70000}, true},
	}
	for _, c := range cases {
		err := c.cfg.Validate()
		if (err != nil) != c.wantErr {
			t.Errorf("%s: Validate() err = %v, wantErr %v", c.name, err, c.wantErr)
		}
	}
}

func TestProxyURLCarriesCredentials(t *testing.T) {
	u := Config{Mode: ModeManual, Scheme: SchemeSOCKS5, Host: "proxy.local", Port: 1080, Username: "u", Password: "p"}.proxyURL()
	if u.Scheme != SchemeSOCKS5 || u.Host != "proxy.local:1080" {
		t.Fatalf("unexpected url %s", u)
	}
	if pw, ok := u.User.Password(); !ok || pw != "p" || u.User.Username() != "u" {
		t.Fatalf("credentials not carried: %s", u.Redacted())
	}
}

func TestDialContextNilWhenOff(t *testing.T) {
	if (Config{Mode: ModeOff}).DialContext() != nil {
		t.Fatal("expected nil dialer when off")
	}
	if (Config{Mode: ModeManual, Scheme: SchemeHTTP, Host: "h", Port: 8080}).DialContext() == nil {
		t.Fatal("expected non-nil dialer when enabled")
	}
}

// TestHTTPConnectTunnel spins up a stub HTTP proxy that answers CONNECT and
// verifies the dialer tunnels to the requested target and forwards auth.
func TestHTTPConnectTunnel(t *testing.T) {
	// a plain tcp listener acting as the origin the tunnel should reach.
	origin, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer origin.Close()

	var gotAuth string
	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyLn.Close()

	go func() {
		conn, err := proxyLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		br := bufio.NewReader(conn)
		req, err := http.ReadRequest(br)
		if err != nil {
			return
		}
		gotAuth = req.Header.Get("Proxy-Authorization")
		if req.Method != http.MethodConnect {
			_, _ = conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
			return
		}
		_, _ = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	}()

	_, portStr, _ := net.SplitHostPort(proxyLn.Addr().String())
	port, _ := strconv.Atoi(portStr)
	cfg := Config{Mode: ModeManual, Scheme: SchemeHTTP, Host: "127.0.0.1", Port: port, Username: "alice", Password: "s3cret"}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := cfg.DialContext()(ctx, "tcp", origin.Addr().String())
	if err != nil {
		t.Fatalf("dial through proxy: %v", err)
	}
	_ = conn.Close()

	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Fatalf("expected Basic proxy auth, got %q", gotAuth)
	}
}

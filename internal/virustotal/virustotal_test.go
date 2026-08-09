package virustotal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// statsBody builds a v3 response carrying the given analysis counts.
func statsBody(harmless, malicious, suspicious, undetected, timeout int) string {
	return fmt.Sprintf(
		`{"data":{"attributes":{"last_analysis_stats":{"harmless":%d,"malicious":%d,"suspicious":%d,"undetected":%d,"timeout":%d}}}}`,
		harmless, malicious, suspicious, undetected, timeout,
	)
}

func TestLookupSummarisesAnalysisStats(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantStatus Status
		wantMal    int
		wantTotal  int
	}{
		{
			name:       "no engine flags it",
			status:     http.StatusOK,
			body:       statsBody(60, 0, 0, 8, 2),
			wantStatus: StatusClean,
			wantTotal:  68,
		},
		{
			name:       "one engine calls it malicious",
			status:     http.StatusOK,
			body:       statsBody(55, 3, 0, 12, 0),
			wantStatus: StatusFlagged,
			wantMal:    3,
			wantTotal:  70,
		},
		{
			name:       "only suspicious still counts as flagged",
			status:     http.StatusOK,
			body:       statsBody(60, 0, 1, 9, 0),
			wantStatus: StatusFlagged,
			wantTotal:  70,
		},
		{
			name:       "never analysed",
			status:     http.StatusNotFound,
			body:       `{"error":{"code":"NotFoundError"}}`,
			wantStatus: StatusUnknown,
		},
		{
			name:       "exists but carries no analysis",
			status:     http.StatusOK,
			body:       statsBody(0, 0, 0, 0, 0),
			wantStatus: StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := newTestClient(srv)
			got, err := c.LookupFile(context.Background(), strings.Repeat("a", 64))
			if err != nil {
				t.Fatalf("LookupFile: %v", err)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Malicious != tt.wantMal {
				t.Errorf("malicious = %d, want %d", got.Malicious, tt.wantMal)
			}
			if got.Total != tt.wantTotal {
				t.Errorf("total = %d, want %d", got.Total, tt.wantTotal)
			}
		})
	}
}

func TestLookupMapsErrorStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{"rejected key", http.StatusUnauthorized, ErrUnauthorized},
		{"forbidden key", http.StatusForbidden, ErrUnauthorized},
		{"quota exhausted", http.StatusTooManyRequests, ErrRateLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			_, err := newTestClient(srv).LookupURL(context.Background(), "https://example.com/")
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

// A missing key must short-circuit: nothing may reach the network, since the
// request itself is the disclosure this feature is opt-in about.
func TestLookupWithoutKeyMakesNoRequest(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	c := &Client{http: srv.Client(), key: ""}
	if _, err := c.LookupURL(context.Background(), "https://example.com/"); !errors.Is(err, ErrNoAPIKey) {
		t.Fatalf("err = %v, want ErrNoAPIKey", err)
	}
	if called {
		t.Fatal("a lookup without an api key reached the network")
	}
}

func TestLookupSendsTheKeyAndNoBody(t *testing.T) {
	var gotKey, gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-apikey")
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(statsBody(70, 0, 0, 0, 0)))
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).LookupURL(context.Background(), "https://example.com/a"); err != nil {
		t.Fatalf("LookupURL: %v", err)
	}

	if gotKey != "test-key" {
		t.Errorf("x-apikey = %q, want %q", gotKey, "test-key")
	}
	// a POST would queue the target for analysis, which this package must never do.
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if len(gotBody) != 0 {
		t.Errorf("sent a request body of %d bytes, want none", len(gotBody))
	}
	wantID := base64.RawURLEncoding.EncodeToString([]byte("https://example.com/a"))
	if gotPath != "/urls/"+wantID {
		t.Errorf("path = %q, want %q", gotPath, "/urls/"+wantID)
	}
}

func TestLookupFileRejectsNonDigests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("a malformed digest reached the network")
	}))
	defer srv.Close()

	for _, bad := range []string{"", "deadbeef", strings.Repeat("a", 63), strings.Repeat("a", 65)} {
		if _, err := newTestClient(srv).LookupFile(context.Background(), bad); err == nil {
			t.Errorf("LookupFile(%q) = nil error, want a rejection", bad)
		}
	}
}

func TestURLIDIsUnpaddedBase64URL(t *testing.T) {
	// the padded form is what VirusTotal rejects, so the absence of "=" is the
	// property worth pinning.
	got := URLID("https://example.com/")
	if strings.Contains(got, "=") {
		t.Errorf("URLID = %q, want no padding", got)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(got)
	if err != nil {
		t.Fatalf("decode %q: %v", got, err)
	}
	if string(decoded) != "https://example.com/" {
		t.Errorf("decoded = %q, want the original url", decoded)
	}
	if URLID("   ") != "" {
		t.Error("URLID of blank input should be empty")
	}
}

// newTestClient points the package's base URL at a test server for the
// duration of the test.
func newTestClient(srv *httptest.Server) *Client {
	return &Client{http: srv.Client(), key: "test-key", base: srv.URL}
}

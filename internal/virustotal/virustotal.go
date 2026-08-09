// Package virustotal is a minimal read-only client for the VirusTotal v3 API,
// covering the two lookups Pelton needs: a URL's existing analysis and a file's
// analysis by SHA-256.
//
// It only ever reads. There is no submission path: no URL is queued for
// analysis and no file bytes are ever uploaded, so the only thing that leaves
// the machine is the URL or hash being asked about, and only for a target the
// user (or a setting the user turned on) asked to scan.
package virustotal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// apiBase is the v3 API root. guiBase is the human-facing report site the
// permalinks point at.
const (
	apiBase = "https://www.virustotal.com/api/v3"
	guiBase = "https://www.virustotal.com/gui"
)

// maxBody bounds how much of a response is read, so a malformed or hostile
// reply cannot exhaust memory.
const maxBody = 1 << 20

// Errors callers distinguish. Anything else surfaces as a wrapped error.
var (
	// ErrNoAPIKey means the client was built without a key; no request was made.
	ErrNoAPIKey = errors.New("virustotal: no api key configured")
	// ErrUnauthorized means VirusTotal rejected the key.
	ErrUnauthorized = errors.New("virustotal: api key was rejected")
	// ErrRateLimited means the account's request quota is exhausted for now.
	// The free tier allows a few requests per minute and a few hundred a day.
	ErrRateLimited = errors.New("virustotal: rate limit reached")
)

// Status is the summarised outcome of a lookup.
type Status string

const (
	// StatusClean means VirusTotal has an analysis and no engine flagged it.
	StatusClean Status = "clean"
	// StatusFlagged means at least one engine called it malicious or suspicious.
	StatusFlagged Status = "flagged"
	// StatusUnknown means VirusTotal has never analysed this target. It is not
	// a statement that the target is safe.
	StatusUnknown Status = "unknown"
)

// Verdict is one lookup's result. Malicious, Suspicious and Total are engine
// counts, and are all zero for StatusUnknown. Permalink is the report page.
type Verdict struct {
	Status     Status `json:"status"`
	Malicious  int    `json:"malicious"`
	Suspicious int    `json:"suspicious"`
	Total      int    `json:"total"`
	Permalink  string `json:"permalink"`
}

// Client performs VirusTotal lookups over the caller's http client, which is
// how proxy settings and timeouts are applied.
type Client struct {
	http *http.Client
	key  string
	base string
}

// New returns a client that authenticates with apiKey. A client built with an
// empty key returns ErrNoAPIKey from every lookup rather than calling out.
func New(httpClient *http.Client, apiKey string) *Client {
	return &Client{http: httpClient, key: strings.TrimSpace(apiKey), base: apiBase}
}

// LookupURL returns the existing analysis for raw. A URL VirusTotal has never
// seen comes back as StatusUnknown with a nil error; it is not submitted for
// analysis.
func (c *Client) LookupURL(ctx context.Context, raw string) (Verdict, error) {
	id := URLID(raw)
	if id == "" {
		return Verdict{}, fmt.Errorf("virustotal: empty url")
	}
	return c.lookup(ctx, "urls", id, guiBase+"/url/"+id)
}

// LookupFile returns the existing analysis for a file's SHA-256 digest. Only
// the digest is sent; the file itself never leaves the machine, so a file
// VirusTotal has not seen before stays StatusUnknown forever rather than being
// uploaded.
func (c *Client) LookupFile(ctx context.Context, sha256 string) (Verdict, error) {
	digest := strings.ToLower(strings.TrimSpace(sha256))
	if len(digest) != 64 {
		return Verdict{}, fmt.Errorf("virustotal: %q is not a sha-256 digest", sha256)
	}
	return c.lookup(ctx, "files", digest, guiBase+"/file/"+digest)
}

// URLID is VirusTotal's identifier for a URL: its unpadded base64url encoding.
// The same value addresses both the API resource and the GUI report.
func URLID(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString([]byte(trimmed))
}

// lookup fetches one resource and summarises its last analysis stats.
func (c *Client) lookup(ctx context.Context, kind, id, permalink string) (Verdict, error) {
	if c.key == "" {
		return Verdict{}, ErrNoAPIKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/"+kind+"/"+id, nil)
	if err != nil {
		return Verdict{}, fmt.Errorf("virustotal: build request: %w", err)
	}
	req.Header.Set("x-apikey", c.key)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Verdict{}, fmt.Errorf("virustotal: %s lookup: %w", kind, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		// never analysed, which is a result rather than a failure.
		return Verdict{Status: StatusUnknown, Permalink: permalink}, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return Verdict{}, ErrUnauthorized
	case http.StatusTooManyRequests:
		return Verdict{}, ErrRateLimited
	default:
		return Verdict{}, fmt.Errorf("virustotal: %s lookup: unexpected status %d", kind, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return Verdict{}, fmt.Errorf("virustotal: read response: %w", err)
	}

	var payload struct {
		Data struct {
			Attributes struct {
				LastAnalysisStats struct {
					Harmless   int `json:"harmless"`
					Malicious  int `json:"malicious"`
					Suspicious int `json:"suspicious"`
					Undetected int `json:"undetected"`
					Timeout    int `json:"timeout"`
				} `json:"last_analysis_stats"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Verdict{}, fmt.Errorf("virustotal: decode response: %w", err)
	}

	stats := payload.Data.Attributes.LastAnalysisStats
	verdict := Verdict{
		Malicious:  stats.Malicious,
		Suspicious: stats.Suspicious,
		// engines that timed out reached no verdict, so they are not part of
		// the denominator the ui shows as "n of m engines".
		Total:     stats.Harmless + stats.Malicious + stats.Suspicious + stats.Undetected,
		Permalink: permalink,
	}
	switch {
	case verdict.Total == 0:
		// the resource exists but carries no analysis, which reads the same to
		// a user as never having been scanned.
		verdict.Status = StatusUnknown
	case verdict.Malicious > 0 || verdict.Suspicious > 0:
		verdict.Status = StatusFlagged
	default:
		verdict.Status = StatusClean
	}
	return verdict, nil
}

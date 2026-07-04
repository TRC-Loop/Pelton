// Package credentials stores account secrets in the OS keyring (macOS Keychain,
// Windows Credential Manager, libsecret on linux) via go-keyring. Only the
// non-secret account metadata lives in the sqlite store; passwords and oauth
// tokens live here, referenced by account id, and never touch the database.
package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	keyring "github.com/zalando/go-keyring"
)

// service is the keyring service name all pelton secrets are filed under.
const service = "Pelton"

// maxEntrySize keeps every keyring write under Windows Credential Manager's
// hard 2560-byte-per-entry cap (with margin), splitting anything larger
// across multiple entries. OAuth secrets with long access/refresh tokens can
// exceed that on their own.
const maxEntrySize = 2000

// chunkMarker prefixes the main entry's value when the secret was split
// across chunkMarker+N and the numbered chunk entries below it.
const chunkMarker = "pelton-chunked:v1:"

// ErrNotFound is returned when no secret is stored for an account.
var ErrNotFound = errors.New("credentials: not found")

// Method is how an account authenticates.
type Method string

const (
	// MethodPassword is plain username/password (or app-specific password).
	MethodPassword Method = "password"
	// MethodOAuth is XOAUTH2 with a refresh token (gmail, outlook).
	MethodOAuth Method = "oauth"
)

// Secret is everything needed to authenticate one account, kept out of the db.
// For oauth, AccessToken/Expiry are a cache the oauth package refreshes from
// RefreshToken; Provider and ClientID identify how to refresh.
type Secret struct {
	Method       Method    `json:"method"`
	Password     string    `json:"password,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	ClientID     string    `json:"clientId,omitempty"`
	ClientSecret string    `json:"clientSecret,omitempty"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	AccessToken  string    `json:"accessToken,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// Store writes the secret for an account, replacing any existing one. The new
// value (and, if chunked, every new chunk) is written before any stale chunk
// from a previous, larger secret is removed, so a failed write never leaves
// the account's secret unreadable.
func Store(accountID int64, s Secret) error {
	encoded, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("credentials: encode secret: %w", err)
	}
	oldChunks := existingChunkCount(accountID)

	if len(encoded) <= maxEntrySize {
		if err := keyring.Set(service, key(accountID), string(encoded)); err != nil {
			return fmt.Errorf("credentials: store for account %d: %w", accountID, err)
		}
		deleteChunkRange(accountID, 0, oldChunks)
		return nil
	}

	chunks := chunkBytes(encoded, maxEntrySize)
	for i, part := range chunks {
		if err := keyring.Set(service, chunkKey(accountID, i), string(part)); err != nil {
			return fmt.Errorf("credentials: store chunk %d for account %d: %w", i, accountID, err)
		}
	}
	if err := keyring.Set(service, key(accountID), chunkMarker+strconv.Itoa(len(chunks))); err != nil {
		return fmt.Errorf("credentials: store for account %d: %w", accountID, err)
	}
	deleteChunkRange(accountID, len(chunks), oldChunks)
	return nil
}

// Load reads the secret for an account, or ErrNotFound.
func Load(accountID int64) (Secret, error) {
	raw, err := keyring.Get(service, key(accountID))
	if errors.Is(err, keyring.ErrNotFound) {
		return Secret{}, ErrNotFound
	}
	if err != nil {
		return Secret{}, fmt.Errorf("credentials: load for account %d: %w", accountID, err)
	}
	if n, ok := chunkCount(raw); ok {
		var sb strings.Builder
		for i := range n {
			part, err := keyring.Get(service, chunkKey(accountID, i))
			if err != nil {
				return Secret{}, fmt.Errorf("credentials: load chunk %d for account %d: %w", i, accountID, err)
			}
			sb.WriteString(part)
		}
		raw = sb.String()
	}
	var s Secret
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return Secret{}, fmt.Errorf("credentials: decode secret for account %d: %w", accountID, err)
	}
	return s, nil
}

// Delete removes the secret for an account. A missing entry is not an error so
// account deletion is idempotent.
func Delete(accountID int64) error {
	deleteChunkRange(accountID, 0, existingChunkCount(accountID))
	err := keyring.Delete(service, key(accountID))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("credentials: delete for account %d: %w", accountID, err)
	}
	return nil
}

// key is the per-account keyring entry name.
func key(accountID int64) string {
	return strconv.FormatInt(accountID, 10)
}

// chunkKey is the entry name for chunk i of a split secret.
func chunkKey(accountID int64, i int) string {
	return key(accountID) + "." + strconv.Itoa(i)
}

// chunkCount reports whether raw is a chunk marker and, if so, how many
// chunk entries to read.
func chunkCount(raw string) (int, bool) {
	if !strings.HasPrefix(raw, chunkMarker) {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(raw, chunkMarker))
	if err != nil {
		return 0, false
	}
	return n, true
}

// chunkBytes splits data into pieces of at most size bytes.
func chunkBytes(data []byte, size int) [][]byte {
	var out [][]byte
	for len(data) > size {
		out = append(out, data[:size])
		data = data[size:]
	}
	return append(out, data)
}

// existingChunkCount returns how many chunk entries accountID's current
// secret is split across, or 0 if it isn't chunked (or doesn't exist).
func existingChunkCount(accountID int64) int {
	raw, err := keyring.Get(service, key(accountID))
	if err != nil {
		return 0
	}
	n, ok := chunkCount(raw)
	if !ok {
		return 0
	}
	return n
}

// deleteChunkRange removes chunk entries [from, to) for accountID.
func deleteChunkRange(accountID int64, from, to int) {
	for i := from; i < to; i++ {
		_ = keyring.Delete(service, chunkKey(accountID, i))
	}
}

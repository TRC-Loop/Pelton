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
	"time"

	keyring "github.com/zalando/go-keyring"
)

// service is the keyring service name all pelton secrets are filed under.
const service = "Pelton"

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

// Store writes the secret for an account, replacing any existing one.
func Store(accountID int64, s Secret) error {
	encoded, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("credentials: encode secret: %w", err)
	}
	if err := keyring.Set(service, key(accountID), string(encoded)); err != nil {
		return fmt.Errorf("credentials: store for account %d: %w", accountID, err)
	}
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
	var s Secret
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return Secret{}, fmt.Errorf("credentials: decode secret for account %d: %w", accountID, err)
	}
	return s, nil
}

// Delete removes the secret for an account. A missing entry is not an error so
// account deletion is idempotent.
func Delete(accountID int64) error {
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

package desktop

import (
	"errors"
	"os"

	"github.com/TRC-Loop/Pelton/internal/credentials"
	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/oauth"
	psmtp "github.com/TRC-Loop/Pelton/internal/smtp"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// errNoCredentials means an account has no usable secret in the keyring and no
// matching environment fallback, so it cannot be synced or sent from.
var errNoCredentials = errors.New("pelton: no credentials for account")

// loginName is the name to authenticate with: the account's explicit username
// when set, otherwise the email address. Every account created before separate
// usernames existed has an empty Username and so keeps logging in by email.
func loginName(account storage.Account) string {
	if account.Username != "" {
		return account.Username
	}
	return account.Email
}

// resolveIMAP builds an imap config for an account from its keyring secret,
// refreshing an oauth access token if needed. It falls back to environment
// credentials for the legacy cli-created account (matched by email) so existing
// setups keep working before they are re-added through the wizard.
func (a *App) resolveIMAP(account storage.Account) (pimap.Config, error) {
	cfg := pimap.Config{
		Host:     account.IMAPHost,
		Port:     account.IMAPPort,
		Username: loginName(account),
		Dial:     a.proxyDial(),
	}

	secret, err := credentials.Load(account.ID)
	if errors.Is(err, credentials.ErrNotFound) {
		return a.imapFromEnv(cfg)
	}
	if err != nil {
		return pimap.Config{}, err
	}

	switch secret.Method {
	case credentials.MethodOAuth:
		token, err := a.freshAccessToken(account.ID, secret)
		if err != nil {
			return pimap.Config{}, err
		}
		cfg.OAuth2Token = token
	default:
		cfg.Password = secret.Password
	}
	return cfg, nil
}

// resolveSMTP builds an smtp config for an account, mirroring resolveIMAP. With
// an oauth token set the smtp layer auto-selects XOAUTH2.
func (a *App) resolveSMTP(account storage.Account) (psmtp.Config, error) {
	cfg := psmtp.Config{
		Host:     account.SMTPHost,
		Port:     account.SMTPPort,
		Username: loginName(account),
		Dial:     a.proxyDial(),
	}

	secret, err := credentials.Load(account.ID)
	if errors.Is(err, credentials.ErrNotFound) {
		return a.smtpFromEnv(cfg)
	}
	if err != nil {
		return psmtp.Config{}, err
	}

	switch secret.Method {
	case credentials.MethodOAuth:
		token, err := a.freshAccessToken(account.ID, secret)
		if err != nil {
			return psmtp.Config{}, err
		}
		cfg.OAuth2Token = token
	default:
		cfg.Password = secret.Password
	}
	return cfg, nil
}

// freshAccessToken returns a valid oauth access token for an account, refreshing
// from the stored refresh token and persisting any rotated refresh token.
func (a *App) freshAccessToken(accountID int64, secret credentials.Secret) (string, error) {
	token, err := oauth.FreshToken(a.ctx, secret.Provider, secret.ClientID, secret.ClientSecret, secret.RefreshToken)
	if err != nil {
		return "", err
	}
	// persist a rotated refresh token and the cached access token so the next
	// call can reuse it until expiry.
	updated := secret
	updated.AccessToken = token.AccessToken
	updated.Expiry = token.Expiry
	if token.RefreshToken != "" {
		updated.RefreshToken = token.RefreshToken
	}
	if err := credentials.Store(accountID, updated); err != nil {
		a.log.Error("persist refreshed token", "account", accountID, "err", err)
	}
	return token.AccessToken, nil
}

// imapFromEnv applies environment credentials to a config when they match the
// account, otherwise reports that none are available.
func (a *App) imapFromEnv(cfg pimap.Config) (pimap.Config, error) {
	if os.Getenv("IMAP_USER") != cfg.Username || os.Getenv("IMAP_PASSWORD") == "" {
		return pimap.Config{}, errNoCredentials
	}
	cfg.Password = os.Getenv("IMAP_PASSWORD")
	if cfg.Host == "" {
		cfg.Host = os.Getenv("IMAP_HOST")
	}
	cfg.InsecureSkipVerify = os.Getenv("IMAP_INSECURE") == "1"
	return cfg, nil
}

// smtpFromEnv is the smtp counterpart of imapFromEnv.
func (a *App) smtpFromEnv(cfg psmtp.Config) (psmtp.Config, error) {
	if os.Getenv("SMTP_USER") != cfg.Username || os.Getenv("SMTP_PASSWORD") == "" {
		return psmtp.Config{}, errNoCredentials
	}
	cfg.Password = os.Getenv("SMTP_PASSWORD")
	if cfg.Host == "" {
		cfg.Host = os.Getenv("SMTP_HOST")
	}
	return cfg, nil
}

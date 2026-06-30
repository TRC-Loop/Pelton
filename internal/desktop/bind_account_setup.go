package desktop

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TRC-Loop/Pelton/internal/autoconfig"
	"github.com/TRC-Loop/Pelton/internal/credentials"
	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/oauth"
	"github.com/TRC-Loop/Pelton/internal/storage"
	goimap "github.com/emersion/go-imap/v2"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// oauthFlowTimeout bounds how long the interactive consent flow may take.
const oauthFlowTimeout = 5 * time.Minute

// DiscoveredDTO is the autodiscovery result for the wizard.
type DiscoveredDTO struct {
	IMAPHost string `json:"imapHost"`
	IMAPPort int    `json:"imapPort"`
	SMTPHost string `json:"smtpHost"`
	SMTPPort int    `json:"smtpPort"`
	OAuth    bool   `json:"oauth"`
	Source   string `json:"source"`
}

// DiscoverConfig resolves likely imap/smtp settings for an email address using
// autoconfig (ISPDB, the domain's well-known/autoconfig, then a guess). The
// wizard pre-fills the form with this; the user can still edit before testing.
func (a *App) DiscoverConfig(email string) (DiscoveredDTO, error) {
	d, err := autoconfig.Discover(a.ctx, email)
	if err != nil {
		return DiscoveredDTO{}, err
	}
	return DiscoveredDTO{
		IMAPHost: d.IMAPHost,
		IMAPPort: d.IMAPPort,
		SMTPHost: d.SMTPHost,
		SMTPPort: d.SMTPPort,
		OAuth:    d.OAuth,
		Source:   d.Source,
	}, nil
}

// ListOAuthProviders returns the supported oauth provider keys and labels so the
// wizard knows which providers can use the sign-in flow.
func (a *App) ListOAuthProviders() (map[string]string, error) {
	return oauth.Providers(), nil
}

// TestConnectionRequest carries the settings to verify before saving (password
// auth). OAuth is verified by the sign-in flow itself, so it is not tested here.
type TestConnectionRequest struct {
	Email    string `json:"email"`
	IMAPHost string `json:"imapHost"`
	IMAPPort int    `json:"imapPort"`
	Password string `json:"password"`
}

// TestConnection verifies imap credentials by connecting and logging in, so the
// wizard can confirm before creating the account. It returns nil on success.
func (a *App) TestConnection(req TestConnectionRequest) error {
	client, err := pimap.Connect(pimap.Config{
		Host:     req.IMAPHost,
		Port:     req.IMAPPort,
		Username: req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	return client.Logout()
}

// AddAccountRequest is the metadata the wizard collected. For password auth
// Password is set; for oauth Provider and ClientID are set and the flow runs.
type AddAccountRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	IMAPHost    string `json:"imapHost"`
	IMAPPort    int    `json:"imapPort"`
	SMTPHost    string `json:"smtpHost"`
	SMTPPort    int    `json:"smtpPort"`
	// auth
	Password string `json:"password"`
	Provider string `json:"provider"`
	ClientID string `json:"clientId"`
	// ClientSecret is optional and only used for oauth providers registered as
	// confidential clients (some Microsoft Entra app registrations). Empty keeps
	// the default public-client PKCE flow.
	ClientSecret string `json:"clientSecret"`
}

// AddPasswordAccount creates a password-authenticated account: it stores the
// metadata, files the password in the keyring, discovers the folder tree and
// runs an initial sync.
func (a *App) AddPasswordAccount(req AddAccountRequest) (AccountDTO, error) {
	if err := a.ready(); err != nil {
		return AccountDTO{}, err
	}
	secret := credentials.Secret{Method: credentials.MethodPassword, Password: req.Password}
	return a.createAccount(req, secret)
}

// AddOAuthAccount creates an oauth account: it runs the interactive PKCE flow
// (opening the system browser), stores the resulting refresh token in the
// keyring, then discovers folders and syncs. ClientID is the user's own
// registered desktop client id.
func (a *App) AddOAuthAccount(req AddAccountRequest) (AccountDTO, error) {
	if err := a.ready(); err != nil {
		return AccountDTO{}, err
	}

	ctx, cancel := context.WithTimeout(a.ctx, oauthFlowTimeout)
	defer cancel()

	token, err := oauth.Authorize(ctx, req.Provider, req.ClientID, req.ClientSecret, req.Email, func(url string) {
		wailsruntime.BrowserOpenURL(a.ctx, url)
	})
	if err != nil {
		return AccountDTO{}, err
	}
	if token.RefreshToken == "" {
		return AccountDTO{}, fmt.Errorf("pelton: provider returned no refresh token; re-consent may be required")
	}

	secret := credentials.Secret{
		Method:       credentials.MethodOAuth,
		Provider:     req.Provider,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		RefreshToken: token.RefreshToken,
		AccessToken:  token.AccessToken,
		Expiry:       token.Expiry,
	}
	return a.createAccount(req, secret)
}

// createAccount is the shared path for both auth methods: persist metadata, store
// the secret, discover folders, sync, and start idling. On any failure after the
// row is created it rolls the account back so a half-created account is not left.
func (a *App) createAccount(req AddAccountRequest, secret credentials.Secret) (AccountDTO, error) {
	account := &storage.Account{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		IMAPHost:    req.IMAPHost,
		IMAPPort:    req.IMAPPort,
		SMTPHost:    req.SMTPHost,
		SMTPPort:    req.SMTPPort,
	}
	id, err := a.store.CreateAccount(a.ctx, account)
	if err != nil {
		return AccountDTO{}, err
	}

	if err := credentials.Store(id, secret); err != nil {
		_ = a.store.DeleteAccount(a.ctx, id)
		return AccountDTO{}, err
	}

	if err := a.discoverFolders(*account); err != nil {
		// keep the account; folders can be (re)discovered on next sync. surface
		// the error so the wizard can warn, but the account exists.
		a.log.Error("discover folders", "account", account.Email, "err", err)
	}

	// initial sync and idle in the background so the wizard returns promptly.
	go func() {
		if err := a.syncAccount(*account); err != nil {
			a.log.Error("initial sync after add", "account", account.Email, "err", err)
		}
		go a.idleLoop(*account)
	}()

	return toAccountDTO(*account), nil
}

// discoverFolders lists the server's mailboxes and creates the storage folder
// rows, preserving the hierarchy via the per-server delimiter so the sidebar
// tree matches the server.
func (a *App) discoverFolders(account storage.Account) error {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return err
	}
	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	defer client.Logout()

	folders, err := client.ListFolders()
	if err != nil {
		return err
	}
	return a.createFolderTree(account.ID, folders)
}

// createFolderTree inserts folders parent-first so each child can resolve its
// parent id from the path above it, splitting on the server's delimiter.
func (a *App) createFolderTree(accountID int64, folders []pimap.Folder) error {
	// shallowest first so parents exist before their children.
	sort.SliceStable(folders, func(i, j int) bool {
		return depth(folders[i]) < depth(folders[j])
	})

	byPath := make(map[string]int64)
	for _, f := range folders {
		if f.HasAttr(goimap.MailboxAttrNonExistent) {
			continue
		}
		delim := delimString(f.Delimiter)
		name := f.Name
		var parentID *int64
		if delim != "" {
			if idx := strings.LastIndex(f.Name, delim); idx >= 0 {
				name = f.Name[idx+len(delim):]
				if pid, ok := byPath[f.Name[:idx]]; ok {
					parentID = &pid
				}
			}
		}

		row := &storage.Folder{
			AccountID:  accountID,
			Name:       name,
			IMAPPath:   f.Name,
			Delimiter:  delim,
			ParentID:   parentID,
			Attributes: attrsToStrings(f.Attrs),
		}
		id, err := a.store.CreateFolder(a.ctx, row)
		if err != nil {
			return err
		}
		byPath[f.Name] = id
	}
	return nil
}

// depth counts how many delimiter segments a folder path has, for sort order.
func depth(f pimap.Folder) int {
	d := delimString(f.Delimiter)
	if d == "" {
		return 0
	}
	return strings.Count(f.Name, d)
}

// delimString renders the rune delimiter as a string, empty for a flat server.
func delimString(r rune) string {
	if r == 0 {
		return ""
	}
	return string(r)
}

// attrsToStrings converts imap mailbox attributes to plain strings for storage.
func attrsToStrings(attrs []goimap.MailboxAttr) []string {
	out := make([]string, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, string(a))
	}
	return out
}

package desktop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/TRC-Loop/Pelton/internal/credentials"
	"github.com/TRC-Loop/Pelton/internal/proxy"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// proxyTestTarget is the host:port the connection test dials through the proxy.
// It is a well-known always-on endpoint; only reaching it (not any response)
// matters, so nothing is sent over the tunnel.
const proxyTestTarget = "one.one.one.one:443"

// ProxyConfigDTO is the proxy preference exchanged with the ui. Password is
// write-only from the ui's side: HasPassword reports whether one is stored so
// the field can show a placeholder without ever shipping the secret back out.
type ProxyConfigDTO struct {
	Mode        string `json:"mode"`
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	HasPassword bool   `json:"hasPassword"`
}

// loadProxy reads the saved proxy preference (settings json + keyring password)
// into the cache. Called once at startup; a missing or malformed setting leaves
// the zero value, which means "off" (direct connections).
func (a *App) loadProxy() {
	cfg := proxy.Config{Mode: proxy.ModeOff}
	raw, err := a.store.Get(a.ctx, storage.SettingProxy)
	if err == nil && raw != "" {
		if uErr := json.Unmarshal([]byte(raw), &cfg); uErr != nil {
			a.log.Error("parse proxy setting", "err", uErr)
			cfg = proxy.Config{Mode: proxy.ModeOff}
		}
	} else if err != nil && !errors.Is(err, storage.ErrSettingNotFound) {
		a.log.Error("read proxy setting", "err", err)
	}
	if cfg.Mode == proxy.ModeManual {
		if pw, pErr := credentials.LoadProxyPassword(); pErr != nil {
			a.log.Error("load proxy password", "err", pErr)
		} else {
			cfg.Password = pw
		}
	}
	a.proxyMu.Lock()
	a.proxyCfg = cfg
	a.proxyMu.Unlock()
}

// currentProxy returns the cached proxy preference.
func (a *App) currentProxy() proxy.Config {
	a.proxyMu.RLock()
	defer a.proxyMu.RUnlock()
	return a.proxyCfg
}

// proxyDial returns the tcp dial hook for the mail clients, or nil when no proxy
// is configured (nil leaves the clients on their direct-dial path).
func (a *App) proxyDial() func(ctx context.Context, network, addr string) (net.Conn, error) {
	return a.currentProxy().DialContext()
}

// httpClient returns an http client that honours the proxy preference, for the
// app's outbound web calls.
func (a *App) httpClient(timeout time.Duration) *http.Client {
	return a.currentProxy().HTTPClient(timeout)
}

// GetProxyConfig returns the current proxy preference for the settings ui,
// without the password itself.
func (a *App) GetProxyConfig() (ProxyConfigDTO, error) {
	if err := a.ready(); err != nil {
		return ProxyConfigDTO{}, err
	}
	cfg := a.currentProxy()
	return ProxyConfigDTO{
		Mode:        orOff(cfg.Mode),
		Scheme:      orDefaultScheme(cfg.Scheme),
		Host:        cfg.Host,
		Port:        cfg.Port,
		Username:    cfg.Username,
		HasPassword: cfg.Password != "",
	}, nil
}

// SetProxyConfig validates, persists and applies a new proxy preference. An
// empty Password with HasPassword=true keeps the stored one (the ui sends no
// secret unless the user typed a new value); Mode "off"/"system" clears it.
func (a *App) SetProxyConfig(dto ProxyConfigDTO) error {
	if err := a.ready(); err != nil {
		return err
	}
	cfg := proxy.Config{
		Mode:     dto.Mode,
		Scheme:   dto.Scheme,
		Host:     dto.Host,
		Port:     dto.Port,
		Username: dto.Username,
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	// resolve the password to persist: a new one if typed, otherwise the stored
	// one when the ui reported it kept the placeholder.
	password := dto.Password
	if password == "" && dto.HasPassword {
		if pw, err := credentials.LoadProxyPassword(); err == nil {
			password = pw
		}
	}
	if cfg.Mode != proxy.ModeManual {
		password = ""
	}

	encoded, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("proxy: encode setting: %w", err)
	}
	if err := a.store.Set(a.ctx, storage.SettingProxy, string(encoded)); err != nil {
		return err
	}
	if err := credentials.StoreProxyPassword(password); err != nil {
		return err
	}

	cfg.Password = password
	a.proxyMu.Lock()
	a.proxyCfg = cfg
	a.proxyMu.Unlock()
	return nil
}

// TestProxy dials a well-known endpoint through the given settings so the ui can
// confirm a proxy works before relying on it. It uses the typed password, or
// the stored one when the field was left on its placeholder.
func (a *App) TestProxy(dto ProxyConfigDTO) error {
	if err := a.ready(); err != nil {
		return err
	}
	cfg := proxy.Config{
		Mode:     dto.Mode,
		Scheme:   dto.Scheme,
		Host:     dto.Host,
		Port:     dto.Port,
		Username: dto.Username,
		Password: dto.Password,
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Mode == proxy.ModeManual && cfg.Password == "" && dto.HasPassword {
		if pw, err := credentials.LoadProxyPassword(); err == nil {
			cfg.Password = pw
		}
	}
	dial := cfg.DialContext()
	if dial == nil {
		// "off": nothing to test, a direct connection always "works" here.
		return nil
	}
	ctx, cancel := context.WithTimeout(a.ctx, 20*time.Second)
	defer cancel()
	conn, err := dial(ctx, "tcp", proxyTestTarget)
	if err != nil {
		return err
	}
	return conn.Close()
}

func orOff(mode string) string {
	if mode == "" {
		return proxy.ModeOff
	}
	return mode
}

func orDefaultScheme(scheme string) string {
	if scheme == "" {
		return proxy.SchemeSOCKS5
	}
	return scheme
}

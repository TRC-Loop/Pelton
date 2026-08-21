// Package mcpserver exposes Pelton's cached mail over the Model Context Protocol
// so external AI agents can browse, read and search mail. It is read-only: no
// tool here sends, moves, flags or deletes anything (write actions are tracked
// separately). The server is off by default in the app; the desktop layer owns
// the enable toggle, port and bearer token, and drives Start/Stop.
//
// Transport is streamable HTTP bound to loopback only, guarded by a bearer
// token, so another process or a web page on the machine cannot silently drive
// a user's mailbox. The protocol plumbing lives here behind the narrow Mailbox
// interface; the desktop package adapts the store and search index to it.
package mcpserver

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverName is the MCP implementation name reported to clients.
const serverName = "pelton"

// Config is the runtime configuration for one Start. Addr must be a loopback
// host:port (the desktop layer builds it from 127.0.0.1 and the port setting);
// Token is the bearer token every request must present. Version is Pelton's
// build version, surfaced to clients in the initialize handshake.
type Config struct {
	Addr    string
	Token   string
	Version string
	// Permissions says which write tools this run allows. The zero value allows
	// none, so the server is read-only unless something says otherwise.
	Permissions Permissions
}

// Server owns the HTTP listener and the MCP server. It is safe for concurrent
// Start/Stop calls; the desktop layer calls Stop then Start to apply a changed
// port or a regenerated token.
type Server struct {
	mb Mailbox
	// w is nil on a read-only server, and then no write tool is registered at
	// all. The permission check is the second line, not the only one.
	w   Writer
	log *slog.Logger

	mu      sync.Mutex
	httpSrv *http.Server
	ln      net.Listener
}

// New creates a read-only Server backed by mb. It does not listen until Start
// is called.
func New(mb Mailbox, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{mb: mb, log: log}
}

// NewWithWriter creates a Server that can also act on mail, subject to the
// permissions given to Start.
func NewWithWriter(mb Mailbox, w Writer, log *slog.Logger) *Server {
	s := New(mb, log)
	s.w = w
	return s
}

// Running reports whether the server is currently listening.
func (s *Server) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.httpSrv != nil
}

// Start begins listening on cfg.Addr. It fails if the token is empty (an
// unauthenticated local mail endpoint is never acceptable), if the address is
// not loopback, or if the port is already in use. Calling Start while already
// running returns an error; call Stop first.
func (s *Server) Start(cfg Config) error {
	if cfg.Token == "" {
		return fmt.Errorf("mcpserver: refusing to start without a bearer token")
	}
	if err := requireLoopback(cfg.Addr); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpSrv != nil {
		return fmt.Errorf("mcpserver: already running")
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("mcpserver: listen on %s: %w", cfg.Addr, err)
	}

	mcpSrv := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: cfg.Version,
		Title:   "Pelton",
	}, &mcp.ServerOptions{Instructions: serverInstructions(cfg.Permissions)})
	registerTools(mcpSrv, s.mb)
	registerWriteTools(mcpSrv, s.w, cfg.Permissions)

	// Stateless: every request is self-contained. The read-only tools need no
	// per-session state, and this makes the server immune to losing a client's
	// session across a restart (dev hot-reload) - the failure that otherwise
	// surfaces to the client as "session not found".
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpSrv
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	httpSrv := &http.Server{
		Handler:           withBearerAuth(cfg.Token, handler),
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.httpSrv = httpSrv
	s.ln = ln

	go func() {
		if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.log.Error("mcp server serve", "err", err)
		}
	}()
	s.log.Info("mcp server started", "addr", cfg.Addr)
	return nil
}

// Stop shuts the listener down gracefully, bounded by ctx. It is a no-op when
// the server is not running.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	httpSrv := s.httpSrv
	s.httpSrv = nil
	s.ln = nil
	s.mu.Unlock()

	if httpSrv == nil {
		return nil
	}
	s.log.Info("mcp server stopping")
	return httpSrv.Shutdown(ctx)
}

// withBearerAuth wraps next so every request must carry an Authorization:
// Bearer <token> header matching token. The comparison is constant-time.
// A missing or wrong token gets 401 before the request ever reaches the MCP
// handler, so an unauthorized caller cannot even enumerate the tools.
func withBearerAuth(token string, next http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireLoopback reports an error unless addr's host resolves to loopback
// only. It is a defense-in-depth guard: the desktop layer always builds a
// 127.0.0.1 address, and this makes binding to a routable interface impossible
// even if that ever changes.
func requireLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("mcpserver: bad address %q: %w", addr, err)
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("mcpserver: address %q is not loopback", addr)
	}
	return nil
}

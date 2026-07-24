package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeMailbox is a deterministic Mailbox for the tool round-trip tests.
type fakeMailbox struct {
	accounts []Account
	folders  []Folder
	messages []MessageSummary
	message  *Message
	searched SearchParams
}

func (f *fakeMailbox) ListAccounts(context.Context) ([]Account, error) { return f.accounts, nil }
func (f *fakeMailbox) ListFolders(context.Context, int64) ([]Folder, error) {
	return f.folders, nil
}
func (f *fakeMailbox) ListMessages(_ context.Context, _ int64, limit int) ([]MessageSummary, error) {
	if limit < len(f.messages) {
		return f.messages[:limit], nil
	}
	return f.messages, nil
}
func (f *fakeMailbox) GetMessage(context.Context, int64) (*Message, error) { return f.message, nil }
func (f *fakeMailbox) Search(_ context.Context, p SearchParams) ([]MessageSummary, error) {
	f.searched = p
	return f.messages, nil
}

// connect wires an in-memory client to a server carrying the read-only tools.
func connect(t *testing.T, mb Mailbox) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	srv := mcp.NewServer(&mcp.Implementation{Name: "pelton", Version: "test"}, nil)
	registerTools(srv, mb)

	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	cs, err := c.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func callText(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("call %s returned tool error: %v", name, res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatalf("call %s returned no content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("call %s: content[0] is %T, want TextContent", name, res.Content[0])
	}
	return tc.Text
}

func TestToolsExposeReadData(t *testing.T) {
	mb := &fakeMailbox{
		accounts: []Account{{ID: 1, Email: "a@example.com"}},
		folders:  []Folder{{ID: 7, AccountID: 1, Name: "INBOX", Path: "INBOX"}},
		messages: []MessageSummary{
			{ID: 100, Subject: "Hello", FromAddress: "s@example.com", Unread: true},
			{ID: 101, Subject: "World", FromAddress: "t@example.com"},
		},
		message: &Message{
			MessageSummary: MessageSummary{ID: 100, Subject: "Hello"},
			MessageID:      "<abc@example.com>",
			SizeBytes:      2048,
			BodyText:       "the body",
			BodyHTML:       "<p>the body</p>",
			Attachments:    []Attachment{{Filename: "a.pdf", SizeBytes: 12}},
		},
	}
	cs := connect(t, mb)

	if got := callText(t, cs, "list_accounts", map[string]any{}); !strings.Contains(got, "a@example.com") {
		t.Errorf("list_accounts missing email: %s", got)
	}
	if got := callText(t, cs, "list_folders", map[string]any{"account_id": 1}); !strings.Contains(got, "INBOX") {
		t.Errorf("list_folders missing folder: %s", got)
	}

	got := callText(t, cs, "list_messages", map[string]any{"folder_id": 7, "limit": 1})
	if !strings.Contains(got, "Hello") || strings.Contains(got, "World") {
		t.Errorf("list_messages did not honor limit=1: %s", got)
	}

	// the JSON text content HTML-escapes angle brackets, so assert on
	// escape-safe substrings; the structured-content test covers exact values.
	got = callText(t, cs, "get_message", map[string]any{"id": 100})
	for _, want := range []string{"the body", "body_html", "a.pdf", "abc@example.com", "2048"} {
		if !strings.Contains(got, want) {
			t.Errorf("get_message missing %q: %s", want, got)
		}
	}

	if got := callText(t, cs, "search_messages", map[string]any{"query": "hello", "from": "bob"}); !strings.Contains(got, "Hello") {
		t.Errorf("search_messages missing hit: %s", got)
	}
	if mb.searched.Query != "hello" || mb.searched.From != "bob" {
		t.Errorf("search params not forwarded: %+v", mb.searched)
	}
}

// TestGetMessageResultShape confirms the result is a single text block of clean
// JSON: no structured content (so no output-schema validation on the client) and
// no HTML escaping of the body, so it stays readable.
func TestGetMessageResultShape(t *testing.T) {
	mb := &fakeMailbox{message: &Message{
		MessageSummary: MessageSummary{ID: 5, Subject: "Subj"},
		BodyText:       "body",
		BodyHTML:       "<p>hi</p>",
	}}
	cs := connect(t, mb)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_message", Arguments: map[string]any{"id": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if res.StructuredContent != nil {
		t.Errorf("expected no structured content, got %v", res.StructuredContent)
	}
	if len(res.Content) != 1 {
		t.Fatalf("expected one content block, got %d", len(res.Content))
	}
	text := res.Content[0].(*mcp.TextContent).Text
	// the body must be parseable back and its html not escaped.
	if !strings.Contains(text, "<p>hi</p>") {
		t.Errorf("html body escaped or missing: %s", text)
	}
	var got Message
	if err := json.Unmarshal([]byte(text), &got); err != nil {
		t.Fatalf("result text is not valid json: %v", err)
	}
	if got.BodyText != "body" || got.Subject != "Subj" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestBearerAuth(t *testing.T) {
	const token = "secret-token"
	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true })
	h := withBearerAuth(token, next)

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantReach  bool
	}{
		{"no header", "", http.StatusUnauthorized, false},
		{"wrong token", "Bearer nope", http.StatusUnauthorized, false},
		{"missing prefix", token, http.StatusUnauthorized, false},
		{"correct", "Bearer " + token, http.StatusOK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reached = false
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if reached != tt.wantReach {
				t.Errorf("handler reached = %v, want %v", reached, tt.wantReach)
			}
		})
	}
}

func TestStartRequiresToken(t *testing.T) {
	s := New(&fakeMailbox{}, nil)
	if err := s.Start(Config{Addr: "127.0.0.1:0", Token: ""}); err == nil {
		t.Fatal("expected error starting without a token")
	}
}

func TestRequireLoopback(t *testing.T) {
	ok := []string{"127.0.0.1:8765", "localhost:1", "[::1]:80"}
	for _, a := range ok {
		if err := requireLoopback(a); err != nil {
			t.Errorf("requireLoopback(%q) = %v, want nil", a, err)
		}
	}
	bad := []string{"0.0.0.0:8765", "192.168.1.5:80", "example.com:80", "nonsense"}
	for _, a := range bad {
		if err := requireLoopback(a); err == nil {
			t.Errorf("requireLoopback(%q) = nil, want error", a)
		}
	}
}

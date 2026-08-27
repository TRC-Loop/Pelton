package mcpserver

import (
	"context"
	"errors"
	"strings"
	"syscall"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeWriter records what it was asked to do, so a test can assert the tool
// reached the mailbox rather than only that it returned no error.
type fakeWriter struct {
	markedID   int64
	markedRead bool
	movedTo    int64
	archived   int64
	flagged    int64
	color      int
	deleted    int64
	queued     OutgoingMessage
	queuedID   int64
}

func (f *fakeWriter) MarkRead(_ context.Context, id int64, read bool) error {
	f.markedID, f.markedRead = id, read
	return nil
}
func (f *fakeWriter) Move(_ context.Context, id, folderID int64) error {
	f.movedTo = folderID
	return nil
}
func (f *fakeWriter) Archive(_ context.Context, id int64) error      { f.archived = id; return nil }
func (f *fakeWriter) Flag(_ context.Context, id int64, _ bool) error { f.flagged = id; return nil }
func (f *fakeWriter) SetFlagColor(_ context.Context, _ int64, c int) error {
	f.color = c
	return nil
}
func (f *fakeWriter) Delete(_ context.Context, id int64) error { f.deleted = id; return nil }
func (f *fakeWriter) QueueSend(_ context.Context, msg OutgoingMessage) (int64, error) {
	f.queued = msg
	f.queuedID = 77
	return 77, nil
}

// connectWrite wires a client to a server carrying both tool sets.
func connectWrite(t *testing.T, w Writer, perms Permissions) *mcp.ClientSession {
	t.Helper()
	return connectLive(t, w, func() Permissions { return perms })
}

// connectLive wires a client to a server whose permissions can change under it,
// which is what happens when the user flips a switch mid-session.
func connectLive(t *testing.T, w Writer, perms func() Permissions) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	srv := mcp.NewServer(&mcp.Implementation{Name: "pelton", Version: "test"}, nil)
	registerTools(srv, &fakeMailbox{})
	registerWriteTools(srv, w, perms)

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

// allPermissions turns every write tool on, for the tests about behaviour
// rather than about gating.
func allPermissions() Permissions {
	p := Permissions{}
	for _, tool := range WriteTools {
		p[tool] = true
	}
	return p
}

// A server built without a Writer must not carry a write tool at all. The
// permission check is the second line of defence, not the only one.
func TestNoWriterMeansNoWriteTools(t *testing.T) {
	cs := connectWrite(t, nil, allPermissions())
	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		if _, isWrite := ToolGroup[tool.Name]; isWrite {
			t.Errorf("write tool %s registered on a server with no writer", tool.Name)
		}
	}
}

// The zero Permissions allows nothing, so forgetting to fill it in yields a
// server that refuses rather than one that is wide open.
func TestZeroPermissionsRefuseEveryWrite(t *testing.T) {
	w := &fakeWriter{}
	cs := connectWrite(t, w, nil)
	for _, tool := range WriteTools {
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      tool,
			Arguments: map[string]any{"id": 1, "read": true, "folder_id": 2, "flagged": true, "color": 1, "account_id": 1, "to": []string{"a@b.c"}, "subject": "s", "body": "b"},
		})
		if err == nil && !res.IsError {
			t.Errorf("%s ran with no permission", tool)
		}
	}
	if w.deleted != 0 || w.queuedID != 0 {
		t.Error("a refused tool still reached the mailbox")
	}
}

// A tool that is off is still listed, so an agent that is refused is told why
// rather than concluding the capability does not exist and trying something
// else.
func TestDisabledToolIsStillListed(t *testing.T) {
	cs := connectWrite(t, &fakeWriter{}, Permissions{ToolMarkRead: true})
	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	var found bool
	for _, tool := range tools.Tools {
		if tool.Name == ToolDeleteMessage {
			found = true
		}
	}
	if !found {
		t.Error("a disabled tool was hidden rather than listed")
	}
}

// A description must never name the current permission state. Clients cache
// the tool list, so a description saying "switched off" survives the user
// switching it on and then contradicts what the call actually does, which is
// how an agent ends up reporting that its successful calls did nothing.
func TestDescriptionsDoNotClaimAPermissionState(t *testing.T) {
	for _, perms := range []Permissions{nil, allPermissions()} {
		cs := connectWrite(t, &fakeWriter{}, perms)
		tools, err := cs.ListTools(context.Background(), nil)
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		for _, tool := range tools.Tools {
			if _, isWrite := ToolGroup[tool.Name]; !isWrite {
				continue
			}
			for _, claim := range []string{"switched off", "currently enabled", "is allowed"} {
				if strings.Contains(strings.ToLower(tool.Description), claim) {
					t.Errorf("%s description claims a permission state (%q): %q", tool.Name, claim, tool.Description)
				}
			}
			if !strings.Contains(tool.Description, "decided when you call it") {
				t.Errorf("%s description does not say permission is decided per call", tool.Name)
			}
		}
	}
}

// An agent should not have to guess what colour 1 is.
func TestFlagColorToolNamesTheColours(t *testing.T) {
	cs := connectWrite(t, &fakeWriter{}, allPermissions())
	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		if tool.Name != ToolSetFlagColor {
			continue
		}
		for _, want := range []string{"1 red", "8 pink", "0 clears"} {
			if !strings.Contains(tool.Description, want) {
				t.Errorf("set_flag_color does not document %q: %q", want, tool.Description)
			}
		}
	}
}

// Permissions are per tool: one being on says nothing about the others.
func TestPermissionsAreIndependent(t *testing.T) {
	w := &fakeWriter{}
	cs := connectWrite(t, w, Permissions{ToolMarkRead: true})

	if res := call(t, cs, ToolMarkRead, map[string]any{"id": 5, "read": true}); res == nil {
		t.Fatal("mark_read did not run though it was permitted")
	}
	if w.markedID != 5 || !w.markedRead {
		t.Errorf("mark_read reached the mailbox as id %d read %v", w.markedID, w.markedRead)
	}

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      ToolMoveMessage,
		Arguments: map[string]any{"id": 5, "folder_id": 9},
	})
	if err == nil && !res.IsError {
		t.Error("move ran though only mark_read was permitted")
	}
	if w.movedTo != 0 {
		t.Error("a refused move reached the mailbox")
	}
}

func TestWriteToolsReachTheMailbox(t *testing.T) {
	w := &fakeWriter{}
	cs := connectWrite(t, w, allPermissions())

	call(t, cs, ToolArchive, map[string]any{"id": 11})
	call(t, cs, ToolFlagMessage, map[string]any{"id": 12, "flagged": true})
	call(t, cs, ToolSetFlagColor, map[string]any{"id": 13, "color": 3})
	call(t, cs, ToolDeleteMessage, map[string]any{"id": 14})

	if w.archived != 11 {
		t.Errorf("archived = %d, want 11", w.archived)
	}
	if w.flagged != 12 {
		t.Errorf("flagged = %d, want 12", w.flagged)
	}
	if w.color != 3 {
		t.Errorf("color = %d, want 3", w.color)
	}
	if w.deleted != 14 {
		t.Errorf("deleted = %d, want 14", w.deleted)
	}
}

// send_message must not send. It queues, and it has to say so in words the
// agent will repeat back, or the user is told mail went out when it did not.
func TestSendQueuesAndSaysSo(t *testing.T) {
	w := &fakeWriter{}
	cs := connectWrite(t, w, allPermissions())

	text := callText(t, cs, ToolSendMessage, map[string]any{
		"account_id": 1,
		"to":         []string{"someone@example.com"},
		"subject":    "hello",
		"body":       "text",
	})
	if !strings.Contains(text, "awaiting") || !strings.Contains(text, "not been sent") {
		t.Errorf("send result does not say it is only queued: %q", text)
	}
	if len(w.queued.To) != 1 || w.queued.To[0] != "someone@example.com" {
		t.Errorf("queued recipients = %v", w.queued.To)
	}
	if w.queued.Subject != "hello" {
		t.Errorf("queued subject = %q", w.queued.Subject)
	}
}

// The delete tool's own description has to rule out permanent deletion, since
// that is what an agent reads when deciding whether it is safe to call.
func TestDeleteDescriptionRulesOutPermanentRemoval(t *testing.T) {
	cs := connectWrite(t, &fakeWriter{}, allPermissions())
	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == ToolDeleteMessage && !strings.Contains(tool.Description, "trash") {
			t.Errorf("delete_message does not say it only moves to trash: %q", tool.Description)
		}
	}
}

func TestEveryWriteToolHasAGroup(t *testing.T) {
	for _, tool := range WriteTools {
		if ToolGroup[tool] == "" {
			t.Errorf("write tool %s belongs to no group, so the settings ui cannot place it", tool)
		}
	}
}

// A port already held by another copy of Pelton is the common case, and the
// operating system's own wording gives no hint of that.
func TestListenErrorNamesTheLikelyCause(t *testing.T) {
	err := listenError("127.0.0.1:8765", syscall.EADDRINUSE)
	msg := err.Error()
	for _, want := range []string{"8765", "Another copy of Pelton", "different port"} {
		if !strings.Contains(msg, want) {
			t.Errorf("listen error does not mention %q: %q", want, msg)
		}
	}
}

// Anything else is passed through rather than blamed on a second copy.
func TestListenErrorPassesOtherFailuresThrough(t *testing.T) {
	err := listenError("127.0.0.1:8765", errors.New("no such device"))
	if strings.Contains(err.Error(), "Another copy") {
		t.Errorf("an unrelated failure was reported as a port clash: %q", err)
	}
	if !strings.Contains(err.Error(), "no such device") {
		t.Errorf("the underlying failure was lost: %q", err)
	}
}

// Switching a permission on must reach an agent that is already connected. The
// alternative was restarting the listener, which drops the session and fails
// outright when anything else holds the port.
func TestPermissionChangeReachesALiveSession(t *testing.T) {
	w := &fakeWriter{}
	perms := Permissions{}
	cs := connectLive(t, w, func() Permissions { return perms })

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      ToolArchive,
		Arguments: map[string]any{"id": 3},
	})
	if err == nil && !res.IsError {
		t.Fatal("archive ran before it was permitted")
	}

	perms = Permissions{ToolArchive: true}
	call(t, cs, ToolArchive, map[string]any{"id": 3})
	if w.archived != 3 {
		t.Error("the tool did not run after the permission was switched on")
	}

	perms = Permissions{}
	w.archived = 0
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      ToolArchive,
		Arguments: map[string]any{"id": 4},
	})
	if err == nil && !res.IsError {
		t.Error("archive still ran after the permission was switched off")
	}
	if w.archived != 0 {
		t.Error("a revoked permission still reached the mailbox")
	}
}

package imap

import (
	"net"
	"slices"
	"strings"
	"testing"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"
)

// the messages the tests put in the mailbox. Content does not matter, only that
// each is a distinct message with its own uid.
const testMessage = "From: someone@example.com\r\n" +
	"To: me@example.com\r\n" +
	"Subject: %s\r\n" +
	"\r\n" +
	"body\r\n"

// newTestServer starts an in-memory imap server on localhost with the given
// capabilities and returns a logged-in client with INBOX selected, holding one
// message per subject.
//
// Nothing here is a mock: it is the real go-imap server, so a plain EXPUNGE
// behaves exactly as it does against a real one, which is the entire point of
// these tests.
func newTestServer(t *testing.T, caps imap.CapSet, subjects ...string) *Client {
	t.Helper()

	mem := imapmemserver.New()
	user := imapmemserver.NewUser("user", "pass")
	if err := user.Create("INBOX", nil); err != nil {
		t.Fatalf("create INBOX: %v", err)
	}
	mem.AddUser(user)

	srv := imapserver.New(&imapserver.Options{
		NewSession: func(*imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return mem.NewSession(), nil, nil
		},
		Caps:         caps,
		InsecureAuth: true,
		Logger:       discardLogger{},
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	raw := imapclient.New(conn, nil)
	t.Cleanup(func() { _ = raw.Close() })
	if err := raw.Login("user", "pass").Wait(); err != nil {
		t.Fatalf("login: %v", err)
	}

	for _, subject := range subjects {
		body := strings.Replace(testMessage, "%s", subject, 1)
		cmd := raw.Append("INBOX", int64(len(body)), nil)
		if _, err := cmd.Write([]byte(body)); err != nil {
			t.Fatalf("append %q: %v", subject, err)
		}
		if err := cmd.Close(); err != nil {
			t.Fatalf("close append %q: %v", subject, err)
		}
		if _, err := cmd.Wait(); err != nil {
			t.Fatalf("append %q: %v", subject, err)
		}
	}

	c := &Client{raw: raw}
	if _, err := c.Select("INBOX"); err != nil {
		t.Fatalf("select: %v", err)
	}
	return c
}

// discardLogger keeps the server's connection logging out of the test output.
type discardLogger struct{}

func (discardLogger) Printf(string, ...any) {}

// remainingSubjects returns the subjects still in the mailbox, in uid order.
func remainingSubjects(t *testing.T, c *Client) []string {
	t.Helper()
	headers, err := c.FetchRecentHeaders(100)
	if err != nil {
		t.Fatalf("fetch headers: %v", err)
	}
	slices.SortFunc(headers, func(a, b MessageHeader) int { return int(a.UID) - int(b.UID) })
	out := make([]string, 0, len(headers))
	for _, h := range headers {
		out = append(out, h.Subject)
	}
	return out
}

// uidOf returns the uid of the message with the given subject.
func uidOf(t *testing.T, c *Client, subject string) imap.UID {
	t.Helper()
	headers, err := c.FetchRecentHeaders(100)
	if err != nil {
		t.Fatalf("fetch headers: %v", err)
	}
	for _, h := range headers {
		if h.Subject == subject {
			return h.UID
		}
	}
	t.Fatalf("no message with subject %q", subject)
	return 0
}

// TestDeleteMessagesLeavesAnotherClientsDeletedMail is the regression test for
// the destructive case: Thunderbird marked a message \Deleted and has not
// expunged it, Pelton is asked to delete a different one, and the server has no
// UIDPLUS. A plain EXPUNGE here takes both. Only the one Pelton was told to
// delete may go.
func TestDeleteMessagesLeavesAnotherClientsDeletedMail(t *testing.T) {
	c := newTestServer(t, nil, "theirs", "ours", "untouched")

	theirs := uidOf(t, c, "theirs")
	ours := uidOf(t, c, "ours")

	// another client marks its message and leaves it sitting there.
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}

	if err := c.DeleteMessages(ours); err != nil {
		t.Fatalf("DeleteMessages: %v", err)
	}

	got := remainingSubjects(t, c)
	want := []string{"theirs", "untouched"}
	if !slices.Equal(got, want) {
		t.Fatalf("mailbox holds %v, want %v", got, want)
	}

	// and their mark has to survive too, or their pending deletion is silently
	// undone.
	deleted, err := c.searchDeleted()
	if err != nil {
		t.Fatalf("search deleted: %v", err)
	}
	if !slices.Equal(deleted, []imap.UID{theirs}) {
		t.Errorf("\\Deleted uids = %v, want just %v", deleted, theirs)
	}
}

// TestDeleteMessagesWithUIDPlus covers the same scenario on a server that does
// advertise UIDPLUS, where UID EXPUNGE does the scoping for us.
func TestDeleteMessagesWithUIDPlus(t *testing.T) {
	caps := imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapUIDPlus: {}}
	c := newTestServer(t, caps, "theirs", "ours", "untouched")

	theirs := uidOf(t, c, "theirs")
	ours := uidOf(t, c, "ours")
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}

	if err := c.DeleteMessages(ours); err != nil {
		t.Fatalf("DeleteMessages: %v", err)
	}

	got := remainingSubjects(t, c)
	want := []string{"theirs", "untouched"}
	if !slices.Equal(got, want) {
		t.Fatalf("mailbox holds %v, want %v", got, want)
	}
}

// TestDeleteMessagesBatch deletes several at once, which is what a folder full
// of pending deletes or an emptied trash does.
func TestDeleteMessagesBatch(t *testing.T) {
	c := newTestServer(t, nil, "one", "two", "three", "keep")

	uids := []imap.UID{uidOf(t, c, "one"), uidOf(t, c, "two"), uidOf(t, c, "three")}
	if err := c.DeleteMessages(uids...); err != nil {
		t.Fatalf("DeleteMessages: %v", err)
	}

	got := remainingSubjects(t, c)
	if !slices.Equal(got, []string{"keep"}) {
		t.Fatalf("mailbox holds %v, want just [keep]", got)
	}
}

// TestDeleteMessagesNoUIDsIsANoOp guards the case that would be worst of all:
// an empty batch reaching a plain EXPUNGE and clearing whatever the mailbox
// happened to have marked.
func TestDeleteMessagesNoUIDsIsANoOp(t *testing.T) {
	c := newTestServer(t, nil, "theirs", "keep")

	theirs := uidOf(t, c, "theirs")
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}

	if err := c.DeleteMessages(); err != nil {
		t.Fatalf("DeleteMessages: %v", err)
	}

	got := remainingSubjects(t, c)
	if !slices.Equal(got, []string{"theirs", "keep"}) {
		t.Fatalf("mailbox holds %v, want both messages", got)
	}
}

// TestDeleteMessagesKeepsOurOwnEarlierMark covers a retry: a previous attempt
// already flagged our message, so it comes back from the search. It is ours, so
// it must not be treated as foreign and protected from its own deletion.
func TestDeleteMessagesKeepsOurOwnEarlierMark(t *testing.T) {
	c := newTestServer(t, nil, "ours", "keep")

	ours := uidOf(t, c, "ours")
	if err := c.storeMany([]imap.UID{ours}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("pre-mark ours: %v", err)
	}

	if err := c.DeleteMessages(ours); err != nil {
		t.Fatalf("DeleteMessages: %v", err)
	}
	if got := remainingSubjects(t, c); !slices.Equal(got, []string{"keep"}) {
		t.Fatalf("mailbox holds %v, want just [keep]", got)
	}
}

func TestForeignDeleted(t *testing.T) {
	tests := []struct {
		name     string
		existing []imap.UID
		ours     []imap.UID
		want     []imap.UID
	}{
		{"nothing marked", nil, []imap.UID{1}, nil},
		{"only ours marked", []imap.UID{1, 2}, []imap.UID{1, 2}, nil},
		{"one foreign", []imap.UID{1, 2, 3}, []imap.UID{2}, []imap.UID{1, 3}},
		{"all foreign", []imap.UID{7, 8}, []imap.UID{1}, []imap.UID{7, 8}},
		{"deleting nothing makes everything foreign", []imap.UID{7}, nil, []imap.UID{7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := foreignDeleted(tt.existing, tt.ours); !slices.Equal(got, tt.want) {
				t.Errorf("foreignDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

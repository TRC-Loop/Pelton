package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Writer is the mutating half of the mailbox, implemented by the desktop layer.
// It is a separate interface from Mailbox so that a server built without one
// cannot write at all, rather than relying on a permission check to hold.
//
// Nothing here removes mail from the server irreversibly: Delete moves to the
// trash, exactly as the delete key does, and emptying the trash stays something
// only the person at the keyboard can do.
type Writer interface {
	// MarkRead marks a message read or unread.
	MarkRead(ctx context.Context, messageID int64, read bool) error
	// Move moves a message to another folder of the same account.
	Move(ctx context.Context, messageID, folderID int64) error
	// Archive moves a message to the account's archive folder.
	Archive(ctx context.Context, messageID int64) error
	// Flag stars or unstars a message.
	Flag(ctx context.Context, messageID int64, flagged bool) error
	// SetFlagColor sets a message's colour label, or clears it with 0.
	SetFlagColor(ctx context.Context, messageID int64, color int) error
	// Delete moves a message to the trash. It never expunges.
	Delete(ctx context.Context, messageID int64) error
	// QueueSend puts a message in front of the person at the keyboard for
	// approval and returns the queued id. It does not send: an agent that has
	// just read a message cannot be trusted to have decided this on its own,
	// so the decision stays with the reader.
	QueueSend(ctx context.Context, msg OutgoingMessage) (int64, error)
}

// OutgoingMessage is a message an agent proposes sending.
type OutgoingMessage struct {
	AccountID int64
	To        []string
	Cc        []string
	Bcc       []string
	Subject   string
	Body      string
}

// ErrToolNotPermitted is returned to the agent when a write tool is registered
// but switched off. Registering every tool and refusing here, rather than
// hiding the tool, means the agent is told why instead of concluding the
// capability does not exist and trying something else.
var ErrToolNotPermitted = errors.New("this action is switched off in Pelton's settings")

// permissionNote is appended to every write tool's description. It says that
// permission is decided per call rather than naming the current state, which a
// cached tool list would get wrong.
const permissionNote = "The user switches each action on or off in Pelton and can change that at any time, " +
	"so whether this is permitted is decided when you call it: a call that returns without an error did what it says, " +
	"and one that is not permitted fails and says so. Never assume from this description either way."

// registerWriteTools adds the mutating tool set. It is a no-op when the caller
// supplied no Writer, which is what keeps a read-only build read-only.
func registerWriteTools(srv *mcp.Server, w Writer, perms func() Permissions) {
	if w == nil {
		return
	}

	addWrite(srv, perms, ToolMarkRead, "Mark a message read or unread.",
		func(ctx context.Context, in markReadInput) (any, error) {
			return done(w.MarkRead(ctx, in.ID, in.Read))
		})

	addWrite(srv, perms, ToolMoveMessage, "Move a message to another folder in the same account. Use list_folders for folder ids.",
		func(ctx context.Context, in moveInput) (any, error) {
			return done(w.Move(ctx, in.ID, in.FolderID))
		})

	addWrite(srv, perms, ToolArchive, "Move a message to its account's archive folder.",
		func(ctx context.Context, in messageInput) (any, error) {
			return done(w.Archive(ctx, in.ID))
		})

	addWrite(srv, perms, ToolFlagMessage, "Star or unstar a message.",
		func(ctx context.Context, in flagInput) (any, error) {
			return done(w.Flag(ctx, in.ID, in.Flagged))
		})

	addWrite(srv, perms, ToolSetFlagColor,
		"Set a message's colour label. 1 red, 2 orange, 3 yellow, 4 green, 5 teal, 6 blue, 7 purple, 8 pink, 0 clears it.",
		func(ctx context.Context, in colorInput) (any, error) {
			return done(w.SetFlagColor(ctx, in.ID, in.Color))
		})

	addWrite(srv, perms, ToolDeleteMessage,
		"Move a message to the trash. This does not delete it permanently; emptying the trash is not available to agents.",
		func(ctx context.Context, in messageInput) (any, error) {
			return done(w.Delete(ctx, in.ID))
		})

	addWrite(srv, perms, ToolSendMessage,
		"Propose sending a message. It is NOT sent: it is queued in Pelton for the user to read and approve or discard. "+
			"Say so when reporting back, rather than telling the user the mail has gone.",
		func(ctx context.Context, in sendInput) (any, error) {
			id, err := w.QueueSend(ctx, OutgoingMessage{
				AccountID: in.AccountID,
				To:        in.To,
				Cc:        in.Cc,
				Bcc:       in.Bcc,
				Subject:   in.Subject,
				Body:      in.Body,
			})
			if err != nil {
				return nil, err
			}
			return queuedOutput{
				Queued:  true,
				ID:      id,
				Message: "queued in Pelton and awaiting the user's approval. It has not been sent.",
			}, nil
		})
}

// addWrite registers one write tool behind its permission.
//
// The permission is read on every call, not captured here, so switching one on
// or off reaches an agent that is already connected. The description says
// nothing about the current state either: clients cache the tool list, so a
// description naming it would survive the change and then contradict what the
// call actually does.
func addWrite[In any](srv *mcp.Server, perms func() Permissions, name, description string, run func(context.Context, In) (any, error)) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        name,
		Description: description + " " + permissionNote,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		if !perms().Allows(name) {
			return nil, nil, fmt.Errorf("%s: %w", name, ErrToolNotPermitted)
		}
		out, err := run(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(out)
	})
}

// done turns a plain error return into the tool's success payload.
func done(err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return okOutput{OK: true}, nil
}

type okOutput struct {
	OK bool `json:"ok"`
}

type queuedOutput struct {
	Queued  bool   `json:"queued"`
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type messageInput struct {
	ID int64 `json:"id" jsonschema:"the message id, from list_messages or search_messages"`
}

type markReadInput struct {
	ID   int64 `json:"id" jsonschema:"the message id"`
	Read bool  `json:"read" jsonschema:"true to mark read, false to mark unread"`
}

type moveInput struct {
	ID       int64 `json:"id" jsonschema:"the message id"`
	FolderID int64 `json:"folder_id" jsonschema:"the destination folder id, from list_folders"`
}

type flagInput struct {
	ID      int64 `json:"id" jsonschema:"the message id"`
	Flagged bool  `json:"flagged" jsonschema:"true to star, false to unstar"`
}

type colorInput struct {
	ID    int64 `json:"id" jsonschema:"the message id"`
	Color int   `json:"color" jsonschema:"1 red, 2 orange, 3 yellow, 4 green, 5 teal, 6 blue, 7 purple, 8 pink, or 0 to clear the label"`
}

type sendInput struct {
	AccountID int64    `json:"account_id" jsonschema:"the account to send from, from list_accounts"`
	To        []string `json:"to" jsonschema:"recipient addresses"`
	Cc        []string `json:"cc,omitempty" jsonschema:"carbon copy addresses"`
	Bcc       []string `json:"bcc,omitempty" jsonschema:"blind carbon copy addresses"`
	Subject   string   `json:"subject" jsonschema:"the subject line"`
	Body      string   `json:"body" jsonschema:"the plain-text body"`
}

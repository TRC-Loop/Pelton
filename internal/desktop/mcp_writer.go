package desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/mcpserver"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// mcpWriter adapts the app's own message actions to mcpserver.Writer.
//
// Every method goes through the same binding the ui calls, so an agent's move
// is the same operation as the user's, with the same server round-trip and the
// same undo. What it adds is the record: each call is logged with what it did
// and to which message, because a folder change with no author is the one thing
// nobody can investigate afterwards.
type mcpWriter struct {
	app *App
}

func (w *mcpWriter) MarkRead(_ context.Context, id int64, read bool) error {
	err := w.app.SetSeen(id, read)
	state := "unread"
	if read {
		state = "read"
	}
	return w.record(mcpserver.ToolMarkRead, id, "marked "+state, err)
}

func (w *mcpWriter) Move(_ context.Context, id, folderID int64) error {
	_, err := w.app.MoveMessage(id, folderID)
	return w.record(mcpserver.ToolMoveMessage, id, "moved to "+w.folderName(folderID), err)
}

func (w *mcpWriter) Archive(_ context.Context, id int64) error {
	_, err := w.app.ArchiveMessage(id)
	return w.record(mcpserver.ToolArchive, id, "archived", err)
}

func (w *mcpWriter) Flag(_ context.Context, id int64, flagged bool) error {
	err := w.app.SetFlagged(id, flagged)
	action := "unstarred"
	if flagged {
		action = "starred"
	}
	return w.record(mcpserver.ToolFlagMessage, id, action, err)
}

func (w *mcpWriter) SetFlagColor(_ context.Context, id int64, color int) error {
	err := w.app.SetFlagColor(id, color)
	action := "cleared the colour label"
	if color != 0 {
		action = fmt.Sprintf("set colour label %d", color)
	}
	return w.record(mcpserver.ToolSetFlagColor, id, action, err)
}

// Delete moves to the trash. There is deliberately no path from here to an
// expunge: emptying the trash stays with the person at the keyboard.
func (w *mcpWriter) Delete(_ context.Context, id int64) error {
	err := w.app.DeleteMessage(id)
	return w.record(mcpserver.ToolDeleteMessage, id, "moved to the trash", err)
}

// QueueSend stores the message and raises it in the app. It does not send, and
// nothing is uploaded anywhere: a message the user has not seen must not reach
// a server, however confident the agent is that it should.
func (w *mcpWriter) QueueSend(_ context.Context, msg mcpserver.OutgoingMessage) (int64, error) {
	if len(msg.To) == 0 && len(msg.Cc) == 0 && len(msg.Bcc) == 0 {
		return 0, fmt.Errorf("desktop: a proposed message needs at least one recipient")
	}
	to := strings.Join(msg.To, ", ")
	id, err := w.app.store.CreateAgentProposal(w.app.ctx, storage.AgentProposal{
		AccountID: msg.AccountID,
		To:        to,
		Cc:        strings.Join(msg.Cc, ", "),
		Bcc:       strings.Join(msg.Bcc, ", "),
		Subject:   msg.Subject,
		Body:      msg.Body,
	})
	if err != nil {
		_ = w.record(mcpserver.ToolSendMessage, 0, "proposed a message to "+to, err)
		return 0, err
	}
	_ = w.record(mcpserver.ToolSendMessage, 0,
		"proposed a message to "+to+", awaiting your approval", nil)
	w.app.emit(EventAgentProposals, nil)
	return id, nil
}

// record logs the action and passes the error straight back, so logging can
// never change what the tool reports.
func (w *mcpWriter) record(tool string, messageID int64, summary string, err error) error {
	w.app.recordAgentAction(tool, messageID, summary, err)
	return err
}

// folderName resolves a folder id for the log line, falling back to the id when
// it cannot be read: a log entry is worth writing either way.
func (w *mcpWriter) folderName(id int64) string {
	folder, err := w.app.store.GetFolder(w.app.ctx, id)
	if err != nil || folder == nil {
		return fmt.Sprintf("folder %d", id)
	}
	return folder.Name
}

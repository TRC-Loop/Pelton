package desktop

import (
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// errStoreUnavailable is returned by bound methods when the store failed to open
// at startup, so the ui can show a clear error instead of hanging.
var errStoreUnavailable = errors.New("pelton: local store is unavailable")

// errSearchUnavailable is returned when the search index failed to open, so the
// ui can fall back gracefully instead of treating it as a hard error.
var errSearchUnavailable = errors.New("pelton: search index is unavailable")

// Event names emitted on the wails runtime event bus. The frontend subscribes to
// these by the same string keys (see src/lib/events.ts). Keep this list and the
// typescript EventName union in sync.
const (
	// EventMailNew fires when a sync or idle cycle pulled new messages.
	EventMailNew = "mail:new"
	// EventMailRepaired fires when a sync replaced the text of messages that
	// were cached with an encoding nothing could read, so the list reloads.
	EventMailRepaired = "mail:repaired"
	// EventSyncProgress fires as folders are synced so the ui can show progress.
	EventSyncProgress = "sync:progress"
	// EventSyncState fires when background sync starts or stops, with any error.
	EventSyncState = "sync:state"
	// EventOutboxChanged fires when the outbox contents or a message state change.
	EventOutboxChanged = "outbox:changed"
	// EventAgentProposals fires when an agent proposes a message, or one is
	// approved or discarded, so the queue in the ui matches what is stored.
	EventAgentProposals = "agent:proposals"
	// EventMenu fires when a native menu item is chosen. The payload is a short
	// action string (preferences, compose, sync, add-mailbox, about) the frontend
	// maps to an action.
	EventMenu = "menu:action"
	// EventDownloadProgress fires during a bulk offline range download so the
	// status bar can show a progress bar with percent and eta.
	EventDownloadProgress = "download:progress"
	// EventAttachmentProgress fires while saving one or more attachments.
	EventAttachmentProgress = "attachment:progress"
	// EventUpdateAvailable fires after an automatic (frequency-driven) update
	// check completes, so the about section can show a notice without polling.
	// It is not fired for a manual "check now" call, which returns its result
	// directly to the caller instead.
	EventUpdateAvailable = "update:available"
	// EventMailtoCompose fires when a mailto: link is opened while the app is
	// already running. The payload is a MailtoDraft the frontend opens as a
	// prefilled compose. A mailto that launched the app is delivered through
	// ConsumePendingMailto instead.
	EventMailtoCompose = "mailto:compose"
	// EventViewsChanged fires when saved views or their eager-run counts change,
	// so the sidebar can refresh view rows and badges without polling.
	EventViewsChanged = "views:changed"
	// EventImportProgress fires while mail is being imported from files, and
	// once more with Running false when the job ends.
	EventImportProgress = "import:progress"
	// EventProfileChanged fires when the app switches profile, or when the
	// profile it is in is edited. Everything the ui holds is scoped to a profile,
	// so the frontend reloads its preferences, sidebar, list and views rather
	// than trying to patch them.
	EventProfileChanged = "profile:changed"
)

// DownloadProgressEvent is the payload for EventDownloadProgress. Running is
// false on the final event (done, cancelled, or failed); Error is set on failure.
type DownloadProgressEvent struct {
	Running    bool   `json:"running"`
	Done       int    `json:"done"`
	Total      int    `json:"total"`
	Percent    int    `json:"percent"`
	ETASeconds int    `json:"etaSeconds"`
	Label      string `json:"label"`
	Error      string `json:"error"`
}

// AttachmentProgressEvent is the payload for EventAttachmentProgress. It reports
// byte progress for the current file plus how many files are done in a save-all.
type AttachmentProgressEvent struct {
	Running    bool   `json:"running"`
	Filename   string `json:"filename"`
	BytesDone  int64  `json:"bytesDone"`
	BytesTotal int64  `json:"bytesTotal"`
	FilesDone  int    `json:"filesDone"`
	FilesTotal int    `json:"filesTotal"`
	Error      string `json:"error"`
}

// ImportProgressEvent is the payload for EventImportProgress. Running is false
// on the final event, which carries the totals; Error is set on failure.
// BytesDone/BytesTotal measure the source files, because an mbox does not say
// how many messages it holds until it has been read.
type ImportProgressEvent struct {
	Running    bool     `json:"running"`
	Folder     string   `json:"folder"`
	Imported   int      `json:"imported"`
	Skipped    int      `json:"skipped"`
	Failed     int      `json:"failed"`
	BytesDone  int64    `json:"bytesDone"`
	BytesTotal int64    `json:"bytesTotal"`
	Folders    []string `json:"folders"`
	Error      string   `json:"error"`
}

// MailNewEvent is the payload for EventMailNew.
type MailNewEvent struct {
	AccountID int64 `json:"accountId"`
	FolderID  int64 `json:"folderId"`
	Count     int   `json:"count"`
}

// MailRepairedEvent is the payload for EventMailRepaired. Count is how many
// messages in the folder were rewritten.
type MailRepairedEvent struct {
	AccountID int64 `json:"accountId"`
	FolderID  int64 `json:"folderId"`
	Count     int   `json:"count"`
}

// SyncProgressEvent is the payload for EventSyncProgress.
type SyncProgressEvent struct {
	AccountID int64  `json:"accountId"`
	Folder    string `json:"folder"`
	Done      int    `json:"done"`
	Total     int    `json:"total"`
}

// SyncStateEvent is the payload for EventSyncState. Error is empty on success.
type SyncStateEvent struct {
	Running bool   `json:"running"`
	Error   string `json:"error"`
}

// emit sends a runtime event if the context is set. It is a thin wrapper so call
// sites stay terse and a nil context (before startup) is a safe no-op.
func (a *App) emit(name string, payload any) {
	if a.ctx == nil || !a.runtimeReady.Load() {
		return
	}
	runtime.EventsEmit(a.ctx, name, payload)
}

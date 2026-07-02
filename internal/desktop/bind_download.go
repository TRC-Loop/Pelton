package desktop

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// maxPreviewBytes caps how large an attachment we will stream to the ui for the
// in-app preview. Above this we tell the ui to offer "open externally" instead of
// pushing tens of megabytes of base64 across the bridge.
const maxPreviewBytes = 25 << 20 // 25 MiB

// downloadActive guards against two bulk downloads running at once.
var downloadActive atomic.Bool

// AttachmentContentDTO carries one attachment's bytes to the previewer. Data is
// base64 so it crosses the bindings as a plain string. TooLarge is set (with no
// Data) when the file exceeds the preview cap.
type AttachmentContentDTO struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	Data        string `json:"data"`
	TooLarge    bool   `json:"tooLarge"`
}

// ReadAttachment returns an attachment's bytes for the in-app previewer. messageID
// scopes the lookup so an id cannot reach another message's files.
func (a *App) ReadAttachment(messageID, attachmentID int64) (AttachmentContentDTO, error) {
	if err := a.ready(); err != nil {
		return AttachmentContentDTO{}, err
	}
	target, err := a.findAttachment(messageID, attachmentID)
	if err != nil {
		return AttachmentContentDTO{}, err
	}
	dto := AttachmentContentDTO{
		Filename:    target.Filename,
		ContentType: target.ContentType,
		SizeBytes:   target.SizeBytes,
	}
	if target.SizeBytes > maxPreviewBytes {
		dto.TooLarge = true
		return dto, nil
	}
	rc, err := a.store.OpenAttachment(target.DiskPath)
	if err != nil {
		return AttachmentContentDTO{}, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return AttachmentContentDTO{}, err
	}
	dto.Data = base64.StdEncoding.EncodeToString(data)
	return dto, nil
}

// findAttachment resolves one attachment row within a message.
func (a *App) findAttachment(messageID, attachmentID int64) (*storage.Attachment, error) {
	atts, err := a.store.ListAttachments(a.ctx, messageID)
	if err != nil {
		return nil, err
	}
	for i := range atts {
		if atts[i].ID == attachmentID {
			return &atts[i], nil
		}
	}
	return nil, fmt.Errorf("pelton: attachment %d not found", attachmentID)
}

// SaveAllAttachments prompts for a directory and writes every non-inline
// attachment of a message there, emitting progress. It returns the chosen
// directory (empty if cancelled).
func (a *App) SaveAllAttachments(messageID int64) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	atts, err := a.store.ListAttachments(a.ctx, messageID)
	if err != nil {
		return "", err
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Save all attachments to folder"})
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", nil
	}

	total := len(atts)
	defer a.emit(EventAttachmentProgress, AttachmentProgressEvent{Running: false, FilesDone: total, FilesTotal: total})
	for i, att := range atts {
		dest := uniqueDestPath(dir, att.Filename)
		if err := a.copyAttachmentProgress(att, dest, i, total); err != nil {
			a.emit(EventAttachmentProgress, AttachmentProgressEvent{Running: false, Error: err.Error(), FilesDone: i, FilesTotal: total})
			return "", err
		}
	}
	return dir, nil
}

// copyAttachmentProgress streams one attachment to dest, emitting byte progress
// so the ui can show a bar even though the source is a local cached file.
func (a *App) copyAttachmentProgress(att storage.Attachment, dest string, fileIndex, filesTotal int) error {
	src, err := a.store.OpenAttachment(att.DiskPath)
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(filepath.Clean(dest))
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{
		total:    att.SizeBytes,
		filename: att.Filename,
		fileIdx:  fileIndex,
		files:    filesTotal,
		emit:     a.emit,
	}
	if _, err := io.Copy(io.MultiWriter(out, pw), src); err != nil {
		return err
	}
	return nil
}

// progressWriter counts bytes copied and emits attachment progress events. It
// throttles to at most one event per ~64 KiB so a stream of tiny writes does not
// flood the event bus.
type progressWriter struct {
	total    int64
	written  int64
	lastEmit int64
	filename string
	fileIdx  int
	files    int
	emit     func(string, any)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.written += int64(n)
	if w.written-w.lastEmit >= 64<<10 || w.written == w.total {
		w.lastEmit = w.written
		w.emit(EventAttachmentProgress, AttachmentProgressEvent{
			Running:    true,
			Filename:   w.filename,
			BytesDone:  w.written,
			BytesTotal: w.total,
			FilesDone:  w.fileIdx,
			FilesTotal: w.files,
		})
	}
	return n, nil
}

// DownloadRange downloads every message from startDateRFC3339 to today that is
// not already cached, across all accounts and folders, and pins them offline for
// fast local search. includeAttachments controls whether attachment bytes are
// persisted. Progress (percent + eta) is emitted for the status bar.
func (a *App) DownloadRange(startDateRFC3339 string, includeAttachments bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	since, err := time.Parse(time.RFC3339, startDateRFC3339)
	if err != nil {
		// tolerate a plain date (YYYY-MM-DD) from a date picker with no time.
		since, err = time.Parse("2006-01-02", startDateRFC3339)
		if err != nil {
			return fmt.Errorf("pelton: invalid start date %q: %w", startDateRFC3339, err)
		}
	}
	if a.lowPowerMode() {
		return errors.New("pelton: low power mode is on; turn it off to start a bulk download")
	}
	if !downloadActive.CompareAndSwap(false, true) {
		return errors.New("pelton: a download is already running")
	}

	// remember the attachment choice as the default for next time.
	_ = a.store.SetBool(a.ctx, settingDownloadAtts, includeAttachments)
	// remember the range itself so a restart mid-download can pick back up
	// instead of silently dropping the job (see ResumePendingDownload).
	_ = a.store.Set(a.ctx, settingDownloadPending, since.Format(time.RFC3339))

	// run the whole job on a background goroutine so the bound call returns
	// immediately and neither the ui nor the go caller waits on imap. progress
	// and completion are reported entirely through events.
	go a.runRangeDownload(since, includeAttachments)
	return nil
}

// ResumePendingDownload restarts a bulk download that was still running when
// the app last closed. planDownload/planAccount already skip anything cached,
// so replaying the same range only fetches whatever the previous run had not
// gotten to yet. Called once from startup; a no-op if nothing was pending.
func (a *App) ResumePendingDownload() {
	raw, err := a.store.Get(a.ctx, settingDownloadPending)
	if err != nil || raw == "" {
		return
	}
	since, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		_ = a.store.Set(a.ctx, settingDownloadPending, "")
		return
	}
	if a.lowPowerMode() {
		return
	}
	includeAttachments := a.boolSetting(settingDownloadAtts, true)
	if !downloadActive.CompareAndSwap(false, true) {
		return
	}
	go a.runRangeDownload(since, includeAttachments)
}

// DownloadEstimateDTO is the offline-download space estimate: how many
// messages are planned and their combined raw size, so the settings ui can
// show "~4.7 GB" before the user commits to a potentially large download.
type DownloadEstimateDTO struct {
	MessageCount int   `json:"messageCount"`
	TotalBytes   int64 `json:"totalBytes"`
}

// EstimateDownloadRange reuses the same planning pass as DownloadRange (so the
// count always matches what a real run would fetch) and additionally sums the
// RFC822 size of each planned message. This is the actual number of bytes
// that travel over IMAP either way; whether attachments are kept afterward
// only affects what gets written to disk, not what gets fetched.
func (a *App) EstimateDownloadRange(startDateRFC3339 string) (DownloadEstimateDTO, error) {
	if err := a.ready(); err != nil {
		return DownloadEstimateDTO{}, err
	}
	since, err := time.Parse(time.RFC3339, startDateRFC3339)
	if err != nil {
		since, err = time.Parse("2006-01-02", startDateRFC3339)
		if err != nil {
			return DownloadEstimateDTO{}, fmt.Errorf("pelton: invalid start date %q: %w", startDateRFC3339, err)
		}
	}
	tasks, err := a.planDownload(since)
	if err != nil {
		return DownloadEstimateDTO{}, err
	}
	if len(tasks) == 0 {
		return DownloadEstimateDTO{}, nil
	}

	// group planned uids by folder id (folders are not comparable, so the id
	// keys the map) so each folder is fetched in one batch, then group folders
	// by account so each account is connected to only once.
	uidsByFolder := map[int64][]imap.UID{}
	foldersByID := map[int64]storage.Folder{}
	folderIDsByAccount := map[int64][]int64{}
	for _, task := range tasks {
		fid := task.folder.ID
		if _, seen := foldersByID[fid]; !seen {
			foldersByID[fid] = task.folder
			folderIDsByAccount[task.folder.AccountID] = append(folderIDsByAccount[task.folder.AccountID], fid)
		}
		uidsByFolder[fid] = append(uidsByFolder[fid], imap.UID(task.uid))
	}
	accounts, err := a.store.ListAccounts(a.ctx)
	if err != nil {
		return DownloadEstimateDTO{}, err
	}

	var total int64
	for _, account := range accounts {
		folderIDs, ok := folderIDsByAccount[account.ID]
		if !ok {
			continue
		}
		size, err := a.estimateAccountSize(account, folderIDs, foldersByID, uidsByFolder)
		if err != nil {
			a.log.Error("estimate download size", "account", account.Email, "err", err)
			continue
		}
		total += size
	}
	return DownloadEstimateDTO{MessageCount: len(tasks), TotalBytes: total}, nil
}

// estimateAccountSize connects to one account and sums RFC822 sizes across
// its planned folders.
func (a *App) estimateAccountSize(account storage.Account, folderIDs []int64, foldersByID map[int64]storage.Folder, uidsByFolder map[int64][]imap.UID) (int64, error) {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return 0, err
	}
	syncMu.Lock()
	defer syncMu.Unlock()

	client, err := pimap.Connect(cfg)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return 0, err
	}
	defer client.Logout()

	var total int64
	for _, fid := range folderIDs {
		folder := foldersByID[fid]
		if _, err := client.Select(folder.IMAPPath); err != nil {
			a.log.Error("estimate select", "folder", folder.IMAPPath, "err", err)
			continue
		}
		sizes, err := client.FetchSizes(uidsByFolder[fid])
		if err != nil {
			a.log.Error("estimate fetch sizes", "folder", folder.IMAPPath, "err", err)
			continue
		}
		for _, size := range sizes {
			total += size
		}
	}
	return total, nil
}

// runRangeDownload performs the plan-and-fetch passes off the calling goroutine.
func (a *App) runRangeDownload(since time.Time, includeAttachments bool) {
	defer downloadActive.Store(false)
	// the job is no longer resumable once it either finishes or gives up;
	// clear the marker before returning on every exit path.
	defer func() { _ = a.store.Set(a.ctx, settingDownloadPending, "") }()

	a.emit(EventDownloadProgress, DownloadProgressEvent{Running: true, Label: "Scanning"})
	tasks, err := a.planDownload(since)
	if err != nil {
		a.emit(EventDownloadProgress, DownloadProgressEvent{Running: false, Error: err.Error()})
		return
	}
	total := len(tasks)
	if total == 0 {
		a.emit(EventDownloadProgress, DownloadProgressEvent{Running: false, Label: "Nothing to download"})
		return
	}

	a.emit(EventDownloadProgress, DownloadProgressEvent{Running: true, Total: total, Label: "Starting"})
	if err := a.runDownload(tasks, includeAttachments, total); err != nil {
		a.emit(EventDownloadProgress, DownloadProgressEvent{Running: false, Error: err.Error()})
		return
	}
	a.emit(EventDownloadProgress, DownloadProgressEvent{Running: false, Done: total, Total: total, Percent: 100, Label: "Done"})
}

// dlTask is one message to fetch, paired with the folder it belongs to.
type dlTask struct {
	folder storage.Folder
	uid    uint32
}

// planDownload connects to each account, searches every folder for messages
// since the cutoff, and returns the ones not yet cached. It is the cheap counting
// pass that lets the fetch pass report an accurate percentage and eta.
func (a *App) planDownload(since time.Time) ([]dlTask, error) {
	accounts, err := a.store.ListAccounts(a.ctx)
	if err != nil {
		return nil, err
	}
	var tasks []dlTask
	for _, account := range accounts {
		if err := a.ctx.Err(); err != nil {
			return nil, err
		}
		accTasks, err := a.planAccount(account, since)
		if err != nil {
			if errors.Is(err, errNoCredentials) {
				continue
			}
			a.log.Error("plan download", "account", account.Email, "err", err)
			continue
		}
		tasks = append(tasks, accTasks...)
	}
	return tasks, nil
}

// planAccount opens one account and lists the uncached message uids since the
// cutoff across its folders.
func (a *App) planAccount(account storage.Account, since time.Time) ([]dlTask, error) {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return nil, err
	}
	syncMu.Lock()
	defer syncMu.Unlock()

	client, err := pimap.Connect(cfg)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return nil, err
	}
	defer client.Logout()

	folders, err := a.store.ListFolders(a.ctx, account.ID)
	if err != nil {
		return nil, err
	}
	var tasks []dlTask
	for _, folder := range folders {
		if _, err := client.Select(folder.IMAPPath); err != nil {
			a.log.Error("plan select", "folder", folder.IMAPPath, "err", err)
			continue
		}
		uids, err := client.SearchSince(since)
		if err != nil {
			a.log.Error("plan search", "folder", folder.IMAPPath, "err", err)
			continue
		}
		have, err := a.cachedUIDs(folder.ID)
		if err != nil {
			return nil, err
		}
		for _, uid := range uids {
			if _, ok := have[uint32(uid)]; !ok {
				tasks = append(tasks, dlTask{folder: folder, uid: uint32(uid)})
			}
		}
	}
	return tasks, nil
}

// cachedUIDs returns the set of uids already stored for a folder.
func (a *App) cachedUIDs(folderID int64) (map[uint32]struct{}, error) {
	states, err := a.store.ListMessageStates(a.ctx, folderID)
	if err != nil {
		return nil, err
	}
	set := make(map[uint32]struct{}, len(states))
	for _, s := range states {
		set[s.UID] = struct{}{}
	}
	return set, nil
}

// runDownload fetches the planned messages account by account, storing each and
// pinning it offline, and emits progress with percent and a running eta.
func (a *App) runDownload(tasks []dlTask, includeAttachments bool, total int) error {
	byAccount := groupByAccount(tasks)
	start := time.Now()
	done := 0

	for accountID, accTasks := range byAccount {
		if err := a.ctx.Err(); err != nil {
			return err
		}
		account, err := a.store.GetAccount(a.ctx, accountID)
		if err != nil {
			a.log.Error("download get account", "id", accountID, "err", err)
			done += len(accTasks)
			continue
		}
		if err := a.downloadAccount(*account, accTasks, includeAttachments, &done, total, start); err != nil {
			if errors.Is(err, errNoCredentials) {
				done += len(accTasks)
				continue
			}
			a.log.Error("download account", "account", account.Email, "err", err)
		}
	}
	return nil
}

// downloadAccount fetches every task for one account over a single connection.
func (a *App) downloadAccount(account storage.Account, tasks []dlTask, includeAttachments bool, done *int, total int, start time.Time) error {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return err
	}
	syncMu.Lock()
	defer syncMu.Unlock()

	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	defer client.Logout()

	selected := ""
	for _, task := range tasks {
		if err := a.ctx.Err(); err != nil {
			return err
		}
		if selected != task.folder.IMAPPath {
			if _, err := client.Select(task.folder.IMAPPath); err != nil {
				a.log.Error("download select", "folder", task.folder.IMAPPath, "err", err)
				*done++
				continue
			}
			selected = task.folder.IMAPPath
		}
		if err := a.fetchAndPin(client, task, includeAttachments); err != nil {
			a.log.Error("download fetch", "uid", task.uid, "err", err)
		}
		*done++
		a.emitDownloadProgress(*done, total, start, account.Email)
	}
	return nil
}

// fetchAndPin fetches one message and stores it pinned offline. Attachment bytes
// are persisted only when includeAttachments is set.
func (a *App) fetchAndPin(client *pimap.Client, task dlTask, includeAttachments bool) error {
	msg, err := client.FetchMessage(imap.UID(task.uid))
	if err != nil {
		return err
	}
	stored := &storage.Message{
		AccountID:   task.folder.AccountID,
		FolderID:    task.folder.ID,
		UID:         uint32(msg.UID),
		MessageID:   msg.MessageID,
		Subject:     msg.Subject,
		FromAddress: msg.From,
		ToAddresses: msg.To,
		CcAddresses: msg.Cc,
		Date:        msg.Date,
		Flags:       0,
		BodyPlain:   msg.Text,
		BodyHTML:    msg.HTML,
		SizeBytes:   msg.Size,
		Offline:     true,
	}
	var atts []storage.IncomingAttachment
	if includeAttachments {
		for _, at := range msg.Attachments {
			atts = append(atts, storage.IncomingAttachment{
				Filename:    at.Filename,
				ContentType: at.ContentType,
				ContentID:   at.ContentID,
				Content:     bytes.NewReader(at.Content),
			})
		}
	}
	id, err := a.store.InsertMessageWithAttachments(a.ctx, stored, atts)
	if err != nil {
		return err
	}
	return a.store.SetOffline(a.ctx, id, true)
}

// emitDownloadProgress computes percent and eta and emits a progress event.
func (a *App) emitDownloadProgress(done, total int, start time.Time, label string) {
	percent := 0
	if total > 0 {
		percent = done * 100 / total
	}
	eta := 0
	if done > 0 {
		elapsed := time.Since(start).Seconds()
		perItem := elapsed / float64(done)
		eta = int(perItem * float64(total-done))
	}
	a.emit(EventDownloadProgress, DownloadProgressEvent{
		Running: true, Done: done, Total: total, Percent: percent, ETASeconds: eta, Label: label,
	})
}

// groupByAccount buckets download tasks by their folder's account id.
func groupByAccount(tasks []dlTask) map[int64][]dlTask {
	out := make(map[int64][]dlTask)
	for _, t := range tasks {
		out[t.folder.AccountID] = append(out[t.folder.AccountID], t)
	}
	return out
}

// uniqueDestPath appends " (n)" before the extension until the path is free, so a
// save-all never silently overwrites two attachments that share a name.
func uniqueDestPath(dir, filename string) string {
	base := filepath.Base(filename)
	dest := filepath.Join(dir, base)
	if !fileExistsAt(dest) {
		return dest
	}
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	for i := 1; i < 10000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, i, ext))
		if !fileExistsAt(candidate) {
			return candidate
		}
	}
	return dest
}

func fileExistsAt(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

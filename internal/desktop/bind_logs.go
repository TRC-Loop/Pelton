package desktop

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/logging"
)

var (
	// errNoLogDir means no data directory could be resolved, so there is
	// nowhere to put a log.
	errNoLogDir = errors.New("pelton: no log folder is available")
	// errNoCrashReport means the crash the ui was offering is gone, most
	// likely deleted between the prompt and the click.
	errNoCrashReport = errors.New("pelton: there is no crash report to open")
)

// LogStatusDTO is what the settings ui needs to describe logging without
// reading the disk itself.
type LogStatusDTO struct {
	// Dir is the log folder, shown next to the toggle so it is obvious where
	// the files land. Empty when no data directory could be resolved.
	Dir string `json:"dir"`
	// Writing reports whether lines are reaching a file right now, which is not
	// the same as the setting: --debug and PELTON_DEBUG force it on.
	Writing bool `json:"writing"`
	// Forced marks that override, so the ui can say the toggle is not in charge
	// at the moment instead of looking broken.
	Forced bool `json:"forced"`
	// SizeBytes is what the logs and crash files currently take up.
	SizeBytes int64 `json:"sizeBytes"`
	// CrashName and CrashTime describe a crash from a previous run the user has
	// not been told about yet. CrashName is empty when there is none.
	CrashName string `json:"crashName"`
	CrashTime string `json:"crashTime"`
}

// GetLogStatus reports where logs go, whether they are being written, and
// whether the last run ended in a crash the user has not seen yet.
func (a *App) GetLogStatus() LogStatusDTO {
	dir := a.logDir()
	status := LogStatusDTO{
		Dir:       dir,
		Writing:   a.logWriter.Enabled(),
		Forced:    a.debug,
		SizeBytes: dirSize(dir),
	}
	if crash, ok := logging.PendingCrash(dir); ok {
		status.CrashName = crash.Name
		if !crash.When.IsZero() {
			status.CrashTime = crash.When.Format("2006-01-02 15:04")
		}
	}
	return status
}

// OpenLogFolder shows the log folder in the system file manager. It creates the
// folder first, so the button does something sensible before anything has been
// logged rather than failing on a path that does not exist yet.
func (a *App) OpenLogFolder() error {
	dir := a.logDir()
	if dir == "" {
		return errNoLogDir
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return openPath(dir)
}

// OpenCrashReport opens the pending crash file in whatever the system opens
// plain text with, and marks it seen so the next launch stops offering it.
func (a *App) OpenCrashReport() error {
	dir := a.logDir()
	crash, ok := logging.PendingCrash(dir)
	if !ok {
		return errNoCrashReport
	}
	if err := openPath(crash.Path); err != nil {
		return err
	}
	return logging.AcknowledgeCrashes(dir)
}

// DismissCrashReport marks the pending crash seen without opening it. The file
// stays on disk: dismissing a prompt is not the same as deciding the report is
// worthless.
func (a *App) DismissCrashReport() error {
	return logging.AcknowledgeCrashes(a.logDir())
}

// DeleteLogs removes every log and crash file. Offered when logging is turned
// off and when an account is deleted, because a log written while an account
// existed is still a record of using it.
func (a *App) DeleteLogs() error {
	dir := a.logDir()
	if dir == "" {
		return errNoLogDir
	}
	// stop writing first, so the active file is closed before it is removed and
	// the next line starts a fresh one rather than an unlinked handle.
	writing := a.logWriter.Enabled()
	if err := a.logWriter.Disable(); err != nil {
		return err
	}
	if err := logging.RemoveLogs(dir); err != nil {
		return err
	}
	if err := logging.RemoveCrashes(dir); err != nil {
		return err
	}
	if writing {
		return a.logWriter.Enable(dir)
	}
	return nil
}

// GetDiagnostics returns the plain-text summary the About section copies to the
// clipboard, so a bug report can start with the build and platform instead of a
// round of questions.
//
// It is built from what the app already knows about itself. No mail, no
// addresses, no folder names, and it runs through the same redactor the log
// files use.
func (a *App) GetDiagnostics() string {
	channel := a.channel
	if channel == "" {
		channel = "stable"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Pelton %s (%s)\n", a.version, channel)
	fmt.Fprintf(&b, "OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&b, "Go: %s\n", runtime.Version())
	fmt.Fprintf(&b, "Language: %s\n", a.stringSetting(settingLanguage, "en"))
	if a.store != nil {
		if accounts, err := a.store.ListAccounts(a.ctx); err == nil {
			fmt.Fprintf(&b, "Mailboxes: %d\n", len(accounts))
		}
	} else {
		b.WriteString("Mailboxes: none, the database did not open\n")
	}
	fmt.Fprintf(&b, "File logging: %t\n", a.logWriter.Enabled())
	fmt.Fprintf(&b, "Crash reports: %t\n", a.crashLogsOn())
	if crash, ok := logging.PendingCrash(a.logDir()); ok {
		fmt.Fprintf(&b, "Unreported crash: %s\n", crash.Name)
	}
	return logging.Redact(b.String())
}

// dirSize sums the files directly in dir. Returns 0 for a directory that does
// not exist, which is the normal state with logging off.
func dirSize(dir string) int64 {
	if dir == "" {
		return 0
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var total int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if info, err := e.Info(); err == nil {
			total += info.Size()
		}
	}
	return total
}

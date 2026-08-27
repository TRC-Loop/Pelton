//go:build windows

package desktop

import (
	"os"
	"path/filepath"
	"sync"

	"git.sr.ht/~jackmordaunt/go-toast"
)

// Windows notifications (#170). A toast does not carry a name or an icon of its
// own: Windows looks both up from the AppUserModelID the toast was posted
// under. An id nothing has registered gets the placeholder treatment, which is
// how Pelton's toasts came to read "DefaultAppName" next to a grey grid - that
// really is the string beeep posts under when nobody sets beeep.AppName.
//
// Registering the id is one write under HKCU\SOFTWARE\Classes\AppUserModelId,
// and it is also what makes a click reach us: the same key holds the
// CustomActivator that Windows calls back through. beeep exposes none of that,
// so Windows talks to go-toast directly rather than through it.
//
// This is deliberately not left to the installer. A portable build, a dev run
// and every copy installed before this change would all be left with the
// placeholder, and none of them are reinstalling to fix a notification. The
// uninstaller still removes the key (see build/windows/installer/project.nsi),
// so an uninstall leaves nothing behind.

// notifyIconFile is the icon written into the data directory for the registry
// to point at. The .ico is only ever embedded in the binary, and IconUri needs
// a path, so it is unpacked once and left there.
const notifyIconFile = "notification.ico"

// Toast setup is process-wide because the registry and go-toast's activation
// callback are process-wide. There is exactly one App per process, so a
// package-level once is the honest shape for it.
var (
	toastOnce  sync.Once
	toastAppID string
	toastIcon  string
	toastExe   string
)

// prepareToasts registers the app id, unpacks the icon and wires the click
// callback, once per run. Everything it does is idempotent, and go-toast skips
// a registry write when the value is already there, so on every launch after
// the first this is a few registry reads.
func (a *App) prepareToasts() {
	toastOnce.Do(func() {
		// the registry DisplayName is set from the app id, so this string is
		// what the user reads under the toast. Nightlies register their own for
		// the same reason they carry their own tray icon: a nightly toast should
		// never look like it came from the real install.
		toastAppID = notifyAppName
		if a.IsNightly() {
			toastAppID = notifyAppName + " Nightly"
		}
		toastExe, _ = os.Executable()
		toastIcon = a.unpackNotifyIcon()

		toast.SetActivationCallback(a.toastActivated)

		// no GUID: go-toast computes the CLSID key path from its default
		// activator GUID at package init, so a custom one would register the
		// activator under a path Windows never looks at.
		if err := toast.SetAppData(toast.AppData{
			AppID:         toastAppID,
			IconPath:      toastIcon,
			ActivationExe: toastExe,
		}); err != nil {
			// the toast still shows, it just shows under the placeholder name.
			a.log.Warn("register the notification app id", "err", err)
		}
	})
}

// unpackNotifyIcon writes the embedded icon into the data directory and returns
// its path, or "" when there is nothing to write or nowhere to write it. An
// empty path means the toast goes out without an icon rather than not at all.
func (a *App) unpackNotifyIcon() string {
	if len(a.trayIcon) == 0 || a.dataDir == "" {
		return ""
	}
	path := filepath.Join(a.dataDir, notifyIconFile)
	// rewritten on every first-notification of a run rather than only when
	// missing, so a build whose icon changed does not keep showing the old one.
	if err := os.WriteFile(path, a.trayIcon, 0o600); err != nil {
		a.log.Warn("write the notification icon", "path", path, "err", err)
		return ""
	}
	return path
}

// toastActivated runs when a toast is clicked, on a callback goroutine of
// go-toast's making. args is whatever was put on the notification.
func (a *App) toastActivated(args string, _ []toast.UserData) {
	id, ok := notifyMessageID(args)
	if !ok {
		// a click we cannot place still asked for Pelton, so bring it up.
		a.showWindow()
		return
	}
	a.openFromNotification(id)
}

// deliverNotification raises one Windows toast.
//
// Foreground activation is what enables the click callback at all; without it
// Windows has nothing to invoke and a click only dismisses the toast. The audio
// is silenced on purpose, matching macOS: what Pelton should make a noise about
// is its own question (#240).
func (a *App) deliverNotification(n notification) error {
	a.prepareToasts()
	t := toast.Notification{
		AppID:               toastAppID,
		Title:               n.title,
		Body:                n.body,
		Icon:                toastIcon,
		ActivationType:      toast.Foreground,
		ActivationArguments: notifyArgs(n.messageID),
		ActivationExe:       toastExe,
		Audio:               toast.Silent,
		Duration:            toast.Short,
	}
	return t.Push()
}

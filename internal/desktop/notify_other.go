//go:build !darwin && !windows

package desktop

import "github.com/gen2brain/beeep"

// deliverNotification raises one OS notification through beeep, which on Linux
// is a dbus notification. The icon comes from the running executable's desktop
// file, so there is no icon to hand over here, but the app name is not
// something beeep can work out: left alone it sends its own "DefaultAppName"
// placeholder as the dbus app_name (#170).
//
// beeep has no per-notification click callback, so the message id on the
// notification goes unused here. macOS and Windows have their own backends
// (notify_darwin.go, notify_windows.go).
func (a *App) deliverNotification(n notification) error {
	beeep.AppName = notifyAppName
	return beeep.Notify(n.title, n.body, "")
}

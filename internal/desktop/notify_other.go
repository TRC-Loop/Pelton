//go:build !darwin

package desktop

import "github.com/gen2brain/beeep"

// deliverNotification raises one OS notification through beeep: a toast on
// Windows, dbus on Linux. Both take the icon from the running executable, so
// neither needs the platform-specific handling macOS does (see
// notify_darwin.go).
func deliverNotification(title, body string) error {
	return beeep.Notify(title, body, "")
}

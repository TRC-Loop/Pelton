//go:build windows

package desktop

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// bringToFront raises the window above every other application.
//
// WindowShow maps the window but does not necessarily raise it. Windows refuses
// to let a process that does not own the foreground steal it, and when a second
// launch surfaces the running instance the caller is the new process while the
// window belongs to the old one, so nothing comes forward: double-clicking the
// desktop icon with Pelton in the tray appeared to do nothing at all.
//
// Making the window topmost and immediately dropping it back goes through
// SetWindowPos, which has no such restriction, and leaves the z-order with
// Pelton on top. There is no always-on-top preference for this to fight with.
func bringToFront(ctx context.Context) {
	wailsruntime.WindowSetAlwaysOnTop(ctx, true)
	wailsruntime.WindowSetAlwaysOnTop(ctx, false)
}

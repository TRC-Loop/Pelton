package desktop

import (
	"context"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// close button actions stored under settingCloseAction.
const (
	// closeActionBackground hides the window and keeps sync running.
	closeActionBackground = "background"
	// closeActionQuit exits the app.
	closeActionQuit = "quit"
)

// quitApp exits the app from an explicit user action (the Quit menu item, the
// tray's Quit). It marks the quit as intended so beforeClose lets it through
// instead of treating it as a window close.
func (a *App) quitApp() {
	if a.ctx == nil {
		return
	}
	a.quitRequested.Store(true)
	wailsruntime.Quit(a.ctx)
}

// hideOnClose reports whether closing the window should hide it and leave sync,
// idle and the outbox worker running, rather than exit. That is the default; the
// "quit" close action and an explicit quit (quitApp) both exit instead.
func (a *App) hideOnClose() bool {
	if a.quitRequested.Load() {
		return false
	}
	return a.stringSetting(settingCloseAction, closeActionBackground) != closeActionQuit
}

// beforeClose is wails' OnBeforeClose hook. Returning true prevents the quit,
// which is how the window ends up hidden instead of the app exiting.
func (a *App) beforeClose(_ context.Context) bool {
	if !a.hideOnClose() {
		return false
	}
	// Hide, not WindowHide: on macOS this is [NSApp hide:], which the Dock icon
	// can undo. Ordering the window out instead would strand it, since wails has
	// no reopen handler. On Windows and Linux the two are the same call.
	wailsruntime.Hide(a.ctx)
	return true
}

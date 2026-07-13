//go:build !windows

package desktop

// The tray icon is Windows-only (#84): macOS reopens the hidden window via
// the Dock, and on Linux tray protocols are too fragmented to rely on.

func (a *App) startTray() {}

func (a *App) stopTray() {}

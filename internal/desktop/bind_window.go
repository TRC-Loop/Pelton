package desktop

import "github.com/wailsapp/wails/v2/pkg/runtime"

// SetWindowTitle updates the native window title so it can reflect context (the
// open message's subject, the current folder, etc.). A blank title is ignored so
// the window never ends up nameless.
func (a *App) SetWindowTitle(title string) {
	if a.ctx == nil || title == "" {
		return
	}
	runtime.WindowSetTitle(a.ctx, title)
}

// TitleBarDoubleClick runs the system's title-bar double-click action. On macOS
// the app draws its own title bar, so the window server never sees the click
// and cannot act on it; the frontend forwards it here instead. A no-op on every
// other platform, where the native frame handles it.
func (a *App) TitleBarDoubleClick() {
	if a.ctx == nil {
		return
	}
	switch titleBarDoubleClickAction() {
	case "maximize":
		runtime.WindowToggleMaximise(a.ctx)
	case "minimize":
		runtime.WindowMinimise(a.ctx)
	}
}

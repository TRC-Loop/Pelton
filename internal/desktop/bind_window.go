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

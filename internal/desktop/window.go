package desktop

import wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

// showWindow restores and focuses the main window. It is used by the Windows
// tray, by the single-instance handler (a second launch surfaces the running
// window), and by the macOS mailto url handler. A nil context means the webview
// is not up yet, so there is nothing to show.
func (a *App) showWindow() {
	if a.ctx == nil {
		return
	}
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
}

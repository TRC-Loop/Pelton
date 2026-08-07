//go:build !windows

package desktop

// Minimizing to the notification area only means something where there is one,
// which is Windows (see tray_stub.go). Elsewhere the minimize button is left
// alone and the setting is never offered in the ui.

func (a *App) watchMinimize() {}

//go:build windows

package desktop

import "github.com/energye/systray"

// The Windows notification-area icon (#84). With HideWindowOnClose, closing
// the window leaves Pelton syncing in the background with no visible way to
// get it back or quit it; the tray icon is that way. Left click reopens the
// window, right click shows a small menu wired to the same runtime calls and
// menu action events as the native menu. systray's Windows backend is plain
// win32 syscalls: no cgo, no network.

// startTray brings up the tray icon on its own OS thread. systray.Run blocks
// inside its message loop, so it runs in a goroutine for the app's lifetime;
// stopTray ends it at shutdown.
func (a *App) startTray() {
	go systray.Run(a.trayReady, nil)
}

// stopTray removes the tray icon. Safe to call even if the tray never came up.
func (a *App) stopTray() {
	systray.Quit()
}

// trayReady builds the icon and menu once systray's loop is up. Labels use
// the same Go-side translation table as the native menu; like that menu, the
// tray is built in the language active at startup.
func (a *App) trayReady() {
	if len(a.trayIcon) > 0 {
		systray.SetIcon(a.trayIcon)
	}
	systray.SetTooltip("Pelton")

	s := menuStringsFor(a.stringSetting(settingLanguage, "en"))
	open := systray.AddMenuItem(s.openWindow, "")
	compose := systray.AddMenuItem(s.compose, "")
	syncNow := systray.AddMenuItem(s.syncNow, "")
	systray.AddSeparator()
	quit := systray.AddMenuItem(s.quit, "")

	open.Click(a.showWindow)
	compose.Click(func() {
		a.showWindow()
		a.emit(EventMenu, "compose")
	})
	syncNow.Click(func() {
		a.emit(EventMenu, "sync")
	})
	quit.Click(a.quitApp)

	// left click reopens the window; the menu only shows on right click.
	systray.SetOnClick(func(systray.IMenu) {
		a.showWindow()
	})
	systray.SetOnRClick(func(m systray.IMenu) {
		_ = m.ShowMenu()
	})
}

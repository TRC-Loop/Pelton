package desktop

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildMenu builds the native application menubar. Menu items that map to ui
// actions emit the menu event with a short action string; the frontend listens
// and performs the action (open settings, compose, sync, add mailbox). Window
// level items (hide, quit) call the wails runtime directly. Accelerators use
// CmdOrCtrl so they read as Cmd on macos and Ctrl elsewhere automatically, which
// is the localized-to-platform behavior users expect.
func (a *App) buildMenu() *menu.Menu {
	root := menu.NewMenu()

	// the app menu. on macos this folds under the "Pelton" bold menu next to the
	// apple logo; on other platforms it is a normal "Pelton" menu.
	appMenu := root.AddSubmenu("Pelton")
	appMenu.AddText("About Pelton", nil, a.menuAction("about"))
	appMenu.AddSeparator()
	appMenu.AddText("Preferences…", keys.CmdOrCtrl(","), a.menuAction("preferences"))
	appMenu.AddSeparator()
	if goruntime.GOOS == "darwin" {
		appMenu.AddText("Hide Pelton", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
			wailsruntime.WindowHide(a.ctx)
		})
	}
	appMenu.AddText("Quit Pelton", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsruntime.Quit(a.ctx)
	})

	// file menu: composing new mail and exporting the open message.
	fileMenu := root.AddSubmenu("File")
	fileMenu.AddText("Compose", keys.CmdOrCtrl("n"), a.menuAction("compose"))
	fileMenu.AddSeparator()
	fileMenu.AddText("Export Message as PDF…", keys.CmdOrCtrl("p"), a.menuAction("export-pdf"))

	// mailbox menu: sync, account management, and the message-level actions
	// that make sense to reach without the mouse (or a memorized shortcut).
	// Undo has no accelerator here since Cmd/Ctrl+Z is already handled by the
	// app's own keydown handler (the undo-send/delete/archive chain); binding
	// it again here would double-fire.
	mailMenu := root.AddSubmenu("Mailbox")
	mailMenu.AddText("Sync Now", keys.CmdOrCtrl("r"), a.menuAction("sync"))
	mailMenu.AddSeparator()
	mailMenu.AddText("Add Mailbox…", keys.CmdOrCtrl("m"), a.menuAction("add-mailbox"))
	mailMenu.AddSeparator()
	mailMenu.AddText("Undo", nil, a.menuAction("undo"))
	mailMenu.AddSeparator()
	mailMenu.AddText("Mark as Read", nil, a.menuAction("mark-read"))
	mailMenu.AddText("Mark as Unread", nil, a.menuAction("mark-unread"))
	mailMenu.AddText("Flag / Unflag", nil, a.menuAction("flag"))
	mailMenu.AddText("Archive", nil, a.menuAction("archive"))
	mailMenu.AddText("Delete Message", nil, a.menuAction("delete-message"))
	mailMenu.AddSeparator()
	mailMenu.AddText("Low Power Mode", nil, a.menuAction("toggle-low-power"))

	// view menu: a reliable fullscreen toggle. the native green button can be
	// inconsistent in some setups, so this guarantees the app can go fullscreen.
	viewMenu := root.AddSubmenu("View")
	viewMenu.AddText("Toggle Fullscreen", keys.Combo("f", keys.CmdOrCtrlKey, keys.ControlKey), func(_ *menu.CallbackData) {
		if wailsruntime.WindowIsFullscreen(a.ctx) {
			wailsruntime.WindowUnfullscreen(a.ctx)
		} else {
			wailsruntime.WindowFullscreen(a.ctx)
		}
	})

	// the standard edit menu gives copy/paste/select-all their native bindings,
	// which the webview needs on macos to work in inputs and the mail body.
	if goruntime.GOOS == "darwin" {
		root.Append(menu.EditMenu())
	}

	return root
}

// menuAction returns a menu callback that emits the menu event with the given
// action string.
func (a *App) menuAction(action string) menu.Callback {
	return func(_ *menu.CallbackData) {
		a.emit(EventMenu, action)
	}
}

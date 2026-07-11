package desktop

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildMenu builds the native application menubar in the current language
// setting. Menu items that map to ui actions emit the menu event with a short
// action string; the frontend listens and performs the action (open settings,
// compose, sync, add mailbox). Window level items (hide, quit) call the wails
// runtime directly. Accelerators use CmdOrCtrl so they read as Cmd on macos
// and Ctrl elsewhere automatically, which is the localized-to-platform
// behavior users expect.
func (a *App) buildMenu() *menu.Menu {
	s := menuStringsFor(a.stringSetting(settingLanguage, "en"))
	root := menu.NewMenu()

	// the app menu. on macos this folds under the "Pelton" bold menu next to the
	// apple logo; on other platforms it is a normal "Pelton" menu.
	appMenu := root.AddSubmenu(s.appMenu)
	appMenu.AddText(s.about, nil, a.menuAction("about"))
	appMenu.AddSeparator()
	appMenu.AddText(s.preferences, keys.CmdOrCtrl(","), a.menuAction("preferences"))
	appMenu.AddSeparator()
	if goruntime.GOOS == "darwin" {
		appMenu.AddText(s.hide, keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
			wailsruntime.WindowHide(a.ctx)
		})
	}
	appMenu.AddText(s.quit, keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsruntime.Quit(a.ctx)
	})

	// file menu: composing new mail and exporting the open message.
	fileMenu := root.AddSubmenu(s.fileMenu)
	fileMenu.AddText(s.compose, keys.CmdOrCtrl("n"), a.menuAction("compose"))
	fileMenu.AddSeparator()
	fileMenu.AddText(s.exportPDF, keys.CmdOrCtrl("p"), a.menuAction("export-pdf"))

	// mailbox menu: mailbox-level operations - syncing and managing accounts.
	mailboxMenu := root.AddSubmenu(s.mailboxMenu)
	mailboxMenu.AddText(s.syncNow, keys.CmdOrCtrl("r"), a.menuAction("sync"))
	mailboxMenu.AddSeparator()
	mailboxMenu.AddText(s.addMailbox, keys.CmdOrCtrl("m"), a.menuAction("add-mailbox"))
	mailboxMenu.AddText(s.manageMailboxes, nil, a.menuAction("open-mailboxes"))

	// mail menu: actions on the open message. Undo stays enabled (it undoes the
	// last send/delete/archive, which needs no open message), but the message-
	// level items below start disabled and are only enabled while a message is
	// open (SetMailActionsEnabled, driven by the frontend's selection). Undo has
	// no accelerator here since Cmd/Ctrl+Z is already handled by the app's own
	// keydown handler; binding it again would double-fire.
	mailMenu := root.AddSubmenu(s.mailMenu)
	mailMenu.AddText(s.undo, nil, a.menuAction("undo"))
	mailMenu.AddSeparator()
	a.mailMenuItems = []*menu.MenuItem{
		mailMenu.AddText(s.markRead, nil, a.menuAction("mark-read")),
		mailMenu.AddText(s.markUnread, nil, a.menuAction("mark-unread")),
		mailMenu.AddText(s.flagUnflag, nil, a.menuAction("flag")),
		mailMenu.AddText(s.archive, nil, a.menuAction("archive")),
		mailMenu.AddText(s.deleteMessage, nil, a.menuAction("delete-message")),
	}
	for _, item := range a.mailMenuItems {
		item.Disabled = !a.mailActionsEnabled
	}

	// view menu: a reliable fullscreen toggle (the native green button can be
	// inconsistent in some setups) plus the low-power mode toggle.
	viewMenu := root.AddSubmenu(s.viewMenu)
	viewMenu.AddText(s.toggleFullscreen, keys.Combo("f", keys.CmdOrCtrlKey, keys.ControlKey), func(_ *menu.CallbackData) {
		if wailsruntime.WindowIsFullscreen(a.ctx) {
			wailsruntime.WindowUnfullscreen(a.ctx)
		} else {
			wailsruntime.WindowFullscreen(a.ctx)
		}
	})
	viewMenu.AddSeparator()
	viewMenu.AddText(s.lowPowerMode, nil, a.menuAction("toggle-low-power"))

	// the standard edit menu gives copy/paste/select-all their native bindings,
	// which the webview needs on macos to work in inputs and the mail body.
	// wails' EditMenu() ships with its own fixed English labels (Cut/Copy/
	// Paste/Select All/Undo/Redo), which it doesn't expose a way to translate.
	if goruntime.GOOS == "darwin" {
		root.Append(menu.EditMenu())
	}

	return root
}

// RebuildMenu rebuilds and applies the native menubar in the current language
// setting. The frontend calls this right after writing a new language setting
// so the menu updates immediately instead of waiting for the next launch.
//
// Skipped on Linux: wails' GTK MenuSetApplicationMenu does not cleanly replace
// the previous native menu's click-handler wiring there, and once poisoned
// every subsequent menu click (not just the item that triggered the rebuild)
// nil-pointer-crashes the whole process inside wails' own gtk.go
// handleMenuItemClick. buildMenu already reads the persisted language setting
// fresh each call, so the menu still comes up correctly translated on the
// next launch; it just doesn't live-update mid-session on Linux, which is a
// far better tradeoff than crashing the app.
func (a *App) RebuildMenu() {
	if a.ctx == nil || goruntime.GOOS == "linux" {
		return
	}
	wailsruntime.MenuSetApplicationMenu(a.ctx, a.buildMenu())
}

// SetMailActionsEnabled greys out or restores the Mail menu's message-level
// items. The frontend calls it as its open message changes, so those actions are
// only selectable while a message is actually open. The chosen state is kept on
// the App so a later menu rebuild (RebuildMenu, on a language change) restores
// it instead of resetting every item back to disabled.
func (a *App) SetMailActionsEnabled(enabled bool) {
	a.mailActionsEnabled = enabled
	for _, item := range a.mailMenuItems {
		item.Disabled = !enabled
	}
	if a.ctx != nil {
		wailsruntime.MenuUpdateApplicationMenu(a.ctx)
	}
}

// menuAction returns a menu callback that emits the menu event with the given
// action string.
func (a *App) menuAction(action string) menu.Callback {
	return func(_ *menu.CallbackData) {
		a.emit(EventMenu, action)
	}
}

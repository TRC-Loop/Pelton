# Shortcuts

Pelton uses ++cmd++ on macOS and ++ctrl++ on Windows and Linux for the same bindings. The tables below write ++cmd++; substitute accordingly.

## App

| Shortcut | Action |
| -------- | ------ |
| ++cmd+n++ | Compose a new message |
| ++cmd+k++ | [Command palette](command-palette.md) |
| ++cmd+f++ | Search |
| ++cmd+r++ | Sync now |
| ++cmd+m++ | Add mailbox |
| ++cmd+comma++ | Settings |
| ++cmd+p++ | Export open message as PDF |
| ++cmd+z++ | Undo the last send, delete or archive |
| ++ctrl+cmd+f++ | Toggle fullscreen |
| ++cmd+h++ | Hide window (macOS) |
| ++cmd+w++ | Close the front compose window, settings, or the window itself |
| ++cmd+q++ | Quit |

## Menu bar access keys

On Windows and Linux, where the in-app menu bar is the only menu there is, ++alt++ on its own focuses the bar and underlines one letter in each menu title. ++alt++ plus that letter opens the menu straight away. Arrow keys move between menus and items from there, ++esc++ leaves.

The letters are worked out from the titles themselves, so they follow your language and any menu you have renamed or added. A binding you set under **Settings, Shortcuts** always wins over an access key using the same combination.

macOS has no such convention and does not use them.

## Message actions

Reply, reply all, forward, mark read or unread, flag, snooze, archive, delete and download-for-offline ship unbound so they cannot collide with anything. Bind them to whatever you like under **Settings, Shortcuts**; each one acts on the currently open message.

## Developer overlays

Only in a development run, started with `PELTON_DEV` (what `make run` sets) or `PELTON_DEVTOOLS=1`. A normal build does not bind these keys at all.

| Shortcut | Action |
| -------- | ------ |
| ++f6++ | Activity log: what the backend is doing, live |
| ++f7++ | Process: goroutines, heap, database and cache sizes |
| ++f8++ | Frames: frame timing for the ui |

The overlays are draggable and several can be open at once. They always start closed.

The activity log shows the same lines Pelton writes to its log, so it follows the log level. Set `PELTON_DEBUG=1` for the full sync detail.

The webview inspector is the browser's own, not Pelton's: ++f12++ on Windows, ++cmd+alt+i++ on macOS, right-click and Inspect on Linux. It only exists in a development build.

## Rebinding

**Settings, Shortcuts** lists every action with its current combo. Click one and press the new combination to rebind it, including the defaults above. If a combo is already taken, Pelton tells you what it is bound to.

## Vim modes

Two independent toggles for keyboard-centric use:

- **App vim mode** (**Settings, Shortcuts**): `h`/`j`/`k`/`l` style navigation across the message list and panes.
- **Compose vim mode** (**Settings, Composing**): vim keybindings inside the compose editor.

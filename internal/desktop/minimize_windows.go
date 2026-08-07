//go:build windows

package desktop

import (
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Minimize to the notification area (#167). Wails v2 exposes no minimize hook
// on any platform, so there is nothing to intercept SC_MINIMIZE with: the
// window state is watched instead, and a window that has just gone minimized
// while the setting says "tray" is hidden. That leaves the taskbar minimize
// animation visible for one tick before the window disappears, which is the
// cost of not reaching behind wails to subclass its window procedure.
//
// The alternative, replacing the window proc with a syscall.NewCallback, would
// hide the animation but puts a hand-written callback in front of every message
// wails' own window handling depends on, and a mistake there crashes the
// process rather than misbehaving. Not worth it for one frame of animation.

// minimizeWatchInterval is how often the window state is sampled. Fast enough
// that the window does not sit in the taskbar, slow enough to be free.
const minimizeWatchInterval = 150 * time.Millisecond

// watchMinimize sends the window to the tray when the user minimizes it and the
// minimize action is set to "tray". It runs for the app's lifetime; the setting
// is read per tick so a change applies without a restart.
//
// wasMinimized keeps it to the transition: without it, a window the user
// minimized before turning the setting on would be yanked away the moment they
// did, and an unhidden-but-still-minimized window would fight the user.
func (a *App) watchMinimize() {
	ticker := time.NewTicker(minimizeWatchInterval)
	defer ticker.Stop()

	wasMinimized := false
	for range ticker.C {
		if a.ctx == nil {
			continue
		}
		minimized := wailsruntime.WindowIsMinimised(a.ctx)
		if minimized && !wasMinimized && a.minimizeToTray() {
			// unminimise first: a window hidden while iconic comes back
			// minimized when the tray reopens it.
			wailsruntime.WindowUnminimise(a.ctx)
			wailsruntime.WindowHide(a.ctx)
			minimized = false
		}
		wasMinimized = minimized
	}
}

// minimizeToTray reports whether the minimize button should hide the window to
// the notification area instead of minimizing it normally.
func (a *App) minimizeToTray() bool {
	return a.stringSetting(settingMinimizeAction, minimizeActionNormal) == minimizeActionTray
}

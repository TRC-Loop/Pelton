package desktop

// SetDockBadge puts the unread count on the dock icon. The frontend calls it
// whenever the unified inbox count changes, which already covers syncing, marking
// read and deleting, so there is no second count to keep in step here.
//
// A no-op anywhere without a dock tile; see dockbadge_other.go.
func (a *App) SetDockBadge(unread int) {
	if unread < 0 {
		unread = 0
	}
	a.badgeMu.Lock()
	a.unreadBadge = unread
	a.badgeMu.Unlock()
	a.applyDockBadge()
}

// applyDockBadge pushes the remembered count to the platform, or clears the
// badge when the setting is off. Also called when that setting changes, so
// turning it back on does not wait for the next sidebar refresh.
func (a *App) applyDockBadge() {
	a.badgeMu.Lock()
	unread := a.unreadBadge
	a.badgeMu.Unlock()

	if !a.boolSetting(settingDockBadge, true) {
		unread = 0
	}
	setPlatformBadge(unread)
}

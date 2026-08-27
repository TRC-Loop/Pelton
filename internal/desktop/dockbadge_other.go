//go:build !darwin

package desktop

// Only macOS has a dock tile to badge. Windows wants a taskbar overlay icon
// through ITaskbarList3, which needs COM and an icon rendered per count, and
// Linux has no convention that works across desktops; both are their own
// problem rather than a stub away.
func setPlatformBadge(_ int) {}

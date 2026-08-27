//go:build !windows

package desktop

import "context"

// bringToFront is a no-op away from Windows. macOS and the Linux window
// managers both raise a window that is shown, and forcing always-on-top there
// would reorder Pelton against other applications for no reason. See
// window_windows.go for why Windows needs the nudge.
func bringToFront(context.Context) {}

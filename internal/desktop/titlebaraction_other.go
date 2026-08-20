//go:build !darwin

package desktop

// Only macOS has the app draw its own title bar, and only macOS has a system
// setting for what double-clicking one does. Everywhere else the window keeps
// its native frame and the system handles the click itself.

func titleBarDoubleClickAction() string {
	return "none"
}

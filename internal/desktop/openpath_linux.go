//go:build linux

package desktop

import "os/exec"

// xdg-open is the desktop-agnostic launcher, and inside a flatpak sandbox it is
// the portal shim, so the packaged build opens the same folders as the plain
// binary without a separate path.
func openCommand(path string) (*exec.Cmd, error) {
	return exec.Command("xdg-open", path), nil
}

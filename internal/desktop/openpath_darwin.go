//go:build darwin

package desktop

import "os/exec"

// open is macOS's launcher: a folder opens in Finder, a file in the
// application registered for its type.
func openCommand(path string) (*exec.Cmd, error) {
	return exec.Command("open", path), nil
}

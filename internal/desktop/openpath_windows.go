//go:build windows

package desktop

import "os/exec"

// explorer is the shell's own launcher and takes both folders and files. It is
// run directly rather than through "cmd /c start", which would put the path
// through the command interpreter.
func openCommand(path string) (*exec.Cmd, error) {
	return exec.Command("explorer", path), nil
}

//go:build !darwin && !linux && !windows

package desktop

import (
	"errors"
	"os/exec"
)

// Any other platform has no launcher Pelton knows how to call, so the buttons
// that reveal a folder report that instead of appearing to work.
func openCommand(string) (*exec.Cmd, error) {
	return nil, errors.New("pelton: opening a folder is not supported on this platform")
}

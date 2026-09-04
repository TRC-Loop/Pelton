package desktop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// errNothingToOpen means no path was resolved, so there is nothing to hand the
// desktop.
var errNothingToOpen = errors.New("pelton: there is nothing to open")

// openPath asks the desktop to open a file or folder: a folder lands in the
// file manager, a file in whatever application handles its type. The
// invocation itself is built per platform in openpath_<os>.go, and arguments go
// through argv rather than a shell, so a path with spaces needs no escaping.
//
// wailsruntime.BrowserOpenURL cannot do this. Wails runs every url it is handed
// through a validator that rejects the file scheme outright, so a
// "file://"+path call logs "Invalid URL" to Wails' own logger and returns,
// leaving the button dead with nothing for Pelton to report. Every platform
// takes that path, so it is not one desktop misbehaving.
func openPath(path string) error {
	if path == "" {
		return errNothingToOpen
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("pelton: open %q: %w", path, err)
	}
	// checked here rather than left to the launcher, which reports a missing
	// path as a bare exit code long after this call has returned.
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("pelton: open %q: %w", abs, err)
	}
	cmd, err := openCommand(abs)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("pelton: open %q: %w", abs, err)
	}
	// the launcher hands the path to the real application and exits. reap it in
	// the background so it leaves no zombie and the caller is not left waiting
	// on a file manager that may take a moment to appear.
	go func() { _ = cmd.Wait() }()
	return nil
}

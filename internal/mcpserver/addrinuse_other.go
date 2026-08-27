//go:build !windows

package mcpserver

import (
	"errors"
	"syscall"
)

// addrInUse reports whether a listen failed because something already holds the
// port. Windows reports this as its own winsock code, so the check is split
// rather than compared against a message, which would depend on the locale.
func addrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}

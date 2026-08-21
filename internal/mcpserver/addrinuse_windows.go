package mcpserver

import (
	"errors"
	"syscall"
)

// wsaEAddrInUse is winsock's WSAEADDRINUSE. Go's syscall package does not name
// it on Windows, so the value is written out rather than matched against the
// error's text, which changes with the system language.
const wsaEAddrInUse = syscall.Errno(10048)

// addrInUse reports whether a listen failed because something already holds the
// port. Windows returns the winsock code rather than the posix one.
func addrInUse(err error) bool {
	return errors.Is(err, wsaEAddrInUse) || errors.Is(err, syscall.EADDRINUSE)
}

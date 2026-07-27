package desktop

import (
	"errors"
	"net"
	"strings"
)

// errOffline is returned when an operation failed because the machine could not
// reach the mail server (no DNS, no route, refused dial, timeout). the frontend
// recognizes the "offline" marker and shows one honest "no internet connection"
// line instead of the raw dial error, or the misleading "no credentials" the
// sync path reported when a dropped connection was the real cause.
var errOffline = errors.New("pelton: offline")

// offlineOrErr maps a connectivity failure to errOffline so the frontend can
// show its one honest offline line, and passes any other error through.
func offlineOrErr(err error) error {
	if isNetworkError(err) {
		return errOffline
	}
	return err
}

// isNetworkError reports whether err is a connectivity failure rather than a
// protocol or credential problem. it checks the typed net errors first, then
// falls back to the message text for the DNS/dial cases go wraps as opaque
// strings across platforms.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"no such host",
		"dial tcp",
		"connection refused",
		"network is unreachable",
		"no route to host",
		"i/o timeout",
		"timeout",
		"tls handshake",
		"lookup ",
		"connection reset",
		"host is down",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

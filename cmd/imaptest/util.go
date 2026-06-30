package main

import (
	"strings"

	"github.com/emersion/go-imap/v2"
)

func attrString(attrs []imap.MailboxAttr) string {
	parts := make([]string, len(attrs))
	for i, a := range attrs {
		parts[i] = string(a)
	}
	return strings.Join(parts, " ")
}

func flagString(flags []imap.Flag) string {
	if len(flags) == 0 {
		return "(none)"
	}
	parts := make([]string, len(flags))
	for i, f := range flags {
		parts[i] = string(f)
	}
	return strings.Join(parts, " ")
}

func snippet(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(no plain-text body)"
	}
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

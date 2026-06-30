package imap

import (
	"fmt"
	"strings"

	"github.com/emersion/go-imap/v2"
)

// formatAddresses renders an address list as `Name <user@host>, ...`.
func formatAddresses(addrs []imap.Address) string {
	if len(addrs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		email := a.Addr()
		switch {
		case a.Name != "" && email != "":
			parts = append(parts, fmt.Sprintf("%s <%s>", a.Name, email))
		case email != "":
			parts = append(parts, email)
		case a.Name != "":
			parts = append(parts, a.Name)
		}
	}
	return strings.Join(parts, ", ")
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

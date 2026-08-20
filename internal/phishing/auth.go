// Package phishing decides whether a message is pretending to be from someone
// it is not. It answers a different question from a malware scanner: a phishing
// mail usually carries nothing malicious to find, and the whole trick is that
// it looks like it came from your bank, your employer or a colleague.
//
// Everything here is local and pure. No lookups, no reputation service, no
// network of any kind: the inputs are the message's own headers and body plus
// what Pelton already knows about who the user corresponds with.
//
// The checks are weighed rather than fired one at a time. Mailing lists,
// marketing senders and forwarders break naive SPF and link-domain rules
// constantly, so a single mismatch is not an accusation, and the wording that
// comes out of this never claims more than the signals support.
package phishing

import (
	"strings"
)

// Auth is what the receiving server said about the sender's authentication,
// read out of the Authentication-Results header. Empty fields mean the header
// said nothing about that method, which is not the same as a failure.
type Auth struct {
	// SPF, DKIM and DMARC are the raw results: pass, fail, softfail, neutral,
	// none, temperror, permerror, or empty when unstated.
	SPF   string
	DKIM  string
	DMARC string
	// SPFDomain is the smtp.mailfrom (or smtp.helo) domain the spf result
	// applies to; DKIMDomain is the signing domain from header.d. Both are what
	// alignment is judged on.
	SPFDomain  string
	DKIMDomain string
}

// Known results, lowercased as they appear in the header.
const (
	ResultPass      = "pass"
	ResultFail      = "fail"
	ResultSoftFail  = "softfail"
	ResultNeutral   = "neutral"
	ResultNone      = "none"
	ResultTempError = "temperror"
	ResultPermError = "permerror"
)

// Stated reports whether the header said anything at all about authentication.
// Mail from before Pelton stored these, and mail from a server that adds no
// header, has nothing to judge and must not be shown as suspicious.
func (a Auth) Stated() bool {
	return a.SPF != "" || a.DKIM != "" || a.DMARC != ""
}

// ParseAuth reads every Authentication-Results header of a message into one
// result. A message can carry several, added by each hop: the first is the one
// the user's own server wrote and is the only one that can be trusted, so
// later headers only fill in methods the first left unstated.
//
// The format is RFC 8601: an authserv-id, then method=result pairs, each
// optionally followed by property=value pairs.
func ParseAuth(headers []string) Auth {
	var out Auth
	for _, header := range headers {
		merge(&out, parseOne(header))
	}
	return out
}

// merge fills only the fields the earlier headers left empty.
func merge(dst *Auth, src Auth) {
	if dst.SPF == "" {
		dst.SPF, dst.SPFDomain = src.SPF, src.SPFDomain
	}
	if dst.DKIM == "" {
		dst.DKIM, dst.DKIMDomain = src.DKIM, src.DKIMDomain
	}
	if dst.DMARC == "" {
		dst.DMARC = src.DMARC
	}
}

// parseOne reads a single Authentication-Results value.
func parseOne(header string) Auth {
	var out Auth
	// the authserv-id comes first, before the first semicolon, and carries no
	// result of its own.
	parts := strings.Split(header, ";")
	if len(parts) < 2 {
		return out
	}
	for _, part := range parts[1:] {
		method, result, props := parseMethod(part)
		switch method {
		case "spf":
			if out.SPF == "" {
				out.SPF = result
				out.SPFDomain = domainOf(firstOf(props, "smtp.mailfrom", "smtp.helo"))
			}
		case "dkim":
			// several dkim signatures can be reported; a pass wins over an
			// earlier fail, since one valid signature is enough to authenticate.
			if out.DKIM == "" || (out.DKIM != ResultPass && result == ResultPass) {
				out.DKIM = result
				out.DKIMDomain = domainOf(firstOf(props, "header.d", "header.i"))
			}
		case "dmarc":
			if out.DMARC == "" {
				out.DMARC = result
			}
		}
	}
	return out
}

// parseMethod splits one "method=result prop=value prop=value" clause.
func parseMethod(clause string) (method, result string, props map[string]string) {
	fields := strings.Fields(clause)
	if len(fields) == 0 {
		return "", "", nil
	}
	method, result = splitPair(fields[0])
	method = strings.ToLower(method)
	// a result can carry a parenthesised comment, e.g. "pass (good signature)".
	result = strings.ToLower(strings.TrimSpace(strings.SplitN(result, "(", 2)[0]))

	props = make(map[string]string, len(fields)-1)
	for _, field := range fields[1:] {
		key, value := splitPair(field)
		key = strings.ToLower(strings.TrimSpace(key))
		if key != "" {
			props[key] = strings.Trim(strings.TrimSpace(value), `"`)
		}
	}
	return method, result, props
}

// splitPair splits "key=value" on the first equals sign.
func splitPair(s string) (string, string) {
	key, value, found := strings.Cut(strings.TrimSpace(s), "=")
	if !found {
		return strings.TrimSpace(key), ""
	}
	return strings.TrimSpace(key), strings.TrimSpace(value)
}

// firstOf returns the first property present out of keys.
func firstOf(props map[string]string, keys ...string) string {
	for _, key := range keys {
		if v, ok := props[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

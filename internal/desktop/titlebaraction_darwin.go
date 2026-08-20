//go:build darwin

package desktop

/*
#cgo LDFLAGS: -framework CoreFoundation
#include <stdlib.h>
#include <CoreFoundation/CoreFoundation.h>

// peltonDoubleClickAction copies the global AppleActionOnDoubleClick preference
// into buf, returning 1 on success and 0 when the key is unset or not a string.
// The setting lives in the global domain, hence kCFPreferencesAnyApplication.
static int peltonDoubleClickAction(char *buf, int len) {
    CFPropertyListRef v = CFPreferencesCopyAppValue(CFSTR("AppleActionOnDoubleClick"),
                                                    kCFPreferencesAnyApplication);
    if (v == NULL) {
        return 0;
    }
    int ok = 0;
    if (CFGetTypeID(v) == CFStringGetTypeID()) {
        ok = CFStringGetCString((CFStringRef)v, buf, len, kCFStringEncodingUTF8) ? 1 : 0;
    }
    CFRelease(v);
    return ok;
}
*/
import "C"

import "strings"

// titleBarDoubleClickAction reports what System Settings' "double-click a
// window's title bar to" is set to, lowercased: "maximize", "minimize" or
// "none". macOS leaves the key unset until it is changed, and unset means
// Maximize.
func titleBarDoubleClickAction() string {
	var buf [32]C.char
	if C.peltonDoubleClickAction(&buf[0], C.int(len(buf))) == 0 {
		return "maximize"
	}
	return strings.ToLower(C.GoString(&buf[0]))
}

//go:build darwin

package desktop

/*
// the preamble is Objective-C, not C: cgo compiles it as C unless told
// otherwise, and AppKit's headers do not parse that way.
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#import <Cocoa/Cocoa.h>

// peltonSetDockBadge writes label onto the dock tile, or clears it when label
// is empty. AppKit is main-thread only and bound methods run on their own
// goroutine, hence the hop onto the main queue.
static void peltonSetDockBadge(const char *label) {
    NSString *text = label == NULL || label[0] == '\0'
        ? nil
        : [NSString stringWithUTF8String:label];
    dispatch_async(dispatch_get_main_queue(), ^{
        [[NSApp dockTile] setBadgeLabel:text];
    });
}
*/
import "C"

import (
	"strconv"
	"unsafe"
)

// badgeCap is where the badge stops counting. Past it the dock shows "999+",
// the way Mail does, rather than a number too wide for the tile.
const badgeCap = 999

// setPlatformBadge puts the unread count on the dock tile. Zero clears it.
func setPlatformBadge(count int) {
	label := ""
	switch {
	case count > badgeCap:
		label = strconv.Itoa(badgeCap) + "+"
	case count > 0:
		label = strconv.Itoa(count)
	}
	c := C.CString(label)
	defer C.free(unsafe.Pointer(c))
	C.peltonSetDockBadge(c)
}

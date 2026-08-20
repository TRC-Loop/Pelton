package desktop

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework UserNotifications
#include <stdbool.h>
#include <stdlib.h>
#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>

// peltonCanNotify reports whether this process is running from an app bundle.
// UNUserNotificationCenter requires one: currentNotificationCenter raises an
// exception in a process with no bundle identifier, which would take the app
// down rather than fail to notify. A `go run` or an unbundled dev build has
// none, so the caller falls back to beeep there.
static bool peltonCanNotify(void) {
	// UserNotifications arrived in 10.14 and the deployment target is older, so
	// every use of it is inside an @available guard. On anything older the
	// framework is simply absent and the caller keeps the old path.
	if (@available(macOS 10.14, *)) {
		return [[NSBundle mainBundle] bundleIdentifier] != nil;
	}
	return false;
}

// peltonNotify posts one notification through UserNotifications.
//
// The point of using this over an osascript "display notification" is the
// icon: macOS shows the icon of whichever process posted, so going through
// osascript brands every new-mail alert as Script Editor. Posting from inside
// the bundle makes macOS use the bundle's own icon, with no asset to ship.
//
// Authorization is requested on every call. After the first time it resolves
// from the stored answer without prompting again, so this costs nothing and
// means the prompt appears the first time a notification would actually be
// shown rather than at launch.
static void peltonNotify(const char *title, const char *body) {
	if (@available(macOS 10.14, *)) {
		NSString *t = [NSString stringWithUTF8String:title];
		NSString *b = [NSString stringWithUTF8String:body];
		UNUserNotificationCenter *center = [UNUserNotificationCenter currentNotificationCenter];

		[center requestAuthorizationWithOptions:UNAuthorizationOptionAlert
		                      completionHandler:^(BOOL granted, NSError *error) {
			if (!granted) {
				// the only signal there is: delivery is asynchronous and the Go
				// caller has already returned. Without this a user whose
				// notifications never arrive has nothing at all to look at.
				NSLog(@"pelton: notification authorization denied: %@", error);
				return;
			}
			UNMutableNotificationContent *content = [[UNMutableNotificationContent alloc] init];
			content.title = t;
			content.body = b;
			// no sound: this change is about the icon, and what Pelton should make
			// a noise about is its own question (#240).
			UNNotificationRequest *request =
			    [UNNotificationRequest requestWithIdentifier:[[NSUUID UUID] UUIDString]
			                                         content:content
			                                         trigger:nil];
			[center addNotificationRequest:request
			         withCompletionHandler:^(NSError *addError) {
				if (addError != nil) {
					NSLog(@"pelton: notification not posted: %@", addError);
				}
			}];
			[content release];
		}];
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/gen2brain/beeep"
)

// deliverNotification raises one notification. On macOS this goes through
// UserNotifications rather than beeep, which falls back to osascript when
// terminal-notifier is absent and so shows the Script Editor automation icon
// instead of the Pelton logo (#143). Posting from the app bundle makes macOS
// use the bundle icon on its own.
//
// A process with no bundle (a dev run, a cli tool) cannot use the framework at
// all, and neither can macOS before 10.14, where it does not exist. Both keep
// the old path and so keep the wrong icon. Delivery is asynchronous once handed over, so
// a rejected authorization or a failed post is not reported back here; a
// notification is not worth failing a sync over.
func deliverNotification(title, body string) error {
	if !bool(C.peltonCanNotify()) {
		return beeep.Notify(title, body, "")
	}

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cBody := C.CString(body)
	defer C.free(unsafe.Pointer(cBody))

	C.peltonNotify(cTitle, cBody)
	return nil
}

//go:build !linux

package desktop

// detectSystemColorScheme returns "" on macOS and Windows, where the frontend's
// prefers-color-scheme media query already reflects the OS setting; the caller
// falls back to that. Only Linux (WebKitGTK) needs the native portal query.
func detectSystemColorScheme() string {
	return ""
}

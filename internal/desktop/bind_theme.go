package desktop

// SystemColorScheme reports the operating system's dark/light preference as
// "dark" or "light", or "" when it cannot be determined. It exists because
// WebKitGTK on Linux does not surface the desktop color-scheme to CSS's
// prefers-color-scheme media query, so the frontend cannot detect dark mode on
// its own there; it consults this and falls back to the media query (which is
// correct on macOS and Windows) whenever this returns "".
func (a *App) SystemColorScheme() string {
	return detectSystemColorScheme()
}

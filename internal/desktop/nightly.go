package desktop

import "github.com/TRC-Loop/Pelton/internal/storage"

// Nightly builds are cut automatically from the dev branch by
// .github/workflows/nightly.yml. They are untested, unreviewed and expected to
// break, so they behave differently in two ways:
//
//   - they open their own data directory (storage.DefaultPathForChannel), so a
//     nightly cannot damage a stable install's accounts, cache or settings, and
//   - the frontend blocks on an acknowledgement dialog at every launch and keeps
//     a marker in the status bar for as long as the app is open.
//
// The channel is set at build time with -ldflags "-X main.channel=nightly"; a
// normal build has an empty channel and none of this applies.

// IsNightly reports whether this is an automated nightly build. The frontend
// reads it once at startup to decide whether to show the launch warning and the
// status bar marker.
func (a *App) IsNightly() bool {
	return a.channel == storage.ChannelNightly
}

// aboutMessage is the body of the native macOS About panel. A nightly says so
// there too, since that panel is reachable without the app's own settings.
func aboutMessage(channel, version string) string {
	msg := "An open-source desktop mail client.\nVersion " + version
	if channel == storage.ChannelNightly {
		msg += "\n\nUntested nightly build. Do not use it with your real inbox."
	}
	return msg
}

//go:build linux

package desktop

import "github.com/godbus/dbus/v5"

// detectSystemColorScheme asks the XDG desktop portal for the freedesktop
// "color-scheme" appearance setting, the same signal browsers use to resolve
// prefers-color-scheme on Linux. It returns "dark", "light", or "" when the
// preference is unset or the portal is unavailable. The session bus connection
// is shared and must not be closed here.
func detectSystemColorScheme() string {
	conn, err := dbus.SessionBus()
	if err != nil {
		return ""
	}
	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	var v dbus.Variant
	err = obj.Call("org.freedesktop.portal.Settings.ReadOne", 0, "org.freedesktop.appearance", "color-scheme").Store(&v)
	if err != nil {
		// older portals only expose Read, which wraps the value in an extra variant.
		if err = obj.Call("org.freedesktop.portal.Settings.Read", 0, "org.freedesktop.appearance", "color-scheme").Store(&v); err != nil {
			return ""
		}
	}

	// unwrap any nested variants down to the concrete value.
	for {
		inner, ok := v.Value().(dbus.Variant)
		if !ok {
			break
		}
		v = inner
	}

	// 0 = no preference, 1 = prefer dark, 2 = prefer light.
	switch scheme, _ := v.Value().(uint32); scheme {
	case 1:
		return "dark"
	case 2:
		return "light"
	default:
		return ""
	}
}

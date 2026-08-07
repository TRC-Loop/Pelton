//go:build windows

package desktop

import "testing"

func TestMinimizeToTray(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		want    bool
	}{
		{name: "unset minimizes normally", want: false},
		{name: "minimize stays on the taskbar", setting: minimizeActionNormal},
		{name: "tray hides the window", setting: minimizeActionTray, want: true},
		{name: "unknown value falls back to normal", setting: "nonsense"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newCloseTestApp(t)
			if tt.setting != "" {
				if err := a.store.Set(a.ctx, settingMinimizeAction, tt.setting); err != nil {
					t.Fatalf("set: %v", err)
				}
			}
			if got := a.minimizeToTray(); got != tt.want {
				t.Errorf("minimizeToTray() = %v, want %v", got, tt.want)
			}
		})
	}
}

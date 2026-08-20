package desktop

import "testing"

func TestFitGeometry(t *testing.T) {
	tests := []struct {
		name             string
		in               geometry
		screenW, screenH int
		want             geometry
	}{
		{
			name:    "a frame that fits is left alone",
			in:      geometry{Width: 1000, Height: 700, X: 100, Y: 80, Placed: true},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: 1000, Height: 700, X: 100, Y: 80, Placed: true},
		},
		{
			name:    "a window larger than the screen shrinks to it",
			in:      geometry{Width: 3000, Height: 2000, X: 0, Y: 0, Placed: true},
			screenW: 1440, screenH: 900,
			want: geometry{Width: 1440, Height: 900, X: 0, Y: 0, Placed: true},
		},
		{
			name:    "a window smaller than the minimum grows to it",
			in:      geometry{Width: 200, Height: 100},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: minWindowWidth, Height: minWindowHeight},
		},
		{
			// the unplugged-second-monitor case: a position further right than
			// this screen goes cannot be verified, so it is dropped.
			name:    "a position off the right of the screen is dropped",
			in:      geometry{Width: 1000, Height: 700, X: 2400, Y: 100, Placed: true},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: 1000, Height: 700, X: 2400, Y: 100, Placed: false},
		},
		{
			name:    "a negative position is dropped",
			in:      geometry{Width: 1000, Height: 700, X: -1200, Y: 40, Placed: true},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: 1000, Height: 700, X: -1200, Y: 40, Placed: false},
		},
		{
			name:    "a position too low to leave the title bar grabbable is dropped",
			in:      geometry{Width: 1000, Height: 700, X: 100, Y: 1040, Placed: true},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: 1000, Height: 700, X: 100, Y: 1040, Placed: false},
		},
		{
			name:    "an unreadable screen keeps the size and drops the position",
			in:      geometry{Width: 1100, Height: 700, X: 10, Y: 10, Placed: true},
			screenW: 0, screenH: 0,
			want: geometry{Width: 1100, Height: 700, X: 10, Y: 10, Placed: false},
		},
		{
			// a screen smaller than the app's own minimum: the minimum wins, so
			// the window stays usable even if it overflows.
			name:    "a screen below the minimum size still yields the minimum",
			in:      geometry{Width: 1000, Height: 700},
			screenW: 640, screenH: 480,
			want: geometry{Width: minWindowWidth, Height: minWindowHeight},
		},
		{
			name:    "the maximised flag survives clamping",
			in:      geometry{Width: 1000, Height: 700, Maximised: true},
			screenW: 1920, screenH: 1080,
			want: geometry{Width: 1000, Height: 700, Maximised: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fitGeometry(tt.in, tt.screenW, tt.screenH)
			if got != tt.want {
				t.Errorf("fitGeometry() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

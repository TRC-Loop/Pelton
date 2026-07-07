package desktop

import (
	"errors"
	"testing"
	"time"
)

func TestResolveNotBefore(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		sendAt       string
		delaySeconds int
		wantZero     bool
		wantAt       time.Time
		wantErr      error
	}{
		{
			name:         "no sendAt, no delay sends immediately",
			sendAt:       "",
			delaySeconds: 0,
			wantZero:     true,
		},
		{
			name:         "no sendAt, undo delay applies",
			sendAt:       "",
			delaySeconds: 10,
			wantAt:       now.Add(10 * time.Second),
		},
		{
			name:         "future sendAt schedules at that time",
			sendAt:       "2026-07-05T08:00:00Z",
			delaySeconds: 10,
			wantAt:       time.Date(2026, 7, 5, 8, 0, 0, 0, time.UTC),
		},
		{
			name:         "sendAt takes precedence over undo delay",
			sendAt:       "2026-07-04T12:00:05Z",
			delaySeconds: 30,
			wantAt:       time.Date(2026, 7, 4, 12, 0, 5, 0, time.UTC),
		},
		{
			name:    "past sendAt is rejected",
			sendAt:  "2026-07-04T11:00:00Z",
			wantErr: ErrSendAtPast,
		},
		{
			name:    "sendAt equal to now is rejected",
			sendAt:  "2026-07-04T12:00:00Z",
			wantErr: ErrSendAtPast,
		},
		{
			name:    "invalid sendAt format is rejected",
			sendAt:  "tomorrow morning",
			wantErr: ErrSendAtInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveNotBefore(tc.sendAt, tc.delaySeconds, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("resolveNotBefore() err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveNotBefore() unexpected err: %v", err)
			}
			if tc.wantZero {
				if !got.IsZero() {
					t.Fatalf("resolveNotBefore() = %v, want zero time", got)
				}
				return
			}
			if !got.Equal(tc.wantAt) {
				t.Fatalf("resolveNotBefore() = %v, want %v", got, tc.wantAt)
			}
		})
	}
}

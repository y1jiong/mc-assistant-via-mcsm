package service

import (
	"testing"
	"time"
)

func TestTicksPerInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dayMinutes int
		interval   time.Duration
		want       int
		wantError  bool
	}{
		{name: "vanilla day", dayMinutes: 20, interval: 550 * time.Millisecond, want: 11},
		{name: "custom delay", dayMinutes: 20, interval: time.Second, want: 20},
		{name: "invalid day", dayMinutes: 0, interval: time.Second, wantError: true},
		{name: "invalid interval", dayMinutes: 20, interval: 0, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := ticksPerInterval(test.dayMinutes, test.interval)
			if (err != nil) != test.wantError {
				t.Fatalf("ticksPerInterval() error = %v, wantError %v", err, test.wantError)
			}
			if got != test.want {
				t.Fatalf("ticksPerInterval() = %d, want %d", got, test.want)
			}
		})
	}
}

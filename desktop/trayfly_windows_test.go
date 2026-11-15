//go:build windows

package main

import (
	"testing"
	"time"
)

func TestFlyDuration(t *testing.T) {
	old := flyMillis
	t.Cleanup(func() { flyMillis = old })
	for in, want := range map[string]time.Duration{
		"220":  220 * time.Millisecond,
		"2000": 2 * time.Second,
		"0":    220 * time.Millisecond, // nonsense falls back to the default
		"-5":   220 * time.Millisecond,
		"x":    220 * time.Millisecond,
		"9999": 220 * time.Millisecond, // never longer than 5 s
	} {
		flyMillis = in
		if got := flyDuration(); got != want {
			t.Errorf("flyDuration(%q) = %v, want %v", in, got, want)
		}
	}
}

// Without a Wails window (as in tests) the animation does nothing and returns a usable
// restore function.
func TestFlyToTrayWithoutWindow(t *testing.T) {
	if mainWindow() != 0 {
		t.Skip("a Peregrine window exists in this process")
	}
	start := time.Now()
	restore := flyToTray()
	if restore == nil {
		t.Fatal("restore is nil")
	}
	restore()
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("flyToTray waited although there was no window")
	}
}

func TestTrayPointOnScreen(t *testing.T) {
	x, y := trayPoint()
	if x <= 0 || y <= 0 {
		t.Fatalf("tray point (%d,%d) is off screen", x, y)
	}
}

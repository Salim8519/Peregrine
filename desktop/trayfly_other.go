//go:build !windows

package main

// flyToTray: no animation outside Windows yet (macOS hides the app natively).
func flyToTray() (restore func()) { return func() {} }

//go:build !windows

package main

// canShowBehind: macOS has no reliable "show without activating" for the Wails window, so a
// link while the app is open opens the dialog inside it, as before (ADR 0021).
const canShowBehind = false

func showWindowBehind(bool) {}

func rememberFrontWindow() {}

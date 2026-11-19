//go:build !darwin

package main

import (
	"context"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// watchMacApp is macOS-only (macapp_darwin.go): Windows and Linux learn the window state
// from beforeClose and the tray.
func watchMacApp(func(hidden bool), func()) {}

// hideToTray hides the window; the tray icon brings it back.
func hideToTray(ctx context.Context) { wruntime.WindowHide(ctx) }

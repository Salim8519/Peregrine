package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
void peregrineWatchApp(void);
*/
import "C"

import (
	"context"
	"sync"
	"sync/atomic"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// On macOS "hidden in the menu bar" means the whole app is hidden (NSApp hide), which is
// what the close button and Cmd+H do. macapp_darwin.m reports hide/unhide to Go, and a
// Dock click with no window on screen (after launch at login, for example) shows the
// window, because Wails has no reopen handler (ADR 0025).

var macHooks struct {
	mu      sync.Mutex // applies hide/unhide in order (see peregrineAppHidden)
	hidden  atomic.Bool
	onHide  func(hidden bool)
	reopen  func()
	started atomic.Bool
}

// watchMacApp installs the hooks once. Both callbacks run on a goroutine, never on the
// main thread, so they may call the Wails runtime.
func watchMacApp(onHide func(hidden bool), reopen func()) {
	if !macHooks.started.CompareAndSwap(false, true) {
		return
	}
	macHooks.onHide, macHooks.reopen = onHide, reopen
	C.peregrineWatchApp()
}

//export peregrineAppHidden
func peregrineAppHidden(hidden C.int) {
	macHooks.hidden.Store(hidden != 0)
	go func() {
		// The last goroutine to take the lock reads the newest state, so a quick
		// hide-unhide never ends as "hidden".
		macHooks.mu.Lock()
		defer macHooks.mu.Unlock()
		if macHooks.onHide != nil {
			macHooks.onHide(macHooks.hidden.Load())
		}
	}()
}

//export peregrineReopen
func peregrineReopen() {
	if macHooks.reopen != nil {
		go macHooks.reopen()
	}
}

// hideToTray hides the app like its close button does, so a Dock click brings it back.
// (Hiding only the window would leave a Dock icon that shows nothing.)
func hideToTray(ctx context.Context) { wruntime.Hide(ctx) }

//go:build windows

package main

import (
	"os"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

// canShowBehind: Windows can show a window below the others without activating it, so a
// link from the browser while the app is open behind it uses the popup too (ADR 0021).
const canShowBehind = true

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	procFindWindowEx = user32.NewProc("FindWindowExW")
	procShowWindow   = user32.NewProc("ShowWindow")
	procSetWindowPos = user32.NewProc("SetWindowPos")
	procGetFG        = user32.NewProc("GetForegroundWindow")
	procSetFG        = user32.NewProc("SetForegroundWindow")
	procIsWindow     = user32.NewProc("IsWindow")

	frontWindow atomic.Uintptr // the window in front when the popup opened (the browser)
)

// rememberFrontWindow records the foreground window (not ours: the popup is not shown yet).
func rememberFrontWindow() {
	h, _, _ := procGetFG.Call()
	if h != 0 && windows.HWND(h) == mainWindow() {
		h = 0
	}
	frontWindow.Store(h)
}

const (
	swShowMaximized  = 3
	swShowNoActivate = 4
	swpNoSize        = 0x0001
	swpNoMove        = 0x0002
	swpNoActivate    = 0x0010
	hwndBottom       = 1
)

// mainWindow finds this process's Wails window (class "wailsWindow").
func mainWindow() windows.HWND {
	class, _ := windows.UTF16PtrFromString("wailsWindow")
	pid := uint32(os.Getpid())
	var h uintptr
	for i := 0; i < 64; i++ {
		h, _, _ = procFindWindowEx.Call(0, h, uintptr(unsafe.Pointer(class)), 0)
		if h == 0 {
			return 0
		}
		var owner uint32
		if _, err := windows.GetWindowThreadProcessId(windows.HWND(h), &owner); err == nil && owner == pid {
			return windows.HWND(h)
		}
	}
	return 0
}

// showWindowBehind shows the hidden main window at the bottom of the z-order without
// activating it, and gives the focus back to the window that was in front (the browser).
// Hiding the popup does not reliably hand the focus back by itself.
func showWindowBehind(maximised bool) {
	h := mainWindow()
	if h == 0 {
		return
	}
	cmd := uintptr(swShowNoActivate)
	if maximised {
		cmd = swShowMaximized // there is no "maximised, not activated"; sent to the bottom next
	}
	procShowWindow.Call(uintptr(h), cmd)
	procSetWindowPos.Call(uintptr(h), hwndBottom, 0, 0, 0, 0, swpNoSize|swpNoMove|swpNoActivate)
	if f := frontWindow.Swap(0); f != 0 {
		if ok, _, _ := procIsWindow.Call(f); ok != 0 {
			procSetFG.Call(f) // allowed: our popup was the foreground window
		}
	}
}

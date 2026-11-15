//go:build windows

package main

import (
	"os"
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Fly-to-tray animation: before the window hides in the tray it slides toward the
// notification area and fades out, so the user sees where Peregrine went. It is short
// (flyMillis), skipped when Windows animations are turned off, and fully undone after
// the window is hidden (position, opacity, window style).

var (
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procGetWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	procSetLayeredAttrs  = user32.NewProc("SetLayeredWindowAttributes")
	procSystemParamsInfo = user32.NewProc("SystemParametersInfoW")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procIsWindowVisible  = user32.NewProc("IsWindowVisible")
	procIsIconic         = user32.NewProc("IsIconic")
	procShellIconRect    = windows.NewLazySystemDLL("shell32.dll").NewProc("Shell_NotifyIconGetRect")
	procIsZoomed         = user32.NewProc("IsZoomed")
)

// flyMillis is the animation length. It is a string variable so a manual test build can
// slow it down to watch it: go build -ldflags "-X main.flyMillis=2000".
var flyMillis = "220"

const (
	gwlExStyle           = -20 // GWL_EXSTYLE
	wsExLayered          = 0x00080000
	lwaAlpha             = 0x2
	spiGetClientAreaAnim = 0x1042
	smCxScreen           = 0
	smCyScreen           = 1
	swpNoZOrder          = 0x0004
	flyFrameInterval     = 15 * time.Millisecond
)

type rect struct{ left, top, right, bottom int32 }

// animationsOn reports Windows' "Animation effects" setting (Accessibility > Visual effects).
func animationsOn() bool {
	var on int32 = 1
	if r, _, _ := procSystemParamsInfo.Call(spiGetClientAreaAnim, 0, uintptr(unsafe.Pointer(&on)), 0); r == 0 {
		return true
	}
	return on != 0
}

// Our tray icon, as created by fyne.io/systray (pinned in go.mod): a hidden window of class
// "SystrayClass" in this process, icon ID 100. If a library update changes either, the
// lookup fails and the animation falls back to the notification area.
const (
	systrayClass  = "SystrayClass"
	systrayIconID = 100
)

// notifyIconIdentifier is NOTIFYICONIDENTIFIER.
type notifyIconIdentifier struct {
	size uint32
	hwnd uintptr
	id   uint32
	guid windows.GUID
}

// ownWindow finds a top-level window of the given class that belongs to this process.
func ownWindow(class string) uintptr {
	c, _ := windows.UTF16PtrFromString(class)
	pid := uint32(os.Getpid())
	var h uintptr
	for i := 0; i < 64; i++ {
		h, _, _ = procFindWindowEx.Call(0, h, uintptr(unsafe.Pointer(c)), 0)
		if h == 0 {
			return 0
		}
		var owner uint32
		if _, err := windows.GetWindowThreadProcessId(windows.HWND(h), &owner); err == nil && owner == pid {
			return h
		}
	}
	return 0
}

func centre(r rect) (int32, int32) { return (r.left + r.right) / 2, (r.top + r.bottom) / 2 }

// trayPoint is where the window flies to, wherever the taskbar is (bottom, top, left,
// right, any monitor): the centre of Peregrine's own tray icon when Windows can tell
// (Shell_NotifyIconGetRect), else of the notification area, else of the taskbar, else the
// bottom-right corner of the primary screen.
func trayPoint() (int32, int32) {
	if tw := ownWindow(systrayClass); tw != 0 {
		id := notifyIconIdentifier{hwnd: tw, id: systrayIconID}
		id.size = uint32(unsafe.Sizeof(id))
		var r rect
		if hr, _, _ := procShellIconRect.Call(uintptr(unsafe.Pointer(&id)), uintptr(unsafe.Pointer(&r))); hr == 0 && r.right > r.left {
			return centre(r)
		}
	}
	shell, _ := windows.UTF16PtrFromString("Shell_TrayWnd")
	notify, _ := windows.UTF16PtrFromString("TrayNotifyWnd")
	if bar, _, _ := procFindWindowEx.Call(0, 0, uintptr(unsafe.Pointer(shell)), 0); bar != 0 {
		var r rect
		if n, _, _ := procFindWindowEx.Call(bar, 0, uintptr(unsafe.Pointer(notify)), 0); n != 0 {
			if ok, _, _ := procGetWindowRect.Call(n, uintptr(unsafe.Pointer(&r))); ok != 0 && r.right > r.left {
				return centre(r)
			}
		}
		if ok, _, _ := procGetWindowRect.Call(bar, uintptr(unsafe.Pointer(&r))); ok != 0 && r.right > r.left {
			return centre(r)
		}
	}
	w, _, _ := procGetSystemMetrics.Call(smCxScreen)
	h, _, _ := procGetSystemMetrics.Call(smCyScreen)
	return int32(w) - 24, int32(h) - 24
}

func flyDuration() time.Duration {
	ms, err := strconv.Atoi(flyMillis)
	if err != nil || ms <= 0 || ms > 5000 {
		ms = 220
	}
	return time.Duration(ms) * time.Millisecond
}

// flyToTray slides the visible main window toward the tray while fading it out. It returns
// a function that restores the window's position and opacity; call it after hiding the
// window so the user never sees the restore. Both are no-ops when there is nothing to
// animate (no window, already hidden, minimised, maximised, animations off).
func flyToTray() (restore func()) {
	restore = func() {}
	h := uintptr(mainWindow())
	if h == 0 || !animationsOn() {
		return restore
	}
	if v, _, _ := procIsWindowVisible.Call(h); v == 0 {
		return restore
	}
	if ic, _, _ := procIsIconic.Call(h); ic != 0 {
		return restore
	}
	if z, _, _ := procIsZoomed.Call(h); z != 0 {
		return restore // a maximised window cannot be moved without un-maximising it
	}
	var start rect
	if ok, _, _ := procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&start))); ok == 0 {
		return restore
	}
	gwl := int32(gwlExStyle)
	ex, _, _ := procGetWindowLongPtr.Call(h, uintptr(gwl))
	layered := ex&wsExLayered != 0
	if !layered {
		procSetWindowLongPtr.Call(h, uintptr(gwl), ex|wsExLayered)
	}
	restore = func() {
		procSetWindowPos.Call(h, 0, uintptr(start.left), uintptr(start.top), 0, 0, swpNoSize|swpNoZOrder|swpNoActivate)
		procSetLayeredAttrs.Call(h, 0, 255, lwaAlpha)
		if !layered {
			procSetWindowLongPtr.Call(h, uintptr(gwl), ex)
		}
	}

	tx, ty := trayPoint()
	w, hgt := start.right-start.left, start.bottom-start.top
	// Move the window's centre onto the tray point.
	dx := float64(tx - w/2 - start.left)
	dy := float64(ty - hgt/2 - start.top)
	d := flyDuration()
	began := time.Now()
	for {
		t := float64(time.Since(began)) / float64(d)
		if t > 1 {
			t = 1
		}
		e := t * t // ease-in: starts gently, then speeds into the tray
		x := start.left + int32(dx*e)
		y := start.top + int32(dy*e)
		alpha := uintptr(255 * (1 - e))
		procSetWindowPos.Call(h, 0, uintptr(x), uintptr(y), 0, 0, swpNoSize|swpNoZOrder|swpNoActivate)
		procSetLayeredAttrs.Call(h, 0, alpha, lwaAlpha)
		if t >= 1 {
			break
		}
		time.Sleep(flyFrameInterval)
	}
	return restore
}

package main

import (
	"slices"
	"sync"
	"testing"

	"github.com/Salim8519/Peregrine/internal/api"
)

type fakeWindow struct {
	mu       sync.Mutex
	min, max bool
	visible  bool
	w, h     int
	x, y     int
	minW     int
	minH     int
	onTop    bool
	events   []string
	centered bool
	behind   bool
}

func (f *fakeWindow) isMinimised() bool    { return f.min }
func (f *fakeWindow) isMaximised() bool    { return f.max }
func (f *fakeWindow) size() (int, int)     { return f.w, f.h }
func (f *fakeWindow) position() (int, int) { return f.x, f.y }
func (f *fakeWindow) setSize(w, h int)     { f.w, f.h = w, h }
func (f *fakeWindow) setMinSize(w, h int)  { f.minW, f.minH = w, h }
func (f *fakeWindow) setPosition(x, y int) { f.x, f.y = x, y }
func (f *fakeWindow) center()              { f.centered = true }
func (f *fakeWindow) unmaximise()          { f.max = false }
func (f *fakeWindow) maximise()            { f.max = true }
func (f *fakeWindow) minimise()            { f.min, f.visible = true, false }
func (f *fakeWindow) show()                { f.min, f.visible = false, true }
func (f *fakeWindow) hide()                { f.visible = false }
func (f *fakeWindow) flyToTray() func() {
	ev := "fly"
	if !f.visible {
		ev = "fly-while-hidden" // the animation must run before the window hides
	}
	f.events = append(f.events, ev)
	return func() {
		ev := "unfly"
		if f.visible {
			ev = "unfly-while-visible" // the restore must run after the window hides
		}
		f.events = append(f.events, ev)
	}
}
func (f *fakeWindow) setOnTop(on bool) { f.onTop = on }
func (f *fakeWindow) rememberFront()   { f.events = append(f.events, "front") }
func (f *fakeWindow) showBehind(max bool) {
	f.visible, f.behind, f.max = true, true, max
}
func (f *fakeWindow) emitPopup(mode string) {
	f.mu.Lock()
	f.events = append(f.events, mode)
	f.mu.Unlock()
}

func TestPopupNotUsedWhenWindowVisible(t *testing.T) {
	var p popup
	w := &fakeWindow{visible: true, w: 1100, h: 700}
	if p.open(w, false, true) {
		t.Fatal("popup opened over a visible window")
	}
	if w.w != 1100 || len(w.events) != 0 {
		t.Fatalf("window changed: %+v", w)
	}
}

func TestPopupFromTrayAndBack(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1100, h: 700, x: 40, y: 50}
	if !p.open(w, true, true) {
		t.Fatal("popup did not open from the tray")
	}
	if !w.visible || !w.onTop || !w.centered || w.w != popupWidth || w.h != popupHeight || w.minW != popupWidth {
		t.Fatalf("popup shape wrong: %+v", w)
	}
	if !p.close(w, false) {
		t.Fatal("close reported inactive")
	}
	if w.visible || w.onTop || w.w != 1100 || w.h != 700 || w.x != 40 || w.y != 50 || w.minW != mainMinW || w.minH != mainMinH {
		t.Fatalf("window not restored: %+v", w)
	}
	// Going back to the tray flies the window there, then undoes the animation once hidden.
	if !slices.Equal(w.events, []string{"open", "fly", "unfly", "closed"}) {
		t.Fatalf("events = %v", w.events)
	}
	if p.close(w, false) {
		t.Fatal("second close reported active")
	}
}

func TestPopupFromMinimisedGoesBackToTaskbar(t *testing.T) {
	var p popup
	w := &fakeWindow{min: true, max: true, w: 1100, h: 700}
	if !p.open(w, false, true) {
		t.Fatal("popup did not open from a minimised window")
	}
	if w.max {
		t.Fatal("popup stayed maximised")
	}
	p.close(w, false)
	if !w.min || !w.max || w.visible {
		t.Fatalf("window not minimised+maximised again: %+v", w)
	}
	if slices.Contains(w.events, "fly") {
		t.Fatalf("a minimised window flew to the tray: %v", w.events)
	}
}

func TestPopupExpandKeepsWindowOnScreen(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1000, h: 650}
	p.open(w, true, true)
	p.close(w, true)
	if !w.visible || w.w != 1000 || w.h != 650 || w.onTop {
		t.Fatalf("expanded window wrong: %+v", w)
	}
	if !slices.Equal(w.events, []string{"open", "expanded"}) {
		t.Fatalf("events = %v", w.events)
	}
}

func TestPopupSecondCaptureReusesPopup(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1100, h: 700}
	p.open(w, true, true)
	w.visible = false
	if !p.open(w, true, true) || !w.visible {
		t.Fatal("second capture did not bring the popup forward")
	}
	p.close(w, false)
	if w.w != 1100 || w.h != 700 {
		t.Fatalf("restored to the popup size instead of the app size: %+v", w)
	}
}

func TestPopupFit(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1100, h: 700}
	p.fit(w, 300, 290) // inactive: ignored
	if w.h != 700 {
		t.Fatal("fit changed an inactive window")
	}
	p.open(w, true, true)
	// Title bar 32 px: inner 298 of 330. Content 400 tall -> 432.
	p.fit(w, 400, popupHeight-32)
	if w.h != 432 {
		t.Fatalf("h = %d, want 432", w.h)
	}
	p.fit(w, 5000, 400)
	if w.h != popupMaxH {
		t.Fatalf("h = %d, want max %d", w.h, popupMaxH)
	}
	p.fit(w, 10, 700)
	if w.h != popupMinH {
		t.Fatalf("h = %d, want min %d", w.h, popupMinH)
	}
}

func TestPopupKeepsPendingCapture(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1100, h: 700}
	p.captured(&api.Capture{Source: "extension"}) // no popup: not kept
	if p.pendingCapture() != nil {
		t.Fatal("kept a capture without a popup")
	}
	p.open(w, true, true)
	c := &api.Capture{Source: "extension"}
	c.URL = "https://x.test/a.zip"
	p.captured(c)
	if got := p.pendingCapture(); got == nil || got.URL != c.URL {
		t.Fatalf("pending = %+v", got)
	}
	p.close(w, false)
	if p.pendingCapture() != nil {
		t.Fatal("pending kept after close")
	}
}

func TestPopupWhenAppBehindBrowser(t *testing.T) {
	var p popup
	w := &fakeWindow{visible: true, max: true, w: 1100, h: 700, x: 10, y: 20}
	if p.open(w, false, false) != canShowBehind {
		t.Fatalf("open behind = %v, want %v", !canShowBehind, canShowBehind)
	}
	if !canShowBehind {
		return
	}
	if w.max || w.w != popupWidth || !w.onTop {
		t.Fatalf("popup shape wrong: %+v", w)
	}
	p.close(w, false)
	if !w.visible || !w.behind || !w.max || w.w != 1100 || w.h != 700 || w.x != 10 || w.y != 20 || w.onTop {
		t.Fatalf("window not put back behind: %+v", w)
	}
	if p.wasFrom() != fromBehind {
		t.Fatal("from not recorded")
	}
}

func TestPopupNotUsedWhenUserInApp(t *testing.T) {
	var p popup
	w := &fakeWindow{visible: true, w: 1100, h: 700}
	if p.open(w, false, true) {
		t.Fatal("popup opened while the user was in the app")
	}
}

// ADR 0023: the "download complete" popup uses the same window. It never replaces a
// quick-add popup that is waiting for the user; a captured link replaces it.
func TestDonePopupAndQuickAdd(t *testing.T) {
	var p popup
	w := &fakeWindow{w: 1100, h: 700}
	if !p.openAs(w, kindDone, true, true) || !p.shows(kindDone) {
		t.Fatal("done popup did not open from the tray")
	}
	if !p.openAs(w, kindDone, true, true) {
		t.Fatal("a second finished download did not reuse the done popup")
	}
	if !p.open(w, true, true) || !p.shows(kindAdd) {
		t.Fatal("a captured link did not take over the done popup")
	}
	if p.openAs(w, kindDone, true, true) || !p.shows(kindAdd) {
		t.Fatal("a finished download replaced the quick-add popup")
	}
	p.close(w, false)
	if p.isActive() || w.w != 1100 || w.h != 700 {
		t.Fatalf("window not restored: %+v", w)
	}
	// In the app: no popup (the dialog shows inside the app).
	w = &fakeWindow{visible: true, w: 1100, h: 700}
	if p.openAs(w, kindDone, false, true) {
		t.Fatal("done popup opened while the user is in the app")
	}
}

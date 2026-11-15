package main

import (
	"context"
	"sync"

	"github.com/Salim8519/Peregrine/internal/api"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Quick-add popup (ADR 0021): a link captured from the browser while the window is hidden
// in the tray (or minimised) shows only a small add dialog, not the whole app. Wails v2 has
// one window, so the popup is that window shrunk, centred and kept on top; finishing the
// dialog puts it back the way it was and hides it again.

const (
	popupWidth  = 520
	popupHeight = 330 // first guess; the UI fits the height to its content (PopupFit)
	popupMinH   = 200
	popupMaxH   = 760
	mainMinW    = 820 // keep in sync with main.go
	mainMinH    = 520
)

// window is the part of the Wails window API the popup uses (a fake in tests).
type window interface {
	isMinimised() bool
	isMaximised() bool
	size() (w, h int)
	position() (x, y int)
	setSize(w, h int)
	setMinSize(w, h int)
	setPosition(x, y int)
	center()
	unmaximise()
	maximise()
	minimise()
	show()
	hide()
	setOnTop(on bool)
	emitPopup(mode string) // "open", "closed" (back to the tray) or "expanded" (full app)
}

// popup remembers the main window's shape while the popup is shown.
type popup struct {
	mu     sync.Mutex
	active bool
	wasMin bool // the window was minimised (not in the tray) before the popup
	wasMax bool
	prevW  int
	prevH  int
	prevX  int
	prevY  int
	// pending is the link the popup is showing, for a UI that loads after the capture
	// event went out (a link arriving right after start).
	pending *api.Capture
}

// open turns the hidden or minimised window into the popup. It returns false (and does
// nothing) when the window is already on screen: then the dialog opens inside the app.
func (p *popup) open(w window, hidden bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active {
		w.show() // another capture while the popup is up: just bring it forward
		return true
	}
	minimised := !hidden && w.isMinimised()
	if !hidden && !minimised {
		return false
	}
	p.active = true
	p.wasMin = minimised
	p.wasMax = w.isMaximised()
	if p.wasMax {
		w.unmaximise()
	}
	p.prevW, p.prevH = w.size()
	p.prevX, p.prevY = w.position()
	w.emitPopup("open") // before the window shows, so the UI never flashes the full app
	w.setMinSize(popupWidth, popupMinH)
	w.setSize(popupWidth, popupHeight)
	w.center()
	w.setOnTop(true)
	w.show()
	return true
}

// close ends the popup. expand keeps the window on screen as the full app (tray click,
// "open Peregrine"); otherwise the window goes back to the tray or the taskbar.
// It returns whether the popup was active.
func (p *popup) close(w window, expand bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.active {
		return false
	}
	p.active = false
	p.pending = nil
	w.setOnTop(false)
	if !expand {
		// Hide first, then restore the shape, so the user never sees the big window.
		if p.wasMin {
			w.minimise()
		} else {
			w.hide()
		}
	}
	w.setMinSize(mainMinW, mainMinH)
	w.setSize(p.prevW, p.prevH)
	w.setPosition(p.prevX, p.prevY)
	if p.wasMax {
		w.maximise()
	}
	if expand {
		w.emitPopup("expanded")
	} else {
		w.emitPopup("closed")
	}
	return true
}

// fit sets the popup's height so the UI content (contentH CSS px) fits exactly, given the
// web view's current inner height. Working from the difference keeps it right whatever the
// title bar height and display scaling are.
func (p *popup) fit(w window, contentH, innerH int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.active || contentH <= 0 || innerH <= 0 {
		return
	}
	cw, ch := w.size()
	h := min(max(ch+contentH-innerH, popupMinH), popupMaxH)
	if h != ch {
		w.setSize(cw, h)
	}
}

// captured records a link offered while the popup is up.
func (p *popup) captured(c *api.Capture) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active && c != nil {
		cp := *c
		p.pending = &cp
	}
}

// pendingCapture is the link the popup shows, or nil.
func (p *popup) pendingCapture() *api.Capture {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.pending
}

func (p *popup) isActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active
}

// wailsWindow is the real window.
type wailsWindow struct{ ctx context.Context }

func (w wailsWindow) isMinimised() bool       { return wruntime.WindowIsMinimised(w.ctx) }
func (w wailsWindow) isMaximised() bool       { return wruntime.WindowIsMaximised(w.ctx) }
func (w wailsWindow) size() (int, int)        { return wruntime.WindowGetSize(w.ctx) }
func (w wailsWindow) position() (int, int)    { return wruntime.WindowGetPosition(w.ctx) }
func (w wailsWindow) setSize(width, h int)    { wruntime.WindowSetSize(w.ctx, width, h) }
func (w wailsWindow) setMinSize(width, h int) { wruntime.WindowSetMinSize(w.ctx, width, h) }
func (w wailsWindow) setPosition(x, y int)    { wruntime.WindowSetPosition(w.ctx, x, y) }
func (w wailsWindow) center()                 { wruntime.WindowCenter(w.ctx) }
func (w wailsWindow) unmaximise()             { wruntime.WindowUnmaximise(w.ctx) }
func (w wailsWindow) maximise()               { wruntime.WindowMaximise(w.ctx) }
func (w wailsWindow) minimise()               { wruntime.WindowMinimise(w.ctx) }
func (w wailsWindow) hide()                   { wruntime.WindowHide(w.ctx) }
func (w wailsWindow) setOnTop(on bool)        { wruntime.WindowSetAlwaysOnTop(w.ctx, on) }
func (w wailsWindow) emitPopup(mode string)   { wruntime.EventsEmit(w.ctx, "peregrine:popup", mode) }
func (w wailsWindow) show() {
	wruntime.WindowUnminimise(w.ctx)
	wruntime.WindowShow(w.ctx)
	wruntime.Show(w.ctx) // macOS: unhide the app (no-op elsewhere)
}

// showForCapture is the API's ShowWindow hook: it runs when a captured link needs the add
// dialog. A hidden or minimised window becomes the quick-add popup; a visible one comes
// to the front as before.
func (a *App) showForCapture() {
	if a.ctx == nil {
		return
	}
	a.life.mu.Lock()
	hidden := a.life.hidden
	a.life.mu.Unlock()
	if !a.pop.open(wailsWindow{a.ctx}, hidden) {
		a.showWindow()
		return
	}
	// The window is on screen: let UI events (progress) flow while the popup is up.
	a.life.mu.Lock()
	a.life.setHiddenLocked(false)
	a.life.mu.Unlock()
}

// PopupDone is called by the UI when the quick-add dialog closes (added or cancelled).
// expand opens the full app instead of going back to the tray.
func (a *App) PopupDone(expand bool) {
	if a.ctx == nil {
		return
	}
	if !a.pop.close(wailsWindow{a.ctx}, expand) {
		return
	}
	if !expand {
		a.life.mu.Lock()
		if !a.pop.wasMinimised() {
			a.life.setHiddenLocked(true)
		}
		a.life.mu.Unlock()
	}
}

// PopupFit is called by the UI with its content height so the popup has no empty space.
func (a *App) PopupFit(contentH, innerH int) {
	if a.ctx == nil {
		return
	}
	a.pop.fit(wailsWindow{a.ctx}, contentH, innerH)
}

// PopupActive tells a freshly loaded UI whether it is showing as the popup.
func (a *App) PopupActive() bool { return a.pop.isActive() }

// PopupPending returns the link the popup is for (nil if none), for a UI that loaded
// after the capture event was sent.
func (a *App) PopupPending() *api.Capture { return a.pop.pendingCapture() }

func (p *popup) wasMinimised() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.wasMin
}

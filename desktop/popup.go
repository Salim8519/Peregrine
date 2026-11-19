package main

import (
	"context"
	"sync"

	"github.com/Salim8519/Peregrine/engine"
	"github.com/Salim8519/Peregrine/internal/api"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Quick-add popup (ADR 0021): a link captured from the browser while the window is hidden
// in the tray, minimised, or (Windows) open behind the browser shows only a small add
// dialog, not the whole app. Wails v2 has
// one window, so the popup is that window shrunk, centred and kept on top; finishing the
// dialog puts it back the way it was and hides it again. The same popup shows the
// "download complete" dialog (ADR 0023).

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
	// flyToTray plays the short "fly to the tray" animation and returns a function that
	// undoes it (call it after hide).
	flyToTray() (restore func())
	setOnTop(on bool)
	// showBehind shows the window again without taking the focus, below the other windows
	// (the browser stays in front). Windows only; see canShowBehind.
	showBehind(maximised bool)
	// rememberFront records the window in front (the browser) so showBehind can give it
	// the focus back.
	rememberFront()
	emitPopup(mode string) // "open", "closed" (back to the tray) or "expanded" (full app)
}

// popup remembers the main window's shape while the popup is shown.
type popup struct {
	mu     sync.Mutex
	active bool
	from   popupFrom // where the window was before the popup
	wasMax bool
	prevW  int
	prevH  int
	prevX  int
	prevY  int
	kind   popupKind // what the popup shows
	// pending is the link the popup is showing, for a UI that loads after the capture
	// event went out (a link arriving right after start).
	pending *api.Capture
}

// popupKind is what the popup shows.
type popupKind int

const (
	kindAdd  popupKind = iota // the quick-add dialog for a captured link
	kindDone                  // "download complete" (ADR 0023)
)

// popupFrom is the window's state before the popup, where closing it puts it back.
type popupFrom int

const (
	fromTray   popupFrom = iota // hidden in the tray / menu bar
	fromMin                     // minimised to the taskbar / Dock
	fromBehind                  // open, but the user was in another app (the browser)
)

// open turns the window into the popup when the user is not looking at the app: hidden in
// the tray, minimised, or open but not focused (behind the browser; Windows only, where the
// window can be put back behind without taking the focus). It returns false (and does
// nothing) when the user is in the app: then the dialog opens inside it.
func (p *popup) open(w window, hidden, focused bool) bool {
	return p.openAs(w, kindAdd, hidden, focused)
}

// openAs is open for a given kind. A "download complete" popup never replaces a quick-add
// popup that is up (a link is waiting for the user); a captured link replaces a "download
// complete" popup.
func (p *popup) openAs(w window, kind popupKind, hidden, focused bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active {
		if kind == kindDone && p.kind == kindAdd {
			return false
		}
		p.kind = kind
		w.show() // another capture while the popup is up: just bring it forward
		return true
	}
	from := fromTray
	switch {
	case hidden:
	case w.isMinimised():
		from = fromMin
	case !focused && canShowBehind:
		from = fromBehind
	default:
		return false
	}
	if from == fromBehind {
		w.rememberFront()
	}
	p.active = true
	p.kind = kind
	p.from = from
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
		switch p.from {
		case fromMin:
			w.minimise()
		case fromTray:
			restore := w.flyToTray() // show where the window went
			w.hide()
			restore()
		default:
			w.hide()
		}
	}
	w.setMinSize(mainMinW, mainMinH)
	w.setSize(p.prevW, p.prevH)
	w.setPosition(p.prevX, p.prevY)
	switch {
	case !expand && p.from == fromBehind:
		w.showBehind(p.wasMax) // back where it was, behind the browser
	case p.wasMax:
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

// shows reports whether the popup is up and showing kind.
func (p *popup) shows(kind popupKind) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active && p.kind == kind
}

func (p *popup) isActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active
}

// wailsWindow is the real window.
type wailsWindow struct{ ctx context.Context }

func (w wailsWindow) isMinimised() bool         { return wruntime.WindowIsMinimised(w.ctx) }
func (w wailsWindow) isMaximised() bool         { return wruntime.WindowIsMaximised(w.ctx) }
func (w wailsWindow) size() (int, int)          { return wruntime.WindowGetSize(w.ctx) }
func (w wailsWindow) position() (int, int)      { return wruntime.WindowGetPosition(w.ctx) }
func (w wailsWindow) setSize(width, h int)      { wruntime.WindowSetSize(w.ctx, width, h) }
func (w wailsWindow) setMinSize(width, h int)   { wruntime.WindowSetMinSize(w.ctx, width, h) }
func (w wailsWindow) setPosition(x, y int)      { wruntime.WindowSetPosition(w.ctx, x, y) }
func (w wailsWindow) center()                   { wruntime.WindowCenter(w.ctx) }
func (w wailsWindow) unmaximise()               { wruntime.WindowUnmaximise(w.ctx) }
func (w wailsWindow) maximise()                 { wruntime.WindowMaximise(w.ctx) }
func (w wailsWindow) minimise()                 { wruntime.WindowMinimise(w.ctx) }
func (w wailsWindow) hide()                     { hideToTray(w.ctx) }
func (w wailsWindow) flyToTray() func()         { return flyToTray() }
func (w wailsWindow) setOnTop(on bool)          { wruntime.WindowSetAlwaysOnTop(w.ctx, on) }
func (w wailsWindow) showBehind(maximised bool) { showWindowBehind(maximised) }
func (w wailsWindow) rememberFront()            { rememberFrontWindow() }
func (w wailsWindow) emitPopup(mode string)     { wruntime.EventsEmit(w.ctx, "peregrine:popup", mode) }
func (w wailsWindow) show() {
	wruntime.WindowUnminimise(w.ctx)
	wruntime.WindowShow(w.ctx)
	wruntime.Show(w.ctx) // macOS: unhide the app (no-op elsewhere)
}

// showForCapture is the API's ShowWindow hook: it runs when a captured link needs the add
// dialog. A hidden, minimised or (Windows) unfocused window becomes the quick-add popup;
// when the user is in the app, the dialog opens inside it.
func (a *App) showForCapture() {
	if a.ctx == nil {
		return
	}
	a.life.mu.Lock()
	hidden, focused := a.life.hidden, a.life.focused
	a.life.mu.Unlock()
	if !a.pop.open(wailsWindow{a.ctx}, hidden, focused) {
		a.showWindow()
		return
	}
	// The window is on screen: let UI events (progress) flow while the popup is up.
	a.life.mu.Lock()
	a.life.setHiddenLocked(false)
	a.life.mu.Unlock()
}

// showDone shows the "download complete" dialog for d (ADR 0023): as the popup when the
// user is not looking at the app, inside the app when they are. It returns false when the
// dialog could not be shown (a quick-add popup is up, or the window is behind another app
// where it cannot be popped up); the caller then falls back to the system notification.
func (a *App) showDone(d engine.Download) bool {
	if a.ctx == nil {
		return false
	}
	a.life.mu.Lock()
	hidden, focused := a.life.hidden, a.life.focused
	a.life.mu.Unlock()
	if a.pop.shows(kindAdd) {
		return false
	}
	// The UI shows the dialog on this event, in the popup or in the app. Sent before the
	// popup opens, so the popup never shows empty.
	wruntime.EventsEmit(a.ctx, "peregrine:done", d)
	if a.pop.openAs(wailsWindow{a.ctx}, kindDone, hidden, focused) {
		a.life.mu.Lock()
		a.life.setHiddenLocked(false)
		a.life.mu.Unlock()
		return true
	}
	return focused && !hidden // in the app: the dialog is on screen
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
		if a.pop.wasFrom() == fromTray {
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

// wasFrom is where the last popup came from.
func (p *popup) wasFrom() popupFrom {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.from
}

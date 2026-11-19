package main

import (
	"testing"
	"time"

	"github.com/Salim8519/Peregrine/engine"
	"github.com/Salim8519/Peregrine/internal/api"
	"github.com/Salim8519/Peregrine/internal/testsrv"
)

// The engine calls subscribers while holding its lock. The desktop's event handling (tray,
// notifications) must never call back into the engine from inside that callback, or the
// whole app freezes on the first status change (v0.2.0 bug: Clear completed "did nothing").
func TestDesktopEventsDoNotFreezeEngine(t *testing.T) {
	store, err := engine.OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	st := engine.DefaultSettings()
	st.DownloadDir = t.TempDir()
	st.Proxy = "none"
	if err := store.SaveSettings(st); err != nil {
		t.Fatal(err)
	}
	m, err := engine.NewManager(store)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{m: m}
	a.life = newLife(a)
	unsub := m.Subscribe(func(e engine.Event) { a.life.event(api.Event{Event: e}) })

	srv := testsrv.New()
	defer srv.Close()

	done := make(chan error, 1)
	go func() {
		d, err := m.Add(engine.Request{URL: srv.FileURL("size=200000&name=a.bin")})
		if err != nil {
			done <- err
			return
		}
		for {
			got, err := m.Get(d.ID)
			if err != nil {
				done <- err
				return
			}
			if got.Status == engine.StatusCompleted {
				m.ClearCompleted()
				if n := len(m.List()); n != 0 {
					t.Errorf("clear completed left %d downloads", n)
				}
				done <- nil
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("engine froze: a desktop event handler called back into the engine under its lock")
	}
	unsub()
	a.life.stopEvents()
	_ = m.Close()
	_ = store.Close()
}

// macOS reports hide/unhide of the whole app (close button, Cmd+H, Dock click). Being
// hidden must reach everything that depends on it: the quick-add popup, the progress
// gate for the UI and memory trimming.
func TestAppHiddenChanged(t *testing.T) {
	a := &App{}
	a.life = newLife(a)
	defer a.life.stopEvents()
	l := a.life

	l.appHiddenChanged(true)
	l.mu.Lock()
	hidden, since := l.hidden, l.hiddenSince
	l.mu.Unlock()
	if !hidden || !l.uiHidden.Load() || since.IsZero() {
		t.Fatalf("after hide: hidden=%v uiHidden=%v since=%v", hidden, l.uiHidden.Load(), since)
	}
	if forwardToUI(api.Event{Event: engine.Event{Type: engine.EventProgress}}, l.uiHidden.Load()) {
		t.Error("progress still sent to a hidden UI")
	}
	if !shouldTrim(trimState{hidden: true, hiddenSince: since}, since.Add(trimAfter)) {
		t.Error("a hidden app is never trimmed")
	}

	l.appHiddenChanged(true) // a repeated hide keeps the first hide time
	l.mu.Lock()
	again := l.hiddenSince
	l.mu.Unlock()
	if !again.Equal(since) {
		t.Errorf("hiddenSince moved from %v to %v", since, again)
	}

	l.appHiddenChanged(false)
	l.mu.Lock()
	hidden = l.hidden
	l.mu.Unlock()
	if hidden || l.uiHidden.Load() {
		t.Fatalf("after unhide: hidden=%v uiHidden=%v", hidden, l.uiHidden.Load())
	}
}

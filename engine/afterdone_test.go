package engine

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Salim8519/Peregrine/internal/testsrv"
)

// The add dialog's "when done" choices are stored with the download, survive a restart and
// are still there when the download completes (the desktop app acts on them then).
func TestAfterDoneKeptWithDownload(t *testing.T) {
	srv := testsrv.New()
	defer srv.Close()
	db := filepath.Join(t.TempDir(), "p.db")
	m := newTestManager(t, db, nil)
	want := AfterDone{OpenFile: true, Popup: true}
	d, err := m.Add(Request{URL: srv.FileURL("size=1000"), Paused: true, AfterDone: want})
	if err != nil {
		t.Fatal(err)
	}
	if d.AfterDone != want {
		t.Fatalf("added: afterDone = %+v, want %+v", d.AfterDone, want)
	}
	m.Close()

	st, err := OpenStore(db)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := NewManager(st)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m2.Close(); st.Close() })
	if err := m2.Resume(d.ID); err != nil {
		t.Fatal(err)
	}
	got := waitStatus(t, m2, d.ID, StatusCompleted, 10*time.Second)
	if got.AfterDone != want {
		t.Fatalf("after restart: afterDone = %+v, want %+v", got.AfterDone, want)
	}

	// A download added without choices has none.
	d2, err := m2.Add(Request{URL: srv.FileURL("size=10"), Paused: true})
	if err != nil {
		t.Fatal(err)
	}
	if d2.AfterDone != (AfterDone{}) {
		t.Fatalf("default afterDone = %+v, want none", d2.AfterDone)
	}
}

// "Show a popup when done" is on by default, also for settings saved before it existed,
// and turning it off is remembered.
func TestDonePopupSetting(t *testing.T) {
	if !DefaultSettings().DonePopup {
		t.Fatal("donePopup should default to on")
	}
	st, err := OpenStore(filepath.Join(t.TempDir(), "p.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.db.Exec(`INSERT INTO settings(k, v) VALUES('settings', ?)`, `{"language":"en","uiVersion":1}`); err != nil {
		t.Fatal(err)
	}
	got, err := st.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if !got.DonePopup {
		t.Fatal("old settings: donePopup should be on")
	}
	got.DonePopup = false
	if err := st.SaveSettings(got); err != nil {
		t.Fatal(err)
	}
	if again, _ := st.LoadSettings(); again.DonePopup {
		t.Fatal("turning donePopup off was not remembered")
	}
}

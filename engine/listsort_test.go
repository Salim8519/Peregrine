package engine

import (
	"path/filepath"
	"testing"
)

// The download list shows the newest download first by default, settings saved before
// the option existed get that default, the user's choice survives a restart, and a bad
// value falls back to the default (ADR 0024).
func TestListSortSetting(t *testing.T) {
	if got := DefaultSettings().ListSort; got != "date-desc" {
		t.Fatalf("default listSort = %q, want date-desc", got)
	}
	db := filepath.Join(t.TempDir(), "p.db")
	st, err := OpenStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.db.Exec(`INSERT INTO settings(k, v) VALUES('settings', ?)`, `{"language":"en","uiVersion":1}`); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(st)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Settings().ListSort; got != "date-desc" {
		t.Fatalf("old settings: listSort = %q, want date-desc", got)
	}
	s := m.Settings()
	s.ListSort = "name-asc"
	if _, err := m.UpdateSettings(s); err != nil {
		t.Fatal(err)
	}
	m.Close()
	st.Close()

	st2, err := OpenStore(db)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := NewManager(st2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m2.Close(); st2.Close() })
	if got := m2.Settings().ListSort; got != "name-asc" {
		t.Fatalf("after restart: listSort = %q, want name-asc", got)
	}
	for _, bad := range []string{"", "name", "name-up", "secret-asc", "date-desc-x"} {
		s := m2.Settings()
		s.ListSort = bad
		got, err := m2.UpdateSettings(s)
		if err != nil {
			t.Fatal(err)
		}
		if got.ListSort != DefaultListSort {
			t.Errorf("listSort %q saved as %q, want %q", bad, got.ListSort, DefaultListSort)
		}
	}
}

package main

import (
	"path/filepath"
	"testing"
)

func TestAppBundle(t *testing.T) {
	root := t.TempDir() // absolute on every OS
	app := filepath.Join(root, "Applications", "Peregrine.app")
	cases := []struct {
		name   string
		exe    string
		want   string
		wantOK bool
	}{
		{"bundle main program", filepath.Join(app, "Contents", "MacOS", "Peregrine"), app, true},
		{"renamed bundle", filepath.Join(root, "My Apps.APP", "Contents", "MacOS", "Peregrine"), filepath.Join(root, "My Apps.APP"), true},
		{"unclean path", filepath.Join(app, "Contents", "MacOS") + string(filepath.Separator) + "." + string(filepath.Separator) + "Peregrine", app, true},
		{"loose binary", filepath.Join(root, "bin", "peregrine"), "", false},
		{"Resources, not MacOS", filepath.Join(app, "Contents", "Resources", "Peregrine"), "", false},
		{"not a .app folder", filepath.Join(root, "Peregrine", "Contents", "MacOS", "Peregrine"), "", false},
		{"bare .app name", filepath.Join(root, ".app", "Contents", "MacOS", "Peregrine"), "", false},
		{"relative path", filepath.Join("Peregrine.app", "Contents", "MacOS", "Peregrine"), "", false},
	}
	for _, c := range cases {
		got, ok := appBundle(c.exe)
		if got != c.want || ok != c.wantOK {
			t.Errorf("%s: appBundle(%q) = %q, %v; want %q, %v", c.name, c.exe, got, ok, c.want, c.wantOK)
		}
	}
}

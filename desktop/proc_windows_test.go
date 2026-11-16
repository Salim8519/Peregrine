//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Regression (0.4.1): "Show in folder" opened an invisible Explorer window because the
// command was started with HideWindow, which Explorer passes on to its window.
func TestRevealStartsVisible(t *testing.T) {
	r := revealCmd(`C:\Users\me\Downloads\a b.zip`)
	if r.SysProcAttr == nil || r.SysProcAttr.HideWindow {
		t.Fatal("revealCmd must not hide the Explorer window")
	}
	if want := `explorer /select,"C:\Users\me\Downloads\a b.zip"`; r.SysProcAttr.CmdLine != want {
		t.Fatalf("CmdLine = %q, want %q", r.SysProcAttr.CmdLine, want)
	}
}

// Regression (0.5.4): "Open file" did nothing for some paths (spaces and "&"), because
// rundll32 url.dll,FileProtocolHandler failed on them without an error.
func TestOpenPathWithSpacesAndAmpersand(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my files & more")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "ran.txt")
	script := filepath.Join(dir, "a b.cmd")
	body := "@echo ok> \"" + marker + "\"\r\nexit\r\n"
	if err := os.WriteFile(script, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := openPath(script); err != nil {
		t.Fatalf("openPath: %v", err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("the file was not opened")
}

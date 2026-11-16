//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}

// openPath opens a file with its default app through ShellExecute, like a double-click
// in Explorer. 0.5.3 and older used "rundll32 url.dll,FileProtocolHandler", which did
// nothing at all for some paths, for example with spaces and "&" (fixed in 0.5.4).
func openPath(path string) error {
	file, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	verb, _ := windows.UTF16PtrFromString("open")
	if err := windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}

// revealPath opens Explorer with the file selected.
func revealPath(path string) error {
	return startVisible(revealCmd(path))
}

// revealCmd starts Explorer, so it must not get HideWindow: Explorer passes the start-up
// show state on to its window, and a hidden flag opened an invisible Explorer window on
// every "Show in folder" click (fixed in 0.4.1).
func revealCmd(path string) *exec.Cmd {
	// explorer parses its own command line: /select,"<path>" must not be re-quoted by Go.
	cmd := exec.Command("explorer")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `explorer /select,"` + path + `"`}
	return cmd
}

func startVisible(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

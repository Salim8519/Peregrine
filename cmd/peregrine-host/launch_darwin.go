package main

import (
	"fmt"
	"os/exec"
)

// openTool is macOS's LaunchServices command-line front end (an absolute path: browsers
// start the host with a minimal PATH).
const openTool = "/usr/bin/open"

// startDetached opens the app bundle through LaunchServices, like a double-click in
// Finder. Started that way the app is its own process, not the browser's child, so macOS
// asks for and remembers permissions (notifications, folders) for Peregrine instead of
// the browser (ADR 0025). Outside a bundle (a dev build) it starts the binary directly.
func startDetached(path string) error {
	bundle, ok := appBundle(path)
	if !ok {
		return execDetached(path)
	}
	if out, err := exec.Command(openTool, bundle).CombinedOutput(); err != nil {
		return fmt.Errorf("open %s: %w: %s", bundle, err, out)
	}
	return nil
}

package main

import (
	"path/filepath"
	"strings"
)

// appBundle returns the .app bundle that holds exe when exe is a bundle's main program
// (<name>.app/Contents/MacOS/<exe>), and false otherwise. Pure path logic, so it is tested
// on every OS; only the macOS launcher uses it.
func appBundle(exe string) (string, bool) {
	if !filepath.IsAbs(exe) {
		return "", false
	}
	macOS := filepath.Dir(filepath.Clean(exe))
	contents := filepath.Dir(macOS)
	bundle := filepath.Dir(contents)
	if filepath.Base(macOS) != "MacOS" || filepath.Base(contents) != "Contents" ||
		!strings.HasSuffix(strings.ToLower(filepath.Base(bundle)), ".app") || len(filepath.Base(bundle)) <= len(".app") {
		return "", false
	}
	return bundle, true
}

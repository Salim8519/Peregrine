//go:build !windows && !darwin

package main

// startDetached starts the app so it outlives the host.
func startDetached(path string) error { return execDetached(path) }

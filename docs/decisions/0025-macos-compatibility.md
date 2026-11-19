# 0025 — macOS compatibility pass

- Status: accepted
- Date: 2026-11-19

## Context
Peregrine was built and released for macOS from the start (universal DMG, ADR 0005), but
only ever run on Windows. A code review against how macOS and Wails v2 actually behave found:

1. Go 1.27, which builds the app, links binaries for **macOS 13** or later; the docs said 12+.
2. Wails generated a default `Info.plist`: bundle ID `com.wails.Peregrine` (while launch at
   login, the single-instance lock and the docs use `com.peregrine.app`), minimum system
   10.13, and no `NSAppleEventsUsageDescription`, which macOS requires before an app may
   send Apple events — so "shut down when done" (System Events) could not work.
3. The close button on macOS hides the **whole app** (`[NSApp hide:]`, Wails'
   `HideWindowOnClose`) without telling Go. `life.hidden` stayed false, so a link from the
   browser never got the quick-add popup (ADR 0021), progress kept streaming to a hidden UI
   and memory was never trimmed (ADR 0015).
4. Wails has no Dock "reopen" handler, and its `WindowHide` only orders the window out. Any
   path that hides just the window (the popup going back "to the tray", launch at login with
   `--hidden`) left a Dock icon that showed nothing when clicked.
5. The native messaging host started the app by executing the binary inside the bundle, as
   a child of the browser. macOS attributes permission prompts (notifications, folders) to
   the "responsible" process, which for a browser child can be the browser.
6. First-launch steps said right-click → Open. Since macOS 15 that no longer bypasses
   Gatekeeper for apps that are not notarized.
7. Cosmetic: a full-colour tile in the menu bar, and shortcuts written "Ctrl+N".

## Decision
- **Our own `desktop/build/darwin/Info.plist`** (Wails uses it instead of its default):
  `CFBundleIdentifier` `com.peregrine.app`, `LSMinimumSystemVersion` 13.0, Apple events and
  folder usage descriptions, URL scheme names `com.peregrine.<scheme>`. The ID changes once
  now, while there are almost no macOS users; it must stay stable afterwards because macOS
  ties permissions (and later notarization) to it.
- **On macOS "hidden in the menu bar" means the app is hidden.** `macapp_darwin.m` observes
  `NSApplicationDidHideNotification` / `DidUnhideNotification` and reports them to Go
  (`life.appHiddenChanged`), so Cmd+H and the close button both count. The popup goes back
  "to the tray" by hiding the app (`hideToTray`), the same way, so a Dock click restores it.
- **Dock reopen:** `applicationShouldHandleReopen:hasVisibleWindows:` is added at run time to
  Wails' app delegate (only if it does not already implement it); with no visible window it
  shows the main window. This covers launch at login with `--hidden`.
- **Host launch on macOS:** `/usr/bin/open <Peregrine.app>` (absolute path; the bundle is
  derived from the host's own location, never from a message). A binary outside a bundle (dev
  build) is still started directly. Linux and Windows are unchanged.
- **Menu-bar icon:** a black silhouette (`build/tray-template.png`, made from `tray.png`)
  set as a template image; Windows/Linux keep the colour tile.
- **Shortcut labels:** the UI shows `Ctrl+X` as `⌘X` on macOS (one rule in `translate`,
  not 16 sets of strings). The key handler already accepts Ctrl and ⌘.
- Docs: macOS 13+ everywhere, first launch via System Settings → Privacy & Security →
  *Open Anyway*.

Windows code paths are untouched: the macOS parts live in `*_darwin.go`/`.m` files, and
`macapp_other.go` keeps `WindowHide` and a no-op watcher for Windows and Linux.

## Alternatives considered
- Turn off `HideWindowOnClose` and handle the close button in `beforeClose` like Windows —
  Wails sends Cmd+Q and the close button through the same "Q" message, so the two cannot be
  told apart there, and Cmd+Q must quit.
- Poll the window's visibility — needs the same Objective-C bridge, and is late by the
  polling interval; notifications are exact.
- Use `NSApplicationDidBecomeActiveNotification` instead of the reopen handler — Wails
  activates the app on every launch, which would defeat `--hidden`.
- `open -a Peregrine` by name — LaunchServices could pick another copy of the app; the
  bundle path opens exactly the one beside the host.

## Consequences
- The release needs `cgo` on macOS (already true: Wails and the tray need it).
- These macOS paths compile only on the macOS CI runner and cannot be unit-tested on Linux;
  the shared logic (`appHiddenChanged`, `appBundle`, `macShortcuts`) is tested everywhere.
  A manual check on a Mac is still needed for each release until notarization and a macOS
  UI test exist.
- Notification and folder permissions granted to `com.wails.Peregrine` (0.5.6 and earlier)
  are asked for again once.

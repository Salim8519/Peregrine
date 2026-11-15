# 0021 — Quick-add popup and notification buttons

- Status: accepted
- Date: 2026-11-15

## Context
A download offered by the browser opened the whole app window with the full add dialog
(address, folder, name, connections, checksum). Most of the time the user only wants to
check the name, the size and where it goes, then click Start, and the app was in the tray.
When the download finished, the notification could only open the app, not the file.

## Decision
- **Short form.** A link that was offered (extension, clipboard monitor, command line, OS
  link handler) opens the add dialog in a short form: file name and type, size ("Checking
  size…" while the probe runs), the folder with Browse, and Start / Add paused / Cancel.
  "More details" shows the full form (address, video-page switch, file name, connections,
  checksum). Typing a link yourself (Add URL) still opens the full form.
- **Popup window.** When a link is offered while the window is hidden in the tray or
  minimised, only the add dialog appears, in a small window (520 px wide) centred and kept on
  top. Wails v2 has one window, so the popup is that window: its size, position and
  maximised state are saved, it is shrunk, and the UI shows only the dialog. Its height
  follows the content (the UI reports its height; the window grows when "More details"
  opens). Start, Add paused, Cancel, Escape or the window's close button put the window back
  as it was and hide it again (tray) or minimise it again. A tray click or anything else that
  opens the app grows the popup into the full app with the dialog still open. When the window
  is already on screen, the dialog opens inside the app as before.
- A link handed over by a second launch or the OS (`Peregrine.exe <url>`, magnet links,
  `.torrent` files) no longer opens the full window first; the capture decides.
- The desktop keeps the link the popup is for, so a UI that loads after the capture event
  (a link right after start) still shows it (`PopupPending`).
- **Notification buttons.** The "Download complete" notification has two buttons, Open file
  and Show in folder, which call the existing `openFile` / `showInFolder` methods for that
  download. Clicking the notification itself still opens the app. Labels follow the app
  language (desktop/locales).
- The new Wails bindings (`PopupDone`, `PopupFit`, `PopupActive`, `PopupPending`) are UI-only:
  they are not RPC methods and the browser extension cannot reach them.

## Alternatives considered
- Wails v3 multi-window — a real second window would be cleaner, but v3 is a large migration
  of the whole desktop app for one dialog.
- A separate small helper program for the popup — a second UI codebase and process, against
  ADR 0002 (one UI codebase).
- Always the short form, also for Add URL — the address field is the reason that dialog is
  opened, so hiding it there would only add a click.
- A setting to turn the popup off — not added: "More details" and a tray click give the full
  dialog and the full app; can be added if users ask.

## Consequences
- Taking a download from the browser no longer pulls the whole app to the front.
- The window shape is changed and restored at runtime; on macOS the window's close button
  hides the app as usual and the shape is restored the next time the app is shown.
- Notification buttons need Wails' notification categories (Windows toasts, macOS
  UNNotificationCategory); where they are not supported the plain notification still shows.

# 0023 — "When done" choices in the add dialog and a "download complete" popup

- Status: accepted
- Date: 2026-11-18

## Context
After a download finished, the user had to find it in the list (or catch the system
notification) to open the file or its folder. Other download managers let you choose this
when you add the download, and show a small "download complete" window.

## Decision
- **Three choices in the add dialog**, on one line under the folder (both the short and the
  full form, so also in the quick-add popup), with emojis: 📄 Open file, 📂 Open folder,
  🔔 Show popup. The line wraps when the window is narrow or the language is long; the
  popup's height already follows its content (ADR 0021), its width stays 520 px.
- **Open file / Open folder start off every time.** Opening a file on its own is a choice
  for this download only, never a default.
- **Show popup starts from the last choice**, stored in the settings as `donePopup`
  (default on; settings saved before it existed read as on). Unticking it in the add dialog
  turns it off for the next downloads too; it is also in Settings.
- The choices are stored with the download (`Request.afterDone` / `Download.afterDone`:
  `openFile`, `openFolder`, `popup`), so they survive a restart. The engine only stores them;
  the desktop app acts when the download completes: open the file, show it in its folder
  (the existing `openFile` / `showInFolder` methods), then the popup.
- **The popup** is the quick-add popup window (same size and look, ADR 0021) showing
  "✅ Download complete": file icon, name, size, folder, and Close / 📂 Open folder /
  📄 Open file. It shows when the user is not looking at the app; when they are, the same
  dialog opens inside the app. It replaces the system notification for that download.
  A quick-add popup that is waiting for the user is never replaced: the notification is
  shown instead. A new captured link replaces a "download complete" popup. A newer finished
  download replaces an older one in the popup.
- Downloads added without the dialog (batch, extension auto-start, CLI) have no choices:
  they get the notification as before.
- The add dialog's Start now / Add paused buttons use emojis (⬇️ / ⏸️) instead of icons.

## Alternatives considered
- Remember Open file / Open folder too — a remembered "open every file" would run
  downloaded programs without the user thinking about it each time.
- A second Wails window for the done popup — same reason as ADR 0021 (Wails v2 has one).
- Per-download popup without a remembered default — the user asked for it to remember.

## Consequences
- A file can open itself when it finishes, so this is a new path from download to OS
  (SECURITY-MODEL T27): only the user's own tick in the app sets it; the native messaging
  host never forwards `afterDone` (its types do not have it), and the file already has its
  Mark-of-the-Web when it opens.

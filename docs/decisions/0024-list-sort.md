# 0024 — Download list: newest first, remembered sort

- Status: accepted
- Date: 2026-11-18

## Context
The download list showed downloads in queue order, which is oldest first, so a new download
landed at the bottom and users had to scroll to find it. Sorting by a column header was
forgotten on every start.

## Decision
- The default sort is "date added, descending": the newest download is at the top.
- The sort is a setting, `Settings.listSort = "<column>-<asc|desc>"`, stored by the engine
  like the other desktop-only settings. Clicking a header saves it at once through
  `updateSettings`.
- A header click still cycles ascending → descending → queue order (`order-asc`), so the
  queue order that Move up / Move down change can still be seen.
- The engine accepts only known columns and directions; anything else becomes `date-desc`.
  Settings saved before this change get `date-desc` (missing fields keep their defaults).

## Alternatives considered
- `localStorage` in the web view — works, but the engine's settings are the app's single
  store (the theme there is only a first-paint cache), and the popup window and a reinstall
  would not always see it.
- Changing the queue order itself to newest first — would also change which download starts
  next; the user asked only how the list looks.

## Consequences
- One more field in `Settings`; `updateSettings` is still not on the extension allowlist.
- Move up / Move down have a visible effect only in queue order, as before with any sort.

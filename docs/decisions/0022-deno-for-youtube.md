# 0022 — Bundle Deno so yt-dlp can download from YouTube

- Status: accepted
- Date: 2026-11-15

## Context
YouTube downloads failed in 0.5.2. Tested live on 2026-11-15 with the bundled yt-dlp
2026.10.26:

- Without a JavaScript runtime, yt-dlp warns that YouTube extraction without one is
  deprecated, and with the user's browser cookies it then fails with
  "No video formats found" / "Requested format is not available". YouTube's player
  challenges (signature and `n` parameter) now need real JavaScript, which yt-dlp runs
  through an external runtime (its "EJS" scripts are bundled in the official
  `yt-dlp.exe` / `yt-dlp_macos` builds; the runtime is not).
- With Deno 2.9.8 and the same cookies the video downloaded (AV1 + AAC merged).
- Without cookies YouTube asked "Sign in to confirm you're not a bot" for every player
  client tried. That is decided by YouTube per network/IP, not by the tool, so other
  libraries (go-ytdlp is only a wrapper around yt-dlp; kkdai/youtube is a pure-Go
  YouTube client) would hit the same wall. The browser extension already sends the
  page's cookies, which passed the check.

Measured on Windows (Deno 2.9.8, one YouTube video): `deno.exe` is 98 MB (43 MB zipped);
it runs only while yt-dlp extracts the video, peaking at about 285 MB RAM and 2 s CPU,
then exits. It does not run during the download itself.

## Decision
- Ship **Deno** next to yt-dlp, like ffmpeg (ADR 0013): `deno.exe` in the Windows install
  and portable folders, a universal `deno` (lipo of the x86_64 and aarch64 builds) in
  `Peregrine.app/Contents/MacOS`. Pinned version (2.9.8), official release zips from
  `github.com/denoland/deno`, SHA-256 sums hard-coded in the release workflow, version
  checked after unpacking.
- `engine/proto/video` finds it like the other tools (`Options.DenoPath`, else next to the
  executable / in Resources, else `PATH`) and passes it explicitly with
  `--js-runtimes deno:<path>`. yt-dlp would find it next to itself on Windows, but not in
  the macOS app folder, and naming it keeps the choice ours. Without Deno nothing is
  passed and other sites still work.
- Deno is MIT licensed; the notice is added to `THIRD-PARTY-NOTICES.txt`.
- No pure-Go YouTube library as a fallback.

## Alternatives considered
- **kkdai/youtube as a hidden fallback** — pure Go and light, but YouTube-only, breaks on
  every player change, and fails the same bot check; ADR 0013 already rejected
  site-specific scraping in our code.
- **go-ytdlp** — Go bindings that run the same yt-dlp binary; adds nothing when yt-dlp
  fails.
- **QuickJS / QuickJS-NG** (about 2 MB) — supported by yt-dlp, but slow (old builds take
  minutes; it writes temporary files per script). Worth measuring later if the size of
  Deno becomes a problem.
- **Node.js** — also supported (22+), but larger to ship and runs with fewer restrictions
  than Deno.
- **Download Deno on first use** — a network call outside the user's downloads and an
  unsigned binary fetched at run time; rejected for the same reasons as auto-updating
  yt-dlp (ADR 0013).

## Consequences
- Deno adds about 29 MB to the Windows installer (LZMA) and 94 MB on disk; the macOS DMG
  more (universal binary).
- To pay for it, **ffprobe is no longer shipped** (planned in ADR 0015). Our code never
  calls it. yt-dlp uses it only to probe files: the audio codec before merging and the
  HLS MPEG-TS fix-up fall back to ffmpeg (tested 2026-11-15: an HLS stream saved as mp4
  without ffprobe gave a warning, was fixed with ffmpeg, and played); the features that
  need ffprobe itself (embedding subtitles or thumbnails, splitting chapters,
  concatenating playlists) are not used. Saves about 100 MB on disk.
- Deno runs only scripts yt-dlp ships, and yt-dlp starts it with restricted permissions
  (no file system or network access), so it adds little attack surface beyond yt-dlp's
  own (SECURITY-MODEL T20). Deno must be bumped with releases like yt-dlp.
- A link pasted into the app has no browser cookies, so on networks YouTube flags it can
  still fail with the bot check; adding a link from the browser extension works.

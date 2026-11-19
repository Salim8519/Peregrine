
---
### Install
- **Windows:** `Peregrine-*-windows-x64-setup.exe` (or the portable zip). Unsigned preview: SmartScreen → *More info* → *Run anyway*.
- **macOS 13+:** `Peregrine-*-macos-universal.dmg`, drag to Applications, open it once, then System Settings → Privacy & Security → *Open Anyway*.
- **Browser extension:** unzip `Peregrine-*-chromium-extension.zip`, open `chrome://extensions`, enable *Developer mode*, *Load unpacked*. Start Peregrine once so it can register the browser connection.
- Verify files with `SHA256SUMS.txt`, and check they were built by this repository's CI (not on anyone's laptop):
  `gh attestation verify <file> -R Salim8519/Peregrine`

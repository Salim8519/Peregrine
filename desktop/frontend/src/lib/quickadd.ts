import { Capture } from "./types";

/**
 * A link offered by the browser, the clipboard or the command line opens the short add
 * form (name, folder, size, "More details"). Typing a link yourself opens the full form,
 * since the address field is what you came for.
 */
export function startsCompact(c: Partial<Capture>): boolean {
  return !!c.url && !!c.source;
}

/** Best-effort file name from a link, for the moment before the probe answers. */
export function nameFromURL(raw: string): string {
  const u = raw.trim();
  if (!u) return "";
  if (/^magnet:/i.test(u)) {
    const dn = /[?&]dn=([^&]+)/i.exec(u);
    return dn ? safeDecode(dn[1].replace(/\+/g, " ")) : "";
  }
  try {
    const url = new URL(u);
    const last = url.pathname.split("/").filter(Boolean).pop() ?? "";
    return safeDecode(last) || url.hostname;
  } catch {
    return "";
  }
}

function safeDecode(s: string): string {
  try {
    return decodeURIComponent(s);
  } catch {
    return s;
  }
}

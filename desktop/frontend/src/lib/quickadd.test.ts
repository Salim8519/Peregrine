import { describe, expect, it } from "vitest";
import { nameFromURL, startsCompact } from "./quickadd";

describe("quick add", () => {
  it("opens the short form only for offered links", () => {
    expect(startsCompact({ url: "https://x.test/a.zip", source: "extension" })).toBe(true);
    expect(startsCompact({ url: "https://x.test/a.zip", source: "clipboard" })).toBe(true);
    expect(startsCompact({})).toBe(false);
    expect(startsCompact({ url: "https://x.test/a.zip" })).toBe(false);
    expect(startsCompact({ source: "extension" })).toBe(false);
  });

  it("guesses a name from the link", () => {
    expect(nameFromURL("https://x.test/files/My%20File.zip?token=1")).toBe("My File.zip");
    expect(nameFromURL("https://x.test/")).toBe("x.test");
    expect(nameFromURL("magnet:?xt=urn:btih:abc&dn=Ubuntu+24.04.iso")).toBe("Ubuntu 24.04.iso");
    expect(nameFromURL("magnet:?xt=urn:btih:abc")).toBe("");
    expect(nameFromURL("not a url")).toBe("");
    expect(nameFromURL("https://x.test/bad%E0%A4%A")).toBe("bad%E0%A4%A");
  });
});

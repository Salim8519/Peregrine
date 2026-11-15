import { describe, expect, it } from "vitest";
import { compareDownloads } from "./download";
import { Download } from "./types";

const dl = (id: string, createdAt: string, order: number) => ({ id, createdAt, order }) as unknown as Download;

describe("compareDownloads by date", () => {
  it("sorts by the time the download was added", () => {
    const rows = [dl("b", "2026-11-15T10:00:00Z", 1), dl("a", "2026-11-14T08:00:00Z", 2), dl("c", "2026-11-15T12:30:00Z", 0)];
    rows.sort((x, y) => compareDownloads(x, y, "date"));
    expect(rows.map((r) => r.id)).toEqual(["a", "b", "c"]);
  });
  it("falls back to list order for the same time", () => {
    const t = "2026-11-15T10:00:00Z";
    expect(compareDownloads(dl("x", t, 1), dl("y", t, 2), "date")).toBeLessThan(0);
  });
});

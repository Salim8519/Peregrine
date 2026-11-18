import { describe, expect, it } from "vitest";
import { compareDownloads, DEFAULT_LIST_SORT, formatListSort, nextListSort, parseListSort } from "./download";
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

describe("list sort (ADR 0024)", () => {
  it("defaults to newest first", () => {
    expect(DEFAULT_LIST_SORT).toEqual({ key: "date", dir: -1 });
    expect(parseListSort(undefined)).toEqual(DEFAULT_LIST_SORT);
    const rows = [dl("old", "2026-11-14T08:00:00Z", 0), dl("new", "2026-11-16T08:00:00Z", 1), dl("mid", "2026-11-15T08:00:00Z", 2)];
    const { key, dir } = parseListSort("date-desc");
    rows.sort((x, y) => compareDownloads(x, y, key) * dir);
    expect(rows.map((r) => r.id)).toEqual(["new", "mid", "old"]);
  });
  it("reads and writes the saved choice", () => {
    expect(parseListSort("name-asc")).toEqual({ key: "name", dir: 1 });
    expect(parseListSort("size-desc")).toEqual({ key: "size", dir: -1 });
    expect(formatListSort({ key: "eta", dir: -1 })).toBe("eta-desc");
    for (const bad of ["", "name", "name-up", "secret-asc", "date-desc-x"]) expect(parseListSort(bad)).toEqual(DEFAULT_LIST_SORT);
  });
  it("cycles a column: ascending, descending, then queue order", () => {
    let s = nextListSort(DEFAULT_LIST_SORT, "name");
    expect(s).toEqual({ key: "name", dir: 1 });
    s = nextListSort(s, "name");
    expect(s).toEqual({ key: "name", dir: -1 });
    expect(nextListSort(s, "name")).toEqual({ key: "order", dir: 1 });
  });
});

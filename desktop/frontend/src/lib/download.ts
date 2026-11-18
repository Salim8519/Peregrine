import { Download, Status } from "./types";

/** True when a failed download looks like an expired/forbidden link (fix: refresh the URL). */
export function isExpired(d: Download): boolean {
  return d.status === "error" && /expired|access was denied|403|401|forbidden/i.test(d.error ?? "");
}

export const STATUS_FILTERS = ["all", "downloading", "queued", "paused", "completed", "error"] as const;
export type StatusFilter = (typeof STATUS_FILTERS)[number];

export const CATEGORIES = ["Video", "Music", "Documents", "Compressed", "Programs", "Images", "Disk Images", "Other"] as const;

export function matchesStatus(d: Download, f: StatusFilter): boolean {
  return f === "all" || d.status === (f as Status);
}

export type SortKey = "order" | "name" | "size" | "progress" | "speed" | "eta" | "status" | "date";

const SORT_KEYS: readonly SortKey[] = ["order", "name", "size", "progress", "speed", "eta", "status", "date"];

export interface ListSort {
  key: SortKey;
  dir: 1 | -1;
}

/** Newest download at the top (ADR 0024). */
export const DEFAULT_LIST_SORT: ListSort = { key: "date", dir: -1 };

/** Reads Settings.listSort ("<column>-<asc|desc>"); anything else gives the default. */
export function parseListSort(s: string | undefined): ListSort {
  const [key, dir, extra] = (s ?? "").split("-");
  if (extra === undefined && SORT_KEYS.includes(key as SortKey) && (dir === "asc" || dir === "desc")) {
    return { key: key as SortKey, dir: dir === "asc" ? 1 : -1 };
  }
  return DEFAULT_LIST_SORT;
}

export function formatListSort(s: ListSort): string {
  return `${s.key}-${s.dir === 1 ? "asc" : "desc"}`;
}

/** A click on a column header: ascending, then descending, then back to queue order. */
export function nextListSort(cur: ListSort, k: SortKey): ListSort {
  if (k !== cur.key) return { key: k, dir: 1 };
  if (cur.dir === 1) return { key: k, dir: -1 };
  return { key: "order", dir: 1 };
}

const statusRank: Record<Status, number> = { downloading: 0, queued: 1, paused: 2, error: 3, completed: 4 };

export function compareDownloads(a: Download, b: Download, key: SortKey): number {
  switch (key) {
    case "name":
      return a.filename.localeCompare(b.filename, undefined, { numeric: true, sensitivity: "base" });
    case "size":
      return a.size - b.size;
    case "progress": {
      const pa = a.size > 0 ? a.downloaded / a.size : a.status === "completed" ? 1 : 0;
      const pb = b.size > 0 ? b.downloaded / b.size : b.status === "completed" ? 1 : 0;
      return pa - pb;
    }
    case "speed":
      return a.speed - b.speed;
    case "eta":
      return (a.eta < 0 ? Infinity : a.eta) - (b.eta < 0 ? Infinity : b.eta);
    case "status":
      return statusRank[a.status] - statusRank[b.status];
    case "date":
      return Date.parse(a.createdAt) - Date.parse(b.createdAt) || a.order - b.order;
    default:
      return a.order - b.order;
  }
}

export function canPause(d: Download) {
  return d.status === "downloading" || d.status === "queued";
}

export function canResume(d: Download) {
  return d.status === "paused" || d.status === "error";
}

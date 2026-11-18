/**
 * Desktop-only bindings outside the RPC API: window focus (for system notifications), the
 * scheduler's when-done countdown, and the platform. All are no-ops in browser mode.
 */
import { Capture, Download, WhenDone } from "./types";

export interface PowerEvent {
  action: WhenDone | "";
  seconds: number;
}

interface DesktopWindow {
  go?: {
    main?: {
      App?: {
        SetFocused?(f: boolean): Promise<void>;
        CancelPowerAction?(): Promise<void>;
        PopupDone?(expand: boolean): Promise<void>;
        PopupFit?(contentH: number, innerH: number): Promise<void>;
        PopupActive?(): Promise<boolean>;
        PopupPending?(): Promise<Capture | null>;
      };
    };
  };
  runtime?: {
    EventsOn(name: string, cb: (...data: unknown[]) => void): () => void;
    Environment?(): Promise<{ platform: string }>;
  };
}

const w = () => window as unknown as DesktopWindow;

export function setFocused(focused: boolean): void {
  void w()
    .go?.main?.App?.SetFocused?.(focused)
    ?.catch(() => undefined);
}

export function cancelPowerAction(): void {
  void w()
    .go?.main?.App?.CancelPowerAction?.()
    ?.catch(() => undefined);
}

export function onPower(fn: (e: PowerEvent) => void): () => void {
  const off = w().runtime?.EventsOn("peregrine:power", (e) => fn(e as PowerEvent));
  return () => off?.();
}

/** "windows" | "darwin" | "linux", or "" in browser mode. */
export async function platform(): Promise<string> {
  try {
    return (await w().runtime?.Environment?.())?.platform ?? "";
  } catch {
    return "";
  }
}

/**
 * Quick-add popup (ADR 0021): a link captured while the app is in the tray shows only the
 * add dialog in a small window. "open" starts it, "closed" means the window went back to the
 * tray, "expanded" means it grew into the full app (tray click).
 */
export type PopupMode = "open" | "closed" | "expanded";

export function onPopup(fn: (mode: PopupMode) => void): () => void {
  const off = w().runtime?.EventsOn("peregrine:popup", (m) => fn(m as PopupMode));
  return () => off?.();
}

export async function popupActive(): Promise<boolean> {
  try {
    return (await w().go?.main?.App?.PopupActive?.()) ?? false;
  } catch {
    return false;
  }
}

/** The link the popup is for, when the UI loaded after its capture event went out. */
export async function popupPending(): Promise<Capture | null> {
  try {
    return (await w().go?.main?.App?.PopupPending?.()) ?? null;
  } catch {
    return null;
  }
}

/** Ends the popup: back to the tray, or (expand) open the full app. */
export function popupDone(expand: boolean): void {
  void w()
    .go?.main?.App?.PopupDone?.(expand)
    ?.catch(() => undefined);
}

/** Asks the window to fit contentH CSS pixels of content. */
export function popupFit(contentH: number): void {
  void w()
    .go?.main?.App?.PopupFit?.(Math.ceil(contentH), window.innerHeight)
    ?.catch(() => undefined);
}

/** A download finished and asked for the "download complete" dialog (ADR 0023). */
export function onDone(fn: (d: Download) => void): () => void {
  const off = w().runtime?.EventsOn("peregrine:done", (d) => fn(d as Download));
  return () => off?.();
}

"use client";

import { useState } from "react";

type CopyState = "idle" | "copied" | "failed";

/** Copies `text` to the clipboard and stamps a short confirmation. */
export function CopyButton({
  text,
  label = "Copy",
}: {
  text: string;
  label?: string;
}) {
  const [state, setState] = useState<CopyState>("idle");

  /** Write the command to the clipboard and show a short result. */
  async function onCopy() {
    try {
      await navigator.clipboard.writeText(text);
      setState("copied");
    } catch {
      setState("failed");
    }
    window.setTimeout(() => setState("idle"), 1600);
  }

  const shown = state === "copied" ? "Copied" : state === "failed" ? "Denied" : label;

  return (
    <button
      type="button"
      onClick={onCopy}
      aria-label={shown}
      className="inline-flex shrink-0 items-center justify-center bg-ink p-2 text-paper hover:bg-stamp"
    >
      {state === "copied" ? <CheckIcon /> : <ClipboardIcon />}
      <span className="sr-only" aria-live="polite">
        {shown}
      </span>
    </button>
  );
}

/** A clipboard, the idle copy control. */
function ClipboardIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden="true">
      <rect x="8" y="2" width="8" height="4" rx="1" fill="none" stroke="currentColor" strokeWidth="2" />
      <path
        d="M8 4H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2h-2"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** A check, shown after the command is copied. */
function CheckIcon() {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" aria-hidden="true">
      <path
        d="M5 12.5 9.5 17 19 7"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

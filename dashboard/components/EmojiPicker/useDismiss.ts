"use client";

import { RefObject, useEffect } from "react";

/** Closes a popover on outside click or Escape. Shared by EmojiPicker + EmojiInsertButton. */
export function useDismiss(ref: RefObject<HTMLElement | null>, onDismiss: () => void) {
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onDismiss();
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") onDismiss();
    };
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [ref, onDismiss]);
}

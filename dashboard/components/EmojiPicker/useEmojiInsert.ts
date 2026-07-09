"use client";

import { RefObject } from "react";

/** Returns a handler that splices an emoji into a controlled input/textarea at the caret
 *  (replacing any selection), then restores focus and places the caret after the inserted text. */
export function useEmojiInsert(
  ref: RefObject<HTMLTextAreaElement | HTMLInputElement | null>,
  value: string,
  setValue: (v: string) => void,
) {
  return (emoji: string) => {
    const el = ref.current;
    const start = el?.selectionStart ?? value.length;
    const end = el?.selectionEnd ?? value.length;
    setValue(value.slice(0, start) + emoji + value.slice(end));
    const caret = start + emoji.length;
    // Runs after React commits the new value so the caret lands right after the emoji.
    requestAnimationFrame(() => {
      el?.focus();
      el?.setSelectionRange(caret, caret);
    });
  };
}

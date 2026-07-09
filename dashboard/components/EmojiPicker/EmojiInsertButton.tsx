"use client";

import { useRef, useState } from "react";
import { Smile } from "lucide-react";
import { DiscordEmoji } from "@/lib/api";
import EmojiPanel from "./EmojiPanel";
import { useDismiss } from "./useDismiss";
import styles from "./EmojiPicker.module.css";

interface EmojiInsertButtonProps {
  /** Called with the picked value (unicode glyph or <:name:id> mention) to splice into a field. */
  onSelect: (value: string) => void;
  /** The guild's custom emojis, powering the "Server" tab. */
  customEmojis?: DiscordEmoji[];
  disabled?: boolean;
  className?: string;
}

/** Compact icon trigger that opens the combined emoji panel and reports the pick to `onSelect`.
 *  Unlike EmojiPicker it never shows a value — meant to insert emoji into a free-text field. */
export default function EmojiInsertButton({
  onSelect,
  customEmojis = [],
  disabled = false,
  className,
}: EmojiInsertButtonProps) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  useDismiss(wrapperRef, () => setOpen(false));

  const pick = (v: string) => {
    onSelect(v);
    setOpen(false);
  };

  return (
    <div ref={wrapperRef} className={styles.insertWrapper}>
      <button
        type="button"
        disabled={disabled}
        aria-label="Insert emoji"
        title="Insert emoji"
        className={`${styles.iconTrigger} ${open ? styles.iconTriggerOpen : ""} ${className || ""}`}
        onClick={() => setOpen(o => !o)}
      >
        <Smile size={16} />
      </button>

      {open && <EmojiPanel customEmojis={customEmojis} onPick={pick} />}
    </div>
  );
}

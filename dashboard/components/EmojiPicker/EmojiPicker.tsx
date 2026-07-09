"use client";

import { useRef, useState } from "react";
import { ChevronDown, X } from "lucide-react";
import { DiscordEmoji } from "@/lib/api";
import EmojiView from "@/components/EmojiView/EmojiView";
import EmojiPanel from "./EmojiPanel";
import { useDismiss } from "./useDismiss";
import styles from "./EmojiPicker.module.css";

interface EmojiPickerProps {
  /** Current value: a unicode glyph, or a custom-emoji mention (<:name:id> / <a:name:id>). */
  value: string;
  onChange: (value: string) => void;
  /** The guild's custom emojis, powering the "Server" tab. */
  customEmojis?: DiscordEmoji[];
  /** Whether the value can be cleared back to empty (shows a clear button). */
  clearable?: boolean;
  placeholder?: string;
  className?: string;
}

export default function EmojiPicker({
  value,
  onChange,
  customEmojis = [],
  clearable = false,
  placeholder = "Pick emoji",
  className,
}: EmojiPickerProps) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  useDismiss(wrapperRef, () => setOpen(false));

  const pick = (v: string) => {
    onChange(v);
    setOpen(false);
  };

  return (
    <div ref={wrapperRef} className={`${styles.wrapper} ${className || ""}`}>
      <button
        type="button"
        className={`${styles.trigger} ${open ? styles.triggerOpen : ""}`}
        onClick={() => setOpen(o => !o)}
      >
        <span className={styles.triggerValue}>
          {value ? <EmojiView emoji={value} size={20} /> : <span className={styles.placeholder}>{placeholder}</span>}
        </span>
        {clearable && value ? (
          <span
            role="button"
            tabIndex={0}
            aria-label="Clear emoji"
            className={styles.clear}
            onClick={(e) => {
              e.stopPropagation();
              onChange("");
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                e.stopPropagation();
                onChange("");
              }
            }}
          >
            <X size={14} />
          </span>
        ) : (
          <ChevronDown size={16} className={`${styles.chevron} ${open ? styles.chevronOpen : ""}`} />
        )}
      </button>

      {open && <EmojiPanel customEmojis={customEmojis} onPick={pick} activeValue={value} />}
    </div>
  );
}

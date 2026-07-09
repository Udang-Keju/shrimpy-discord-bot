"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { ChevronDown, Search, X } from "lucide-react";
import { DiscordEmoji } from "@/lib/api";
import EmojiView from "@/components/EmojiView/EmojiView";
import { EMOJI_CATEGORIES } from "./emojiData";
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

type Tab = "standard" | "server";

export default function EmojiPicker({
  value,
  onChange,
  customEmojis = [],
  clearable = false,
  placeholder = "Pick emoji",
  className,
}: EmojiPickerProps) {
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState<Tab>("standard");
  const [query, setQuery] = useState("");
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, []);

  const q = query.trim().toLowerCase();

  const filteredCategories = useMemo(() => {
    if (!q) return EMOJI_CATEGORIES;
    return EMOJI_CATEGORIES.map(cat => ({
      label: cat.label,
      emojis: cat.emojis.filter(e => e.keywords.includes(q) || e.char === q),
    })).filter(cat => cat.emojis.length > 0);
  }, [q]);

  const filteredCustom = useMemo(() => {
    if (!q) return customEmojis;
    return customEmojis.filter(e => e.name.toLowerCase().includes(q));
  }, [q, customEmojis]);

  const pick = (v: string) => {
    onChange(v);
    setOpen(false);
    setQuery("");
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

      {open && (
        <div className={styles.popover}>
          <div className={styles.tabs}>
            <button
              type="button"
              className={`${styles.tab} ${tab === "standard" ? styles.tabActive : ""}`}
              onClick={() => setTab("standard")}
            >
              Standard
            </button>
            <button
              type="button"
              className={`${styles.tab} ${tab === "server" ? styles.tabActive : ""}`}
              onClick={() => setTab("server")}
            >
              Server{customEmojis.length > 0 ? ` (${customEmojis.length})` : ""}
            </button>
          </div>

          <div className={styles.searchRow}>
            <Search size={14} className={styles.searchIcon} />
            <input
              className={styles.search}
              type="text"
              value={query}
              onChange={e => setQuery(e.target.value)}
              placeholder={tab === "server" ? "Search server emoji..." : "Search emoji..."}
              autoFocus
            />
          </div>

          <div className={styles.grid}>
            {tab === "standard" ? (
              filteredCategories.length === 0 ? (
                <div className={styles.empty}>No emoji match &ldquo;{query}&rdquo;.</div>
              ) : (
                filteredCategories.map(cat => (
                  <div key={cat.label} className={styles.category}>
                    <div className={styles.categoryLabel}>{cat.label}</div>
                    <div className={styles.emojiRow}>
                      {cat.emojis.map(e => (
                        <button
                          key={e.char}
                          type="button"
                          title={e.keywords}
                          className={`${styles.emojiBtn} ${value === e.char ? styles.emojiBtnActive : ""}`}
                          onClick={() => pick(e.char)}
                        >
                          <span className={styles.emojiGlyph}>{e.char}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                ))
              )
            ) : customEmojis.length === 0 ? (
              <div className={styles.empty}>This server has no custom emoji.</div>
            ) : filteredCustom.length === 0 ? (
              <div className={styles.empty}>No server emoji match &ldquo;{query}&rdquo;.</div>
            ) : (
              <div className={styles.emojiRow}>
                {filteredCustom.map(e => (
                  <button
                    key={e.id}
                    type="button"
                    title={`:${e.name}:`}
                    className={`${styles.emojiBtn} ${value === e.mention ? styles.emojiBtnActive : ""}`}
                    onClick={() => pick(e.mention)}
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={e.url} alt={e.name} className={styles.customImg} />
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

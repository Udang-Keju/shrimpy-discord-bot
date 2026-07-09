"use client";

import { useMemo, useState } from "react";
import { Search } from "lucide-react";
import { DiscordEmoji } from "@/lib/api";
import { EMOJI_CATEGORIES } from "./emojiData";
import styles from "./EmojiPicker.module.css";

type Tab = "standard" | "server";

interface EmojiPanelProps {
  customEmojis?: DiscordEmoji[];
  /** Called with the selected value: a unicode glyph or a custom-emoji mention (<:name:id>). */
  onPick: (value: string) => void;
  /** Highlights the currently-selected value (replace-mode pickers pass their value). */
  activeValue?: string;
}

/** Shared popover body: Standard/Server tabs, search, and emoji grid. Used by
 *  EmojiPicker (replace mode) and EmojiInsertButton (insert mode). */
export default function EmojiPanel({ customEmojis = [], onPick, activeValue }: EmojiPanelProps) {
  const [tab, setTab] = useState<Tab>("standard");
  const [query, setQuery] = useState("");
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

  return (
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
                      className={`${styles.emojiBtn} ${activeValue === e.char ? styles.emojiBtnActive : ""}`}
                      onClick={() => onPick(e.char)}
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
                className={`${styles.emojiBtn} ${activeValue === e.mention ? styles.emojiBtnActive : ""}`}
                onClick={() => onPick(e.mention)}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={e.url} alt={e.name} className={styles.customImg} />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

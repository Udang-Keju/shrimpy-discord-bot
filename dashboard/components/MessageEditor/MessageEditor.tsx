"use client";

import { useRef } from "react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import type { DiscordEmoji } from "@/lib/api";
import EmojiInsertButton from "@/components/EmojiPicker/EmojiInsertButton";
import { useEmojiInsert } from "@/components/EmojiPicker/useEmojiInsert";
import { MessageDraft } from "./messageDraft";

interface MessageEditorProps {
  draft: MessageDraft;
  onChange: (draft: MessageDraft) => void;
  customEmojis: DiscordEmoji[];
}

/** Form fields for composing a Discord message: plain text plus an optional embed.
 *  Pair with <MessagePreview> to show the result. */
export default function MessageEditor({ draft, onChange, customEmojis }: MessageEditorProps) {
  const contentRef = useRef<HTMLTextAreaElement>(null);
  const descRef = useRef<HTMLTextAreaElement>(null);

  const set = <K extends keyof MessageDraft>(key: K) => (value: MessageDraft[K]) =>
    onChange({ ...draft, [key]: value });
  const insertContentEmoji = useEmojiInsert(contentRef, draft.content, set("content"));
  const insertDescEmoji = useEmojiInsert(descRef, draft.embedDescription, set("embedDescription"));

  const text = (key: keyof MessageDraft, label: string) => (
    <div className={styles.formGroup}>
      <label className={styles.label}>{label}</label>
      <input className={styles.input} type="text" value={draft[key]} onChange={e => set(key)(e.target.value)} />
    </div>
  );

  return (
    <div className={styles.fieldGroup}>
      <p className={styles.fieldGroupTitle}>Message</p>
      <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 4px' }}>
        Send plain text, an embed, or both. Leave the embed title &amp; description empty to send plain text only.
      </p>
      <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 4px' }}>
        The <strong>Plain Text Message</strong> and <strong>Embed Description</strong> support Discord
        markdown (<code>**bold**</code>, <code>*italic*</code>, <code>`code`</code>, lists, and masked
        links <code>[text](url)</code>). The embed <strong>title</strong>, <strong>author</strong>, and
        <strong> footer</strong> are plain text only.
      </p>

      <div className={styles.formGroup}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <label className={styles.label}>Plain Text Message (optional)</label>
          <EmojiInsertButton onSelect={insertContentEmoji} customEmojis={customEmojis} />
        </div>
        <textarea ref={contentRef} className={styles.textarea} rows={2} value={draft.content} onChange={e => set("content")(e.target.value)} placeholder="Sent as the message's own text, above any embed." />
      </div>

      {text("embedTitle", "Embed Title")}

      <div className={styles.formGroup}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <label className={styles.label}>Embed Description</label>
          <EmojiInsertButton onSelect={insertDescEmoji} customEmojis={customEmojis} />
        </div>
        <textarea ref={descRef} className={styles.textarea} rows={3} value={draft.embedDescription} onChange={e => set("embedDescription")(e.target.value)} />
      </div>

      <div className={styles.formGroup}>
        <label className={styles.label}>Embed Color</label>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <input type="color" value={draft.embedColor} onChange={e => set("embedColor")(e.target.value)} style={{ width: '40px', height: '36px', padding: '2px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', background: 'none', cursor: 'pointer' }} />
          <input className={styles.input} type="text" value={draft.embedColor} onChange={e => set("embedColor")(e.target.value)} style={{ flex: 1 }} />
        </div>
      </div>

      <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
        {text("authorName", "Author Name")}
        {text("authorIconUrl", "Author Icon URL")}
      </div>
      <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
        {text("thumbnailUrl", "Thumbnail URL")}
        {text("imageUrl", "Main Image URL")}
      </div>
      <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
        {text("footerText", "Footer Text")}
        {text("footerIconUrl", "Footer Icon URL")}
      </div>
    </div>
  );
}

"use client";

import { ReactNode } from "react";
import { Eye } from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import EmojiText from "@/components/EmojiView/EmojiText";
import { MessageDraft, draftHasEmbed, draftMedia, isImageUrl } from "./messageDraft";

interface MessagePreviewProps {
  draft: MessageDraft;
  /** Rendered under the message, e.g. buttons or reaction chips. */
  children?: ReactNode;
}

/** Discord-style rendering of a MessageDraft (plain text above an optional embed). */
export default function MessagePreview({ draft, children }: MessagePreviewProps) {
  const media = draftMedia(draft);
  const hasEmbed = draftHasEmbed(draft);

  return (
    <div className={styles.card}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
        <Eye size={16} style={{ color: 'var(--color-accent)' }} />
        <h3 className={styles.cardTitle}>Real-time Discord Preview</h3>
      </div>

      <div style={{ background: '#36393f', border: '1px solid #202225', padding: '16px', borderRadius: '8px' }}>
        {draft.content && (
          <div style={{ color: '#dcddde', fontSize: '14px', whiteSpace: 'pre-wrap', lineHeight: '1.4', marginBottom: hasEmbed ? '10px' : 0 }}>
            <EmojiText text={draft.content} size={18} />
          </div>
        )}

        {hasEmbed && (
          <div style={{ background: '#2f3136', borderLeft: `4px solid ${draft.embedColor}`, borderRadius: '4px', padding: '16px', width: '100%', display: 'flex', gap: '12px' }}>
            <div style={{ flex: 1 }}>
              {media?.author?.name && (
                <div style={{ color: '#ffffff', fontSize: '12px', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  {isImageUrl(media.author.iconUrl) && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={media.author.iconUrl} alt="" style={{ width: '20px', height: '20px', borderRadius: '50%' }} />
                  )}
                  <span>{media.author.name}</span>
                </div>
              )}
              {draft.embedTitle && (
                <div style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '15px', marginBottom: '8px' }}>{draft.embedTitle}</div>
              )}
              {draft.embedDescription && (
                <div style={{ color: '#dcddde', fontSize: '13px', whiteSpace: 'pre-wrap', lineHeight: '1.4' }}>
                  <EmojiText text={draft.embedDescription} size={16} />
                </div>
              )}
              {isImageUrl(media?.image?.url) && (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={media!.image!.url} alt="" style={{ maxWidth: '100%', borderRadius: '4px', marginTop: '10px' }} />
              )}
              {media?.footer?.text && (
                <div style={{ color: '#72767d', fontSize: '11px', marginTop: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  {isImageUrl(media.footer.iconUrl) && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={media.footer.iconUrl} alt="" style={{ width: '16px', height: '16px', borderRadius: '50%' }} />
                  )}
                  <span>{media.footer.text}</span>
                </div>
              )}
            </div>
            {isImageUrl(media?.thumbnail?.url) && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={media!.thumbnail!.url} alt="" style={{ width: '64px', height: '64px', borderRadius: '4px', objectFit: 'cover', flexShrink: 0 }} />
            )}
          </div>
        )}

        {!draft.content && !hasEmbed && (
          <div style={{ color: '#72767d', fontSize: '13px', fontStyle: 'italic' }}>Nothing to show yet. Add plain text or an embed.</div>
        )}

        {children}
      </div>
    </div>
  );
}

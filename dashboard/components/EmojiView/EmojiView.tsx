"use client";

// Mirrors internal/pkg/discordutil.ParseComponentEmoji on the frontend: renders a
// stored emoji string as either a custom-emoji image or a plain unicode glyph.
//
// Accepts the custom-emoji mention forms (<:name:id> / <a:name:id>), the reaction
// API form (name:id — how reaction-role mappings come back from the API), or a raw
// unicode emoji.
export const MENTION_RE = /^<(a)?:([a-zA-Z0-9_]+):(\d+)>$/;
const REACTION_RE = /^([a-zA-Z0-9_]+):(\d+)$/;

/** CDN image URL for a custom emoji string, or null for unicode/plain text. */
export function emojiImageUrl(emoji: string): string | null {
  const mention = MENTION_RE.exec(emoji);
  if (mention) {
    return `https://cdn.discordapp.com/emojis/${mention[3]}.${mention[1] ? "gif" : "png"}`;
  }
  const reaction = REACTION_RE.exec(emoji);
  if (reaction) {
    // The reaction API form drops the animated flag, so default to a static frame.
    return `https://cdn.discordapp.com/emojis/${reaction[2]}.png`;
  }
  return null;
}

interface EmojiViewProps {
  emoji: string;
  size?: number;
}

export default function EmojiView({ emoji, size = 18 }: EmojiViewProps) {
  const url = emojiImageUrl(emoji);
  if (url) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={url}
        alt=""
        width={size}
        height={size}
        style={{ objectFit: "contain", verticalAlign: "middle", display: "inline-block" }}
      />
    );
  }
  return <span style={{ fontSize: size, lineHeight: 1 }}>{emoji}</span>;
}

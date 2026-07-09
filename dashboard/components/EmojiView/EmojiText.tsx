"use client";

import { Fragment } from "react";
import EmojiView from "./EmojiView";

// Global variant of EmojiView's MENTION_RE so we can split a full string into text +
// custom-emoji tokens. Unicode emoji already render as text, so we only special-case mentions.
const MENTION_TOKEN_RE = /<a?:[a-zA-Z0-9_]+:\d+>/g;

interface EmojiTextProps {
  text: string;
  size?: number;
}

/** Renders a string with inline custom-emoji mentions (<:name:id>) as CDN images, leaving
 *  unicode emoji and plain text untouched. Used in dashboard previews for content/description. */
export default function EmojiText({ text, size = 18 }: EmojiTextProps) {
  if (!text) return null;

  const nodes: React.ReactNode[] = [];
  let last = 0;
  let idx = 0;
  for (const m of text.matchAll(MENTION_TOKEN_RE)) {
    const start = m.index ?? 0;
    if (start > last) nodes.push(<Fragment key={idx++}>{text.slice(last, start)}</Fragment>);
    nodes.push(<EmojiView key={idx++} emoji={m[0]} size={size} />);
    last = start + m[0].length;
  }
  if (last < text.length) nodes.push(<Fragment key={idx++}>{text.slice(last)}</Fragment>);

  return <>{nodes}</>;
}

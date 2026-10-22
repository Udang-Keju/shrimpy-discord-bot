import type { EmbedMedia } from "@/lib/api";

/** Editable state of a Discord message: optional plain text plus an optional embed.
 *  Shared by every screen that lets admins compose a message (reaction roles, ...). */
export interface MessageDraft {
  content: string;
  embedTitle: string;
  embedDescription: string;
  embedColor: string; // "#rrggbb"
  authorName: string;
  authorIconUrl: string;
  thumbnailUrl: string;
  imageUrl: string;
  footerText: string;
  footerIconUrl: string;
}

export const DEFAULT_EMBED_COLOR = "#5865F2";

export function createDraft(overrides: Partial<MessageDraft> = {}): MessageDraft {
  return {
    content: "",
    embedTitle: "",
    embedDescription: "",
    embedColor: DEFAULT_EMBED_COLOR,
    authorName: "",
    authorIconUrl: "",
    thumbnailUrl: "",
    imageUrl: "",
    footerText: "",
    footerIconUrl: "",
    ...overrides,
  };
}

export function colorToHex(n?: number): string {
  if (n === undefined || n === null) return DEFAULT_EMBED_COLOR;
  return "#" + Math.max(0, Math.min(0xffffff, n)).toString(16).padStart(6, "0");
}

export function hexToColor(hex: string): number {
  return parseInt(hex.replace("#", ""), 16) || 0;
}

/** Guards preview <img> tags: a half-typed, non-absolute URL would otherwise resolve
 *  relative to the current page and fire a real (404) request on every keystroke. */
export function isImageUrl(url?: string): boolean {
  return !!url && /^https?:\/\//i.test(url);
}

/** The embed media object for a draft, or undefined when no media field is set. */
export function draftMedia(d: MessageDraft): EmbedMedia | undefined {
  if (!(d.authorName || d.thumbnailUrl || d.imageUrl || d.footerText)) return undefined;
  return {
    author: d.authorName ? { name: d.authorName, iconUrl: d.authorIconUrl || undefined } : undefined,
    thumbnail: d.thumbnailUrl ? { url: d.thumbnailUrl } : undefined,
    image: d.imageUrl ? { url: d.imageUrl } : undefined,
    footer: d.footerText ? { text: d.footerText, iconUrl: d.footerIconUrl || undefined } : undefined,
  };
}

/** True when the message would render an embed (matches the backend's embed rules). */
export function draftHasEmbed(d: MessageDraft): boolean {
  return !!(d.embedTitle || d.embedDescription || draftMedia(d));
}

/** True when there is anything at all to post. */
export function draftHasContent(d: MessageDraft): boolean {
  return !!d.content.trim() || draftHasEmbed(d);
}

/** Message fields in the shape the API's embed-capable endpoints share. */
export function draftToFields(d: MessageDraft) {
  const hasEmbed = draftHasEmbed(d);
  return {
    content: d.content || undefined,
    embedTitle: d.embedTitle || undefined,
    embedDescription: d.embedDescription || undefined,
    embedColor: hasEmbed ? hexToColor(d.embedColor) : undefined,
    embedMedia: draftMedia(d),
  };
}

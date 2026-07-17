// dashboard/app/dashboard/[guildId]/translation/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Save, Plus, Trash2 } from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, TranslationConfig, DiscordChannel, DiscordEmoji } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import EmojiInsertButton from "@/components/EmojiPicker/EmojiInsertButton";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

// Common target languages. Codes are ISO 639-1; the selected engine decides
// which it actually supports.
const LANGUAGES: { value: string; label: string }[] = [
  { value: "en", label: "English" },
  { value: "es", label: "Spanish" },
  { value: "fr", label: "French" },
  { value: "de", label: "German" },
  { value: "pt", label: "Portuguese" },
  { value: "it", label: "Italian" },
  { value: "nl", label: "Dutch" },
  { value: "pl", label: "Polish" },
  { value: "ru", label: "Russian" },
  { value: "uk", label: "Ukrainian" },
  { value: "tr", label: "Turkish" },
  { value: "ja", label: "Japanese" },
  { value: "ko", label: "Korean" },
  { value: "zh", label: "Chinese" },
  { value: "id", label: "Indonesian" },
  { value: "ar", label: "Arabic" },
];

const PROVIDERS: { value: string; label: string }[] = [
  { value: "deepl", label: "DeepL" },
  { value: "google", label: "Google Translate" },
  { value: "libretranslate", label: "LibreTranslate (self-hosted)" },
];

const langLabel = (code: string | null) =>
  code ? LANGUAGES.find(l => l.value === code)?.label ?? code : "Server default";

// Convert a picked emoji into the identifier the bot matches on: custom emoji
// mentions (<:name:id> / <a:name:id>) become "name:id"; unicode stays as-is.
function normalizeEmoji(raw: string): string {
  const m = raw.match(/^<a?:(\w+):(\d+)>$/);
  return m ? `${m[1]}:${m[2]}` : raw.trim();
}

// Render a stored emoji identifier back for display: "name:id" shows as :name:.
function displayEmoji(stored: string): string {
  const m = stored.match(/^(\w+):(\d+)$/);
  return m ? `:${m[1]}:` : stored;
}

export default function TranslationPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast } = useToast();

  const [config, setConfig] = useState<TranslationConfig | null>(null);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [customEmojis, setCustomEmojis] = useState<DiscordEmoji[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const [selectedChannel, setSelectedChannel] = useState("");
  const [channelOverride, setChannelOverride] = useState("");
  const [emojiInput, setEmojiInput] = useState("");
  const [emojiOverride, setEmojiOverride] = useState("");

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [confData, chansData, emojiData] = await Promise.all([
          ShrimpyAPI.getTranslationConfig(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordEmojis(guildId),
        ]);
        setConfig(confData);
        setChannels(chansData);
        setCustomEmojis(emojiData);
        if (chansData.length > 0) setSelectedChannel(chansData[0].id);
      } catch (err) {
        console.error(err);
        showToast("Failed to load translation settings.", "error");
      } finally {
        setLoading(false);
      }
    }
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [guildId]);

  const updateField = <K extends keyof TranslationConfig>(key: K, val: TranslationConfig[K]) => {
    setConfig(prev => (prev ? { ...prev, [key]: val } : null));
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config) return;
    setSaving(true);
    try {
      const saved = await ShrimpyAPI.saveTranslationConfig(guildId, config);
      setConfig(prev => (prev ? { ...saved, channels: prev.channels, emojis: prev.emojis } : saved));
      showToast("Translation settings saved.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to save translation settings.", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleAddChannel = async () => {
    if (!config || !selectedChannel) return;
    if (config.channels.some(c => c.channelId === selectedChannel)) {
      showToast("Channel is already configured.", "warning");
      return;
    }
    const override = channelOverride || null;
    try {
      await ShrimpyAPI.addTranslationChannel(guildId, selectedChannel, override);
      updateField("channels", [...config.channels, { channelId: selectedChannel, targetLangOverride: override }]);
      showToast("Channel added.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to add channel.", "error");
    }
  };

  const handleRemoveChannel = async (channelId: string) => {
    if (!config) return;
    try {
      await ShrimpyAPI.removeTranslationChannel(guildId, channelId);
      updateField("channels", config.channels.filter(c => c.channelId !== channelId));
    } catch (err) {
      console.error(err);
      showToast("Failed to remove channel.", "error");
    }
  };

  const handleAddEmoji = async (raw: string) => {
    if (!config) return;
    const emoji = normalizeEmoji(raw);
    if (!emoji) return;
    if (config.emojis.some(e => e.emoji === emoji)) {
      showToast("Emoji is already configured.", "warning");
      return;
    }
    const override = emojiOverride || null;
    try {
      await ShrimpyAPI.addTranslationEmoji(guildId, emoji, override);
      updateField("emojis", [...config.emojis, { emoji, targetLangOverride: override }]);
      setEmojiInput("");
      showToast("Trigger emoji added.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to add emoji.", "error");
    }
  };

  const handleRemoveEmoji = async (emoji: string) => {
    if (!config) return;
    try {
      await ShrimpyAPI.removeTranslationEmoji(guildId, emoji);
      updateField("emojis", config.emojis.filter(e => e.emoji !== emoji));
    } catch (err) {
      console.error(err);
      showToast("Failed to remove emoji.", "error");
    }
  };

  const overrideOptions = [{ value: "", label: "Server default" }, ...LANGUAGES];
  const textChannels = channels.filter(c => c.type === "text" || c.type === "0" || !c.type);

  if (loading) return (
    <div>
      <SkeletonHeader />
      <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)", maxWidth: 720 }}>
        <SkeletonCard fields={4} />
        <SkeletonCard fields={3} />
        <SkeletonCard fields={3} />
      </div>
    </div>
  );
  if (!config) return null;

  const disabled = !config.enabled;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Message Translation</h2>
        <p className={styles.sectionDesc}>
          Automatically translate member messages — in selected channels, or when a member reacts
          with a configured emoji.
        </p>
      </div>

      <form onSubmit={handleSave} style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)", maxWidth: 720 }}>

        {/* Master toggle */}
        <div className={styles.card}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <div style={{ fontSize: "var(--text-sm)", fontWeight: "bold" }}>Enable Translation</div>
              <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
                Master switch. When off, no messages are translated regardless of the triggers below.
              </div>
            </div>
            <label className={styles.toggle}>
              <input type="checkbox" checked={config.enabled} onChange={e => updateField("enabled", e.target.checked)} />
              <span className={styles.slider}></span>
            </label>
          </div>
        </div>

        {/* Engine + credentials */}
        <div className={styles.card} style={{ opacity: disabled ? 0.6 : 1 }}>
          <h3 className={styles.cardTitle}>Translation Engine</h3>
          <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
            <div className={styles.formGroup}>
              <label className={styles.label}>Engine</label>
              <Dropdown
                value={config.provider}
                onChange={val => updateField("provider", val)}
                options={PROVIDERS}
              />
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>API Key</label>
              <input
                className={styles.input}
                type="password"
                value={config.apiKey}
                onChange={e => updateField("apiKey", e.target.value)}
                placeholder={config.hasApiKey ? "•••••••• (stored)" : "Paste your engine API key"}
                autoComplete="off"
              />
              <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 }}>
                Stored encrypted. Leave as-is to keep the current key.
              </div>
            </div>

            {config.provider === "libretranslate" && (
              <div className={styles.formGroup}>
                <label className={styles.label}>Instance URL</label>
                <input
                  className={styles.input}
                  type="text"
                  value={config.endpointUrl ?? ""}
                  onChange={e => updateField("endpointUrl", e.target.value || null)}
                  placeholder="https://libretranslate.example.com"
                />
              </div>
            )}

            <div className={styles.formGroup}>
              <label className={styles.label}>Default Target Language</label>
              <Dropdown
                value={config.targetLang ?? ""}
                onChange={val => updateField("targetLang", val || null)}
                placeholder="Use server language"
                options={[{ value: "", label: "Use server language" }, ...LANGUAGES]}
              />
            </div>
          </div>
        </div>

        {/* Auto-translate channels */}
        <div className={styles.card} style={{ opacity: disabled ? 0.6 : 1 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-2)" }}>
            <div>
              <h3 className={styles.cardTitle} style={{ marginBottom: 2 }}>Auto-translate Channels</h3>
              <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
                Every member message in these channels is translated automatically.
              </div>
            </div>
            <label className={styles.toggle}>
              <input type="checkbox" checked={config.autoEnabled} onChange={e => updateField("autoEnabled", e.target.checked)} />
              <span className={styles.slider}></span>
            </label>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: "8px", margin: "8px 0" }}>
            {config.channels.length === 0 ? (
              <div style={{ color: "var(--color-text-muted)", fontSize: "12px" }}>No channels configured yet.</div>
            ) : (
              config.channels.map(ch => {
                const matched = channels.find(c => c.id === ch.channelId);
                return (
                  <div key={ch.channelId} className={styles.actionBtn} style={{ justifyContent: "space-between", cursor: "default" }}>
                    <span style={{ fontWeight: "bold" }}>#{matched?.name || ch.channelId}</span>
                    <span style={{ display: "flex", alignItems: "center", gap: 8 }}>
                      <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>→ {langLabel(ch.targetLangOverride)}</span>
                      <button type="button" onClick={() => handleRemoveChannel(ch.channelId)}
                        style={{ background: "none", border: "none", color: "var(--color-danger)", cursor: "pointer" }}>
                        <Trash2 size={12} />
                      </button>
                    </span>
                  </div>
                );
              })
            )}
          </div>

          <div style={{ display: "flex", gap: "8px", borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-4)", marginTop: "var(--space-2)" }}>
            <Dropdown
              value={selectedChannel}
              onChange={setSelectedChannel}
              placeholder="Select a channel..."
              options={textChannels.map(c => ({ value: c.id, label: `#${c.name}` }))}
              style={{ flex: 1 }}
            />
            <Dropdown
              value={channelOverride}
              onChange={setChannelOverride}
              options={overrideOptions}
              style={{ width: 160 }}
            />
            <button type="button" onClick={handleAddChannel} className={styles.actionBtn} style={{ padding: "10px 16px", display: "flex", alignItems: "center", gap: 4 }}>
              <Plus size={14} /><span>Add</span>
            </button>
          </div>
        </div>

        {/* Reaction trigger emojis */}
        <div className={styles.card} style={{ opacity: disabled ? 0.6 : 1 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-2)" }}>
            <div>
              <h3 className={styles.cardTitle} style={{ marginBottom: 2 }}>Reaction Trigger Emojis</h3>
              <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
                When a member reacts to any message with one of these emojis, the bot translates it.
              </div>
            </div>
            <label className={styles.toggle}>
              <input type="checkbox" checked={config.reactionEnabled} onChange={e => updateField("reactionEnabled", e.target.checked)} />
              <span className={styles.slider}></span>
            </label>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: "8px", margin: "8px 0" }}>
            {config.emojis.length === 0 ? (
              <div style={{ color: "var(--color-text-muted)", fontSize: "12px" }}>No trigger emojis configured yet.</div>
            ) : (
              config.emojis.map(e => (
                <div key={e.emoji} className={styles.actionBtn} style={{ justifyContent: "space-between", cursor: "default" }}>
                  <span style={{ fontWeight: "bold" }}>{displayEmoji(e.emoji)}</span>
                  <span style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>→ {langLabel(e.targetLangOverride)}</span>
                    <button type="button" onClick={() => handleRemoveEmoji(e.emoji)}
                      style={{ background: "none", border: "none", color: "var(--color-danger)", cursor: "pointer" }}>
                      <Trash2 size={12} />
                    </button>
                  </span>
                </div>
              ))
            )}
          </div>

          <div style={{ display: "flex", gap: "8px", borderTop: "1px solid var(--color-border)", paddingTop: "var(--space-4)", marginTop: "var(--space-2)", alignItems: "center" }}>
            <div style={{ display: "flex", alignItems: "center", gap: 6, flex: 1 }}>
              <input
                className={styles.input}
                type="text"
                value={emojiInput}
                onChange={ev => setEmojiInput(ev.target.value)}
                placeholder="Paste an emoji, e.g. 🇫🇷"
                style={{ flex: 1 }}
              />
              <EmojiInsertButton onSelect={v => setEmojiInput(v)} customEmojis={customEmojis} />
            </div>
            <Dropdown
              value={emojiOverride}
              onChange={setEmojiOverride}
              options={overrideOptions}
              style={{ width: 160 }}
            />
            <button type="button" onClick={() => handleAddEmoji(emojiInput)} className={styles.actionBtn} style={{ padding: "10px 16px", display: "flex", alignItems: "center", gap: 4 }} disabled={!emojiInput}>
              <Plus size={14} /><span>Add</span>
            </button>
          </div>
        </div>

        <button type="submit" className={styles.submitBtn} disabled={saving}>
          <Save size={16} />
          <span>{saving ? "Saving..." : "Save Translation Settings"}</span>
        </button>
      </form>
    </div>
  );
}

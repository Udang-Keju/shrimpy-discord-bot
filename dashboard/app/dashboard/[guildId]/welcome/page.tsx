// dashboard/app/dashboard/[guildId]/welcome/page.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import {
  Save,
  Plus,
  Trash2
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, WelcomeConfig, DiscordChannel, DiscordRole, DiscordEmoji } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import EmojiInsertButton from "@/components/EmojiPicker/EmojiInsertButton";
import { useEmojiInsert } from "@/components/EmojiPicker/useEmojiInsert";
import MessageEditor from "@/components/MessageEditor/MessageEditor";
import MessagePreview from "@/components/MessageEditor/MessagePreview";
import { MessageDraft, colorToHex, createDraft, draftHasContent, draftToFields } from "@/components/MessageEditor/messageDraft";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

const defaultDraft = () => createDraft({
  embedTitle: "Welcome to {server}!",
  embedDescription: "Hey {user}, glad you're here! You are member #{membercount}.",
});

// Example values the live preview substitutes for the template variables the bot resolves on join.
const previewVars = (text: string) => text
  .replace(/\{(user|mention)\}/g, "@NewMember")
  .replace(/\{username\}/g, "NewMember")
  .replace(/\{server\}/g, "Shrimpy Sandbox")
  .replace(/\{membercount\}/g, "128");

// A copy of the draft with every text field resolved, for the preview only.
const resolveDraft = (d: MessageDraft): MessageDraft => ({
  ...d,
  content: previewVars(d.content),
  embedTitle: previewVars(d.embedTitle),
  embedDescription: previewVars(d.embedDescription),
  authorName: previewVars(d.authorName),
  footerText: previewVars(d.footerText),
});

const configToDraft = (c: WelcomeConfig): MessageDraft => {
  if (!c.content && !c.title && !c.description && !c.media) return defaultDraft();
  return createDraft({
    content: c.content,
    embedTitle: c.title,
    embedDescription: c.description,
    embedColor: colorToHex(c.color),
    authorName: c.media?.author?.name ?? "",
    authorIconUrl: c.media?.author?.iconUrl ?? "",
    thumbnailUrl: c.media?.thumbnail?.url ?? "",
    imageUrl: c.media?.image?.url ?? "",
    footerText: c.media?.footer?.text ?? "",
    footerIconUrl: c.media?.footer?.iconUrl ?? "",
  });
};

export default function WelcomePage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast } = useToast();

  const [config, setConfig] = useState<WelcomeConfig | null>(null);
  const [draft, setDraft] = useState(defaultDraft);
  // The toggles gate which greetings are sent; the underlying text/channel survive toggling off.
  const [sendChannel, setSendChannel] = useState(false);
  const [sendDm, setSendDm] = useState(false);
  const [channelId, setChannelId] = useState("");
  const [dmText, setDmText] = useState("");
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [customEmojis, setCustomEmojis] = useState<DiscordEmoji[]>([]);
  const dmTextRef = useRef<HTMLTextAreaElement>(null);
  const [autoRoles, setAutoRoles] = useState<string[]>([]);
  const [selectedAutoRole, setSelectedAutoRole] = useState("");
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [confData, chansData, rolesData, guildData, emojiData] = await Promise.all([
          ShrimpyAPI.getWelcomeConfig(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordRoles(guildId),
          ShrimpyAPI.getGuildConfig(guildId),
          ShrimpyAPI.getDiscordEmojis(guildId)
        ]);
        setConfig(confData);
        setDraft(configToDraft(confData));
        setSendChannel(!!confData.channelId);
        setChannelId(confData.channelId || (chansData[0]?.id ?? ""));
        setSendDm(!!confData.dmMessage);
        setDmText(confData.dmMessage);
        setChannels(chansData);
        setRoles(rolesData);
        setCustomEmojis(emojiData);
        setAutoRoles(guildData.autoRoles);

        if (rolesData.length > 0) {
          setSelectedAutoRole(rolesData[0].id);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [guildId]);

  const handleAddAutoRole = async () => {
    if (!selectedAutoRole) return;
    if (autoRoles.includes(selectedAutoRole)) {
      showToast("Role is already in Auto-Roles list!", "warning");
      return;
    }
    try {
      await ShrimpyAPI.addAutoRole(guildId, selectedAutoRole);
      setAutoRoles(prev => [...prev, selectedAutoRole]);
    } catch (err) {
      console.error(err);
      showToast("Failed to add auto-role.", "error");
    }
  };

  const handleRemoveAutoRole = async (roleId: string) => {
    try {
      await ShrimpyAPI.removeAutoRole(guildId, roleId);
      setAutoRoles(prev => prev.filter(r => r !== roleId));
    } catch (err) {
      console.error(err);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config) return;
    if (sendChannel && !channelId) {
      showToast("Pick a channel for the public greeting!", "warning");
      return;
    }
    if (sendChannel && !draftHasContent(draft)) {
      showToast("Add plain text or embed content to the public greeting!", "warning");
      return;
    }
    if (sendDm && !dmText.trim()) {
      showToast("Enter the direct message text!", "warning");
      return;
    }
    const f = draftToFields(draft);
    const payload: WelcomeConfig = {
      guildId: config.guildId,
      enabled: config.enabled,
      dmMessage: sendDm ? dmText : "",
      channelId: sendChannel ? channelId : "",
      content: f.content ?? "",
      title: f.embedTitle ?? "",
      description: f.embedDescription ?? "",
      color: f.embedColor,
      media: f.embedMedia,
    };
    setSaving(true);
    try {
      await ShrimpyAPI.saveWelcomeConfig(guildId, payload);
      showToast("Welcome settings saved successfully!", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to save welcome settings.", "error");
    } finally {
      setSaving(false);
    }
  };

  const insertDmEmoji = useEmojiInsert(dmTextRef, dmText, setDmText);

  if (loading) return (
    <div>
      <SkeletonHeader />
      <div className={styles.grid}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <SkeletonCard fields={4} />
          <SkeletonCard fields={3} />
          <SkeletonCard fields={2} />
        </div>
        <SkeletonCard fields={5} />
      </div>
    </div>
  );
  if (!config) return null;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Welcome Greetings</h2>
        <p className={styles.sectionDesc}>Customize the public welcome message (plain text and/or embed) and the direct message sent automatically when a user joins the server.</p>
      </div>

      <div className={styles.grid}>
        {/* Configuration settings form */}
        <form onSubmit={handleSave} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          
          {/* Master switch */}
          <div className={styles.card}>
            <div className={styles.formGroupRow} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ fontSize: 'var(--text-sm)', fontWeight: 'bold' }}>Enable welcome greetings</div>
                <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Turn every greeting below on or off without losing your settings.</div>
              </div>
              <label className={styles.toggle}>
                <input
                  type="checkbox"
                  checked={config.enabled}
                  onChange={e => setConfig({ ...config, enabled: e.target.checked })}
                />
                <span className={styles.slider}></span>
              </label>
            </div>
          </div>

          {/* Channel Onboarding Greeting settings */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Public Announcement Greeting</h3>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.formGroupRow} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ fontSize: 'var(--text-sm)', fontWeight: 'bold' }}>Post public greeting</div>
                  <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Send a welcome message to a target guild channel.</div>
                </div>
                <label className={styles.toggle}>
                  <input
                    type="checkbox"
                    checked={sendChannel}
                    onChange={e => setSendChannel(e.target.checked)}
                  />
                  <span className={styles.slider}></span>
                </label>
              </div>

              {sendChannel && (
                <>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Announcement Channel Target</label>
                    <Dropdown
                      value={channelId}
                      onChange={setChannelId}
                      placeholder="Select a channel..."
                      options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                    />
                  </div>

                  <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                    Template variables, replaced when a member joins: <code>{'{user}'}</code> (mention), <code>{'{username}'}</code>,
                    {' '}<code>{'{server}'}</code>, <code>{'{membercount}'}</code>.
                  </p>
                  <MessageEditor draft={draft} onChange={setDraft} customEmojis={customEmojis} />
                </>
              )}
            </div>
          </div>

          {/* Direct Message (DM) Onboarding settings */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Direct Message Welcome</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.formGroupRow} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ fontSize: 'var(--text-sm)', fontWeight: 'bold' }}>Send welcome DM to new user</div>
                  <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Deliver private instructions directly to the joiner.</div>
                </div>
                <label className={styles.toggle}>
                  <input 
                    type="checkbox" 
                    checked={sendDm}
                    onChange={e => setSendDm(e.target.checked)}
                  />
                  <span className={styles.slider}></span>
                </label>
              </div>

              {sendDm && (
                <div className={styles.formGroup}>
                  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <label className={styles.label}>Private Greeting Content</label>
                    <EmojiInsertButton onSelect={insertDmEmoji} customEmojis={customEmojis} />
                  </div>
                  <textarea
                    ref={dmTextRef}
                    className={styles.textarea}
                    rows={4}
                    value={dmText}
                    onChange={e => setDmText(e.target.value)}
                    placeholder="Enter welcome message instructions..."
                    required
                  />
                </div>
              )}
            </div>
          </div>

          {/* Auto-Roles on Join */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Auto-Roles on Join</h3>
            <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
              Roles automatically granted to new members immediately upon entering the server.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
              {autoRoles.length === 0 ? (
                <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No automatic roles configured yet.</div>
              ) : (
                autoRoles.map(rid => {
                  const matched = roles.find(r => r.id === rid);
                  return (
                    <div key={rid} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                      <span style={{ fontWeight: 'bold' }}>{matched?.name || rid}</span>
                      <button
                        type="button"
                        onClick={() => handleRemoveAutoRole(rid)}
                        style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  );
                })
              )}
            </div>

            <div style={{ display: 'flex', gap: '8px', borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-4)', marginTop: 'var(--space-2)' }}>
              <Dropdown
                value={selectedAutoRole}
                onChange={setSelectedAutoRole}
                options={roles.map(r => ({ value: r.id, label: r.name }))}
                style={{ flex: 1 }}
              />
              <button type="button" onClick={handleAddAutoRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                <Plus size={14} />
                <span>Add</span>
              </button>
            </div>
          </div>

          <button type="submit" className={styles.submitBtn} disabled={saving}>
            <Save size={16} />
            <span>{saving ? "Saving..." : "Save Welcome Configurations"}</span>
          </button>
        </form>

        {/* Right Column: Live Discord preview */}
        <div style={{ position: 'sticky', top: '24px', alignSelf: 'start' }}>
          <MessagePreview draft={resolveDraft(draft)} />
        </div>
      </div>
    </div>
  );
}

// dashboard/app/dashboard/[guildId]/roles/page.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import {
  Plus,
  Trash2,
  Send,
  Save,
  Smile,
  AlertTriangle
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, ReactionRole, ReactionRoleMapping, DiscordChannel, DiscordRole, DiscordEmoji } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import EmojiPicker from "@/components/EmojiPicker/EmojiPicker";
import EmojiView from "@/components/EmojiView/EmojiView";
import MessageEditor from "@/components/MessageEditor/MessageEditor";
import MessagePreview from "@/components/MessageEditor/MessagePreview";
import { colorToHex, createDraft, draftHasContent, draftToFields } from "@/components/MessageEditor/messageDraft";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

const defaultDraft = () => createDraft({
  embedTitle: "Select Your Roles",
  embedDescription: "React below to pick up or drop server tags.",
});

// The API returns role IDs only; resolve display names from the guild's roles.
const withRoleNames = (rr: ReactionRole, allRoles: DiscordRole[]): ReactionRole => ({
  ...rr,
  mappings: rr.mappings.map(m => ({
    ...m,
    roleName: allRoles.find(r => r.id === m.roleId)?.name ?? m.roleName ?? m.roleId,
  })),
});

export default function ReactionRolesPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast, updateToast } = useToast();

  const [reactionRoles, setReactionRoles] = useState<ReactionRole[]>([]);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [customEmojis, setCustomEmojis] = useState<DiscordEmoji[]>([]);
  const [loading, setLoading] = useState(true);

  // Form state, shared by Create and Edit mode.
  const [draft, setDraft] = useState(defaultDraft);
  const [newChannelId, setNewChannelId] = useState("");
  const [mappings, setMappings] = useState<ReactionRoleMapping[]>([]);
  // Whether the Create form is open (distinct from editingId, which means Edit mode).
  const [creatingNew, setCreatingNew] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const isSubmitting = useRef(false);

  // Temporary Mapping Adder states
  const [tempEmoji, setTempEmoji] = useState("🦐");
  const [tempRoleId, setTempRoleId] = useState("");

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [rrData, chansData, rolesData, emojiData] = await Promise.all([
          ShrimpyAPI.listReactionRoles(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordRoles(guildId),
          ShrimpyAPI.getDiscordEmojis(guildId)
        ]);
        setReactionRoles(rrData.map(rr => withRoleNames(rr, rolesData)));
        setChannels(chansData);
        setRoles(rolesData);
        setCustomEmojis(emojiData);

        if (chansData.length > 0) {
          setNewChannelId(chansData[0].id);
        }
        if (rolesData.length > 0) {
          setTempRoleId(rolesData[0].id);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [guildId]);

  const resetForm = () => {
    setDraft(defaultDraft());
    setNewChannelId(channels.length > 0 ? channels[0].id : "");
    setMappings([]);
  };

  const handleNewClick = () => {
    resetForm();
    setEditingId(null);
    setCreatingNew(true);
  };

  const handleEditClick = (rr: ReactionRole) => {
    setEditingId(rr.id);
    setCreatingNew(false);
    setNewChannelId(rr.channelId);
    setMappings(rr.mappings);
    setDraft(createDraft({
      content: rr.content ?? "",
      embedTitle: rr.title,
      embedDescription: rr.description,
      embedColor: colorToHex(rr.color),
      authorName: rr.media?.author?.name ?? "",
      authorIconUrl: rr.media?.author?.iconUrl ?? "",
      thumbnailUrl: rr.media?.thumbnail?.url ?? "",
      imageUrl: rr.media?.image?.url ?? "",
      footerText: rr.media?.footer?.text ?? "",
      footerIconUrl: rr.media?.footer?.iconUrl ?? "",
    }));
  };

  const handleCancel = () => {
    setEditingId(null);
    setCreatingNew(false);
    resetForm();
  };

  const handleAddMapping = () => {
    if (!tempRoleId) return;
    const matchedRole = roles.find(r => r.id === tempRoleId);
    if (!matchedRole) return;

    // Check if emoji or role is already mapped
    if (mappings.some(m => m.emoji === tempEmoji || m.roleId === tempRoleId)) {
      showToast("Emoji or Role already mapped in this panel!", "warning");
      return;
    }

    const newMapping: ReactionRoleMapping = {
      emoji: tempEmoji,
      roleId: tempRoleId,
      roleName: matchedRole.name
    };
    setMappings(prev => [...prev, newMapping]);
  };

  const handleRemoveMapping = (idx: number) => {
    setMappings(prev => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!draftHasContent(draft)) {
      showToast("Add plain text or embed content to the message!", "warning");
      return;
    }
    if (mappings.length === 0) {
      showToast("Please add at least one emoji-to-role mapping!", "warning");
      return;
    }
    if (isSubmitting.current) return;
    isSubmitting.current = true;

    const f = draftToFields(draft);
    const snapshot = {
      channelId: newChannelId,
      content: f.content,
      title: f.embedTitle ?? "",
      description: f.embedDescription ?? "",
      color: f.embedColor,
      media: f.embedMedia,
      mappings,
    };
    const editId = editingId;
    const original = editId ? reactionRoles.find(r => r.id === editId) : undefined;
    isSubmitting.current = false;

    if (editId && original) {
      // Editing: keep the form open afterwards so the user can keep refining the message.
      const toastId = showToast("Updating reaction role message...", "loading");
      try {
        const updated = await ShrimpyAPI.updateReactionRole(guildId, editId, snapshot);
        // Reconcile mappings: drop removed ones, then add new or re-pointed ones.
        for (const old of original.mappings) {
          if (!snapshot.mappings.some(m => m.emoji === old.emoji)) {
            await ShrimpyAPI.removeReactionRoleEmoji(guildId, editId, old.emoji);
          }
        }
        for (const m of snapshot.mappings) {
          const prev = original.mappings.find(o => o.emoji === m.emoji);
          if (!prev || prev.roleId !== m.roleId) {
            await ShrimpyAPI.addReactionRoleEmoji(guildId, editId, m.emoji, m.roleId);
          }
        }
        const next = { ...updated, mappings: snapshot.mappings };
        setReactionRoles(prev => prev.map(r => r.id === editId ? next : r));
        updateToast(toastId, "Reaction role message updated!", "success");
      } catch (err) {
        console.error(err);
        updateToast(toastId, "Failed to update reaction role message.", "error");
      }
      return;
    }

    // Creating: reset the form synchronously so the next message can be built while the
    // bot posts it and adds reactions (a multi-round-trip operation).
    resetForm();

    const toastId = showToast("Publishing reaction role panel...", "loading");
    try {
      const newWidget = await ShrimpyAPI.createReactionRole(guildId, snapshot);
      // The create endpoint only posts the message; persist each emoji→role mapping
      // (and add the reaction on Discord) individually.
      for (const m of snapshot.mappings) {
        await ShrimpyAPI.addReactionRoleEmoji(guildId, newWidget.id, m.emoji, m.roleId);
      }
      // Reflect the mappings locally regardless of what the create response echoed back.
      setReactionRoles(prev => [...prev, { ...newWidget, mappings: snapshot.mappings }]);
      updateToast(toastId, "Reaction role panel published successfully!", "success");
    } catch (err) {
      console.error(err);
      updateToast(toastId, "Failed to publish reaction role panel.", "error");
    }
  };

  const handleDeleteRR = async (msgId: string) => {
    try {
      await ShrimpyAPI.deleteReactionRole(guildId, msgId);
      setReactionRoles(prev => prev.filter(r => r.id !== msgId));
      if (editingId === msgId) handleCancel();
      showToast("Reaction role message deleted.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to delete reaction role message.", "error");
    }
  };

  if (loading) return (
    <div>
      <SkeletonHeader />
      <div className={styles.grid}>
        <SkeletonCard fields={4} />
        <SkeletonCard fields={3} />
      </div>
    </div>
  );

  const formOpen = creatingNew || editingId !== null;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Reaction Roles Console</h2>
        <p className={styles.sectionDesc}>Configure self-assignable roles widgets. Users simply react to the posted message to receive tags.</p>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>

        {/* Active Reaction Role Messages List */}
        <div className={styles.card}>
          {reactionRoles.length === 0 ? (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: 'var(--space-3)', padding: 'var(--space-6) var(--space-4)' }}>
              <h3 className={styles.cardTitle}>No reaction role messages yet</h3>
              <p className={styles.sectionDesc} style={{ fontSize: '13px', margin: 0 }}>
                Create one to let members pick up roles by reacting.
              </p>
              <button onClick={handleNewClick} className={styles.submitBtn} style={{ marginTop: 'var(--space-2)' }}>
                <Plus size={16} />
                <span>New Message</span>
              </button>
            </div>
          ) : (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-4)' }}>
                <div>
                  <h3 className={styles.cardTitle}>Active Reaction Role Messages</h3>
                  <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                    Each message is listening for reactions on Discord. Click one to edit it; deleting it stops role assignment.
                  </p>
                </div>
                <button onClick={handleNewClick} className={styles.actionBtn} style={{ flexShrink: 0 }}>
                  <Plus size={14} />
                  <span>New Message</span>
                </button>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {reactionRoles.map(rr => (
                  <div
                    key={rr.id}
                    className={styles.actionBtn}
                    style={{
                      justifyContent: 'space-between',
                      borderColor: editingId === rr.id ? 'var(--color-primary)' : '',
                      background: editingId === rr.id ? 'var(--primary-muted)' : '',
                    }}
                    onClick={() => handleEditClick(rr)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                      <Smile size={14} style={{ color: 'var(--color-primary)' }} />
                      <span style={{ fontWeight: 'bold' }}>{rr.title || rr.content || 'Untitled message'}</span>
                      <span style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>
                        in #{channels.find(c => c.id === rr.channelId)?.name || rr.channelId}
                      </span>
                      {rr.mappings.map((m, idx) => (
                        <span key={idx} style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px' }}>
                          <EmojiView emoji={m.emoji} size={14} />
                          <span>{m.roleName}</span>
                        </span>
                      ))}
                    </div>
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDeleteRR(rr.id); }}
                      style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                      title="Delete Message"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Creator / Editor form, paired with the real-time preview */}
        {formOpen && (
          <div className={styles.grid} style={{ alignItems: 'start' }}>
            <div className={styles.card}>
              <h3 className={styles.cardTitle}>{editingId ? 'Edit Reaction Role Message' : 'Deploy Reaction Role Widget'}</h3>
              <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Destination Channel</label>
                  <Dropdown
                    value={newChannelId}
                    onChange={setNewChannelId}
                    options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                  />
                  {editingId && (
                    <p className={styles.sectionDesc} style={{ fontSize: '11px', margin: '4px 0 0' }}>
                      Changing the channel reposts the message there and deletes the old one.
                    </p>
                  )}
                </div>

                <MessageEditor draft={draft} onChange={setDraft} customEmojis={customEmojis} />

                {/* Mapping editor block */}
                <div style={{ border: '1px solid var(--color-border)', borderRadius: '8px', padding: '16px', backgroundColor: 'rgba(0,0,0,0.1)' }}>
                  <div className={styles.label} style={{ marginBottom: '12px' }}>Role-to-Emoji Mappings</div>

                  {/* Active Mappings List */}
                  {mappings.length === 0 ? (
                    <div style={{ fontSize: '12px', color: 'var(--color-text-muted)', marginBottom: '12px' }}>No mappings added yet. Configure one below.</div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '16px' }}>
                      {mappings.map((m, idx) => (
                        <div key={idx} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <EmojiView emoji={m.emoji} size={16} />
                            <span style={{ fontWeight: 'bold' }}>{m.roleName}</span>
                          </div>
                          <button
                            type="button"
                            onClick={() => handleRemoveMapping(idx)}
                            style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                          >
                            <Trash2 size={12} />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Adding Tool Row */}
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'flex-end', flexWrap: 'wrap' }}>
                    <div className={styles.formGroup} style={{ width: '80px' }}>
                      <label className={styles.label} style={{ fontSize: '10px' }}>Emoji</label>
                      <EmojiPicker
                        value={tempEmoji}
                        onChange={setTempEmoji}
                        customEmojis={customEmojis}
                      />
                    </div>

                    <div className={styles.formGroup} style={{ flex: 1, minWidth: '150px' }}>
                      <label className={styles.label} style={{ fontSize: '10px' }}>Role</label>
                      <Dropdown
                        value={tempRoleId}
                        onChange={setTempRoleId}
                        options={roles.map(r => ({ value: r.id, label: r.name }))}
                      />
                    </div>

                    <button
                      type="button"
                      onClick={handleAddMapping}
                      className={styles.actionBtn}
                      style={{ height: '38px', padding: '0 16px', display: 'flex', alignItems: 'center' }}
                    >
                      <Plus size={14} />
                      <span>Add</span>
                    </button>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: 'var(--space-3)' }}>
                  <button type="submit" className={styles.submitBtn} style={{ flex: 1 }}>
                    {editingId ? <Save size={16} /> : <Send size={16} />}
                    <span>{editingId ? 'Save Changes' : 'Publish Role Panel'}</span>
                  </button>
                  <button type="button" className={styles.actionBtn} onClick={handleCancel}>
                    Cancel
                  </button>
                </div>
              </form>
            </div>

            <MessagePreview draft={draft}>
              {mappings.length > 0 && (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', marginTop: '10px' }}>
                  {mappings.map(m => (
                    <span key={m.emoji} style={{ background: '#2f3136', border: '1px solid #4f545c', borderRadius: '8px', padding: '2px 8px', display: 'flex', alignItems: 'center', gap: '4px', color: '#dcddde', fontSize: '12px' }}>
                      <EmojiView emoji={m.emoji} size={14} />1
                    </span>
                  ))}
                </div>
              )}
            </MessagePreview>
          </div>
        )}

        <div className={styles.card} style={{ backgroundColor: 'rgba(255, 179, 71, 0.05)', borderColor: 'rgba(255, 179, 71, 0.2)' }}>
          <div style={{ display: 'flex', gap: '10px', alignItems: 'flex-start' }}>
            <AlertTriangle size={18} style={{ color: 'var(--color-warning)', flexShrink: 0, marginTop: '2px' }} />
            <div>
              <div style={{ fontWeight: 'bold', fontSize: '13px', color: 'var(--color-warning)' }}>Gateway Requirements</div>
              <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginTop: '4px', lineHeight: '1.4' }}>
                Ensure the Shrimpy Bot role is ordered **above** all target roles in your Server Settings on Discord, otherwise the API encounters a permission constraint error during reactions mapping assignment.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

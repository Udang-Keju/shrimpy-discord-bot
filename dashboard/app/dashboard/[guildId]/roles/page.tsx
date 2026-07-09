// dashboard/app/dashboard/[guildId]/roles/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Plus,
  Trash2,
  Send,
  AlertTriangle
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, ReactionRole, ReactionRoleMapping, DiscordChannel, DiscordRole, DiscordEmoji } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import EmojiPicker from "@/components/EmojiPicker/EmojiPicker";
import EmojiView from "@/components/EmojiView/EmojiView";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

export default function ReactionRolesPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast, updateToast } = useToast();

  const [reactionRoles, setReactionRoles] = useState<ReactionRole[]>([]);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [customEmojis, setCustomEmojis] = useState<DiscordEmoji[]>([]);
  const [loading, setLoading] = useState(true);

  // Form states for new panel
  const [newTitle, setNewTitle] = useState("Select Your Roles");
  const [newDesc, setNewDesc] = useState("React below to pick up or drop server tags.");
  const [newChannelId, setNewChannelId] = useState("");
  const [mappings, setMappings] = useState<ReactionRoleMapping[]>([]);

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
        setReactionRoles(rrData);
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

  const handleCreateRR = async (e: React.FormEvent) => {
    e.preventDefault();
    if (mappings.length === 0) {
      showToast("Please add at least one emoji-to-role mapping!", "warning");
      return;
    }

    // Snapshot then reset the form synchronously so the next panel can be built while
    // the bot posts the embed and adds reactions (a multi-round-trip operation).
    const snapshot = { channelId: newChannelId, title: newTitle, description: newDesc, mappings };
    setNewTitle("Select Your Roles");
    setNewDesc("React below to pick up or drop server tags.");
    setMappings([]);

    const toastId = showToast("Publishing reaction role panel...", "loading");
    try {
      const newWidget = await ShrimpyAPI.createReactionRole(guildId, snapshot);
      // The create endpoint only posts the embed; persist each emoji→role mapping
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
    } catch (err) {
      console.error(err);
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

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Reaction Roles Console</h2>
        <p className={styles.sectionDesc}>Configure self-assignable roles widgets. Users simply react to the posted message to receive tags.</p>
      </div>

      <div className={styles.grid}>
        {/* Left Column: Creator Form */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Deploy Reaction Role Widget</h3>
            <form onSubmit={handleCreateRR} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              
              <div className={styles.formGroup}>
                <label className={styles.label}>Destination Channel</label>
                <Dropdown
                  value={newChannelId}
                  onChange={setNewChannelId}
                  options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                />
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Widget Title Header</label>
                <input className={styles.input} type="text" value={newTitle} onChange={e => setNewTitle(e.target.value)} required />
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Widget Description Text</label>
                <textarea className={styles.textarea} rows={3} value={newDesc} onChange={e => setNewDesc(e.target.value)} required />
              </div>

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

              <button type="submit" className={styles.submitBtn}>
                <Send size={16} />
                <span>Publish Role Panel</span>
              </button>
            </form>
          </div>
        </div>

        {/* Right Column: Active widgets list */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Active Reaction Desks</h3>
            <p className={styles.sectionDesc} style={{ fontSize: '12px', marginBottom: '8px' }}>
              Currently listening message IDs on Discord gateway. Deleting a row stops gateway event mapping checks.
            </p>

            {reactionRoles.length === 0 ? (
              <div style={{ padding: 'var(--space-4) 0', color: 'var(--color-text-muted)', fontSize: '13px' }}>
                No active reaction role panels on this server.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
                {reactionRoles.map(rr => (
                  <div 
                    key={rr.id} 
                    style={{
                      border: '1px solid var(--color-border)', 
                      borderRadius: '8px', 
                      padding: '16px', 
                      backgroundColor: 'var(--color-surface-raised)',
                      position: 'relative'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                      <div>
                        <div style={{ fontWeight: 'bold' }}>{rr.title}</div>
                        <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>
                          Message: `{rr.id}` in #{channels.find(c => c.id === rr.channelId)?.name || rr.channelId}
                        </div>
                      </div>
                      <button 
                        onClick={() => handleDeleteRR(rr.id)}
                        style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                        title="Delete Panel"
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>

                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', marginTop: '12px' }}>
                      {rr.mappings.map((m, idx) => (
                        <span 
                          key={idx} 
                          style={{
                            padding: '4px 8px', 
                            background: 'rgba(255,255,255,0.05)', 
                            border: '1px solid var(--color-border)', 
                            borderRadius: '4px', 
                            fontSize: '12px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px'
                          }}
                        >
                          <EmojiView emoji={m.emoji} size={14} />
                          <span style={{ fontWeight: '500' }}>{m.roleName}</span>
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

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
    </div>
  );
}

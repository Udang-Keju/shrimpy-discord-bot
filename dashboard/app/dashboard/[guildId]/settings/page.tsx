// dashboard/app/dashboard/[guildId]/settings/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Save,
  Plus,
  Trash2,
  ShieldCheck
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, Guild, DiscordChannel, DiscordRole } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import { useToast } from "@/hooks/useToast";
import { Skeleton, SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

export default function SettingsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast } = useToast();

  const [config, setConfig] = useState<Guild | null>(null);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);

  // Local state for adding roles
  const [selectedStaffRole, setSelectedStaffRole] = useState("");

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [configData, chansData, rolesData] = await Promise.all([
          ShrimpyAPI.getGuildConfig(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordRoles(guildId)
        ]);
        setConfig(configData);
        setChannels(chansData);
        setRoles(rolesData);
        
        if (rolesData.length > 0) {
          setSelectedStaffRole(rolesData[0].id);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [guildId]);

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config) return;
    setSaving(true);
    try {
      await Promise.all([
        ShrimpyAPI.updateGuildConfig(guildId, {
          prefix: config.prefix,
          logChannelId: config.logChannelId
        }),
        ShrimpyAPI.updateNickname(guildId, config.nickname || null)
      ]);
      showToast("Guild settings saved successfully!", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to save guild settings.", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleAddStaffRole = async () => {
    if (!config || !selectedStaffRole) return;
    if (config.staffRoles.includes(selectedStaffRole)) {
      showToast("Role is already in Staff Dashboard Access list!", "warning");
      return;
    }
    try {
      await ShrimpyAPI.addStaffRole(guildId, selectedStaffRole);
      setConfig(prev => prev ? ({ ...prev, staffRoles: [...prev.staffRoles, selectedStaffRole] }) : null);
    } catch (err) {
      console.error(err);
      showToast("Failed to add staff role.", "error");
    }
  };

  const handleRemoveStaffRole = async (roleId: string) => {
    if (!config) return;
    try {
      await ShrimpyAPI.removeStaffRole(guildId, roleId);
      setConfig(prev => prev ? ({ ...prev, staffRoles: prev.staffRoles.filter(r => r !== roleId) }) : null);
    } catch (err) {
      console.error(err);
    }
  };

  const updateField = <K extends keyof Guild>(key: K, val: Guild[K]) => {
    setConfig(prev => prev ? ({ ...prev, [key]: val }) : null);
  };

  if (loading) return (
    <div>
      <SkeletonHeader />
      <div className={styles.grid}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <SkeletonCard fields={4} />
          <Skeleton height="40px" width="180px" />
        </div>
        <SkeletonCard fields={4} />
      </div>
    </div>
  );
  if (!config) return null;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>General Configurations</h2>
        <p className={styles.sectionDesc}>Manage bot naming, command prefixes, log channels, and dashboard access credentials.</p>
      </div>

      <div className={styles.grid}>
        {/* Left Column: General Configuration Form */}
        <form onSubmit={handleSaveConfig} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Bot Parameters</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.formGroup}>
                <label className={styles.label}>Bot Server Nickname</label>
                <input 
                  className={styles.input} 
                  type="text" 
                  value={config.nickname || ""} 
                  onChange={e => updateField('nickname', e.target.value)} 
                  placeholder="e.g. Shrimpy Helper"
                />
              </div>

              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Command Prefix</label>
                  <input 
                    className={styles.input} 
                    type="text" 
                    value={config.prefix || "!"} 
                    onChange={e => updateField('prefix', e.target.value)} 
                    maxLength={5} 
                    required 
                  />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Max Concurrent Tickets per Member</label>
                  <input 
                    className={styles.input} 
                    type="number" 
                    value={config.ticketLimit || 3} 
                    onChange={e => updateField('ticketLimit', parseInt(e.target.value) || 1)} 
                    min={1} 
                    max={20} 
                    required 
                  />
                </div>
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Bot Log Channel</label>
                <Dropdown
                  value={config.logChannelId || ""}
                  onChange={val => updateField('logChannelId', val)}
                  placeholder="No logging channel selected"
                  options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                />
              </div>
            </div>
          </div>

          <button type="submit" className={styles.submitBtn} disabled={saving}>
            <Save size={16} />
            <span>{saving ? "Saving..." : "Save Bot Settings"}</span>
          </button>
        </form>

        {/* Right Column: Roles & Authentication settings */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          
          {/* Level 2 Staff Roles - Dashboard Access */}
          <div className={styles.card}>
            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
              <ShieldCheck size={18} style={{ color: 'var(--color-primary)' }} />
              <h3 className={styles.cardTitle}>Dashboard Access Roles</h3>
            </div>
            <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
              Users holding these roles can access and manage configurations in this console (Level 2 credentials). This does not affect who is invited to handle individual tickets — configure that per ticket panel.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
              {config.staffRoles.length === 0 ? (
                <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No Level-2 roles added. Server admins always have access.</div>
              ) : (
                config.staffRoles.map(rid => {
                  const matched = roles.find(r => r.id === rid);
                  return (
                    <div key={rid} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                      <span style={{ fontWeight: 'bold' }}>{matched?.name || rid}</span>
                      <button 
                        type="button" 
                        onClick={() => handleRemoveStaffRole(rid)} 
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
                value={selectedStaffRole}
                onChange={setSelectedStaffRole}
                options={roles.map(r => ({ value: r.id, label: r.name }))}
                style={{ flex: 1 }}
              />
              <button onClick={handleAddStaffRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                <Plus size={14} />
                <span>Add</span>
              </button>
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}

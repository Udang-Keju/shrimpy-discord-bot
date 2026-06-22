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
import styles from "@/app/dashboard/dashboard.module.css";
import { ShrimpyAPI, Guild, DiscordChannel, DiscordRole } from "@/lib/api";

export default function SettingsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;

  const [config, setConfig] = useState<Guild | null>(null);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [saving, setSaving] = useState(false);

  // Local states for adding roles
  const [selectedAutoRole, setSelectedAutoRole] = useState("");
  const [selectedStaffRole, setSelectedStaffRole] = useState("");

  useEffect(() => {
    async function loadData() {
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
          setSelectedAutoRole(rolesData[0].id);
          setSelectedStaffRole(rolesData[0].id);
        }
      } catch (err) {
        console.error(err);
      }
    }
    loadData();
  }, [guildId]);

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config) return;
    setSaving(true);
    try {
      await ShrimpyAPI.updateGuildConfig(guildId, config);
      alert("Guild settings saved successfully!");
    } catch (err) {
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const handleAddAutoRole = () => {
    if (!config || !selectedAutoRole) return;
    if (config.autoRoles.includes(selectedAutoRole)) {
      alert("Role is already in Auto-Roles list!");
      return;
    }
    const updatedRoles = [...config.autoRoles, selectedAutoRole];
    setConfig(prev => prev ? ({ ...prev, autoRoles: updatedRoles }) : null);
  };

  const handleRemoveAutoRole = (roleId: string) => {
    if (!config) return;
    const updatedRoles = config.autoRoles.filter(r => r !== roleId);
    setConfig(prev => prev ? ({ ...prev, autoRoles: updatedRoles }) : null);
  };

  const handleAddStaffRole = () => {
    if (!config || !selectedStaffRole) return;
    if (config.staffRoles.includes(selectedStaffRole)) {
      alert("Role is already in Staff Dashboard Access list!");
      return;
    }
    const updatedRoles = [...config.staffRoles, selectedStaffRole];
    setConfig(prev => prev ? ({ ...prev, staffRoles: updatedRoles }) : null);
  };

  const handleRemoveStaffRole = (roleId: string) => {
    if (!config) return;
    const updatedRoles = config.staffRoles.filter(r => r !== roleId);
    setConfig(prev => prev ? ({ ...prev, staffRoles: updatedRoles }) : null);
  };

  const updateField = <K extends keyof Guild>(key: K, val: Guild[K]) => {
    setConfig(prev => prev ? ({ ...prev, [key]: val }) : null);
  };

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
                <select 
                  className={styles.select} 
                  value={config.logChannelId || ""} 
                  onChange={e => updateField('logChannelId', e.target.value)}
                >
                  <option value="">No logging channel selected</option>
                  {channels.map(c => (
                    <option key={c.id} value={c.id}>#{c.name}</option>
                  ))}
                </select>
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
              Users holding these roles can manage support tickets and access configurations in this console (Level 2 credentials).
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
              <select 
                className={styles.select} 
                value={selectedStaffRole} 
                onChange={e => setSelectedStaffRole(e.target.value)}
                style={{ flex: 1 }}
              >
                {roles.map(r => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
              <button onClick={handleAddStaffRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                <Plus size={14} />
                <span>Add</span>
              </button>
            </div>
          </div>

          {/* Join Auto-Roles list */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Auto-Roles on Join</h3>
            <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
              Roles automatically granted to new members immediately upon entering the server.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
              {config.autoRoles.length === 0 ? (
                <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No automatic roles configured yet.</div>
              ) : (
                config.autoRoles.map(rid => {
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
              <select 
                className={styles.select} 
                value={selectedAutoRole} 
                onChange={e => setSelectedAutoRole(e.target.value)}
                style={{ flex: 1 }}
              >
                {roles.map(r => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
              <button onClick={handleAddAutoRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
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

// dashboard/app/dashboard/[guildId]/welcome/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Save,
  Eye,
  Sparkles
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, WelcomeConfig, DiscordChannel } from "@/lib/api";

export default function WelcomePage() {
  const params = useParams();
  const guildId = params?.guildId as string;

  const [config, setConfig] = useState<WelcomeConfig | null>(null);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        const [confData, chansData] = await Promise.all([
          ShrimpyAPI.getWelcomeConfig(guildId),
          ShrimpyAPI.getDiscordChannels(guildId)
        ]);
        setConfig(confData);
        setChannels(chansData);
      } catch (err) {
        console.error(err);
      }
    }
    loadData();
  }, [guildId]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!config) return;
    setSaving(true);
    try {
      await ShrimpyAPI.saveWelcomeConfig(guildId, config);
      alert("Welcome settings saved successfully!");
    } catch (err) {
      console.error(err);
    } finally {
      setSaving(false);
    }
  };

  const updateField = <K extends keyof WelcomeConfig>(key: K, val: WelcomeConfig[K]) => {
    if (!config) return;
    setConfig(prev => prev ? ({ ...prev, [key]: val }) : null);
  };

  if (!config) return null;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Welcome Greetings Onboarding</h2>
        <p className={styles.sectionDesc}>Customize public banner channels and direct messages sent automatically when a user joins the server.</p>
      </div>

      <div className={styles.grid}>
        {/* Configuration settings form */}
        <form onSubmit={handleSave} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          
          {/* Card customization knobs */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Banner Image Knobs</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Avatar Emoji</label>
                  <select 
                    className={styles.select} 
                    value={config.avatarEmoji} 
                    onChange={e => updateField('avatarEmoji', e.target.value)}
                  >
                    <option value="🦐">🦐 Shrimp</option>
                    <option value="🐠">🐠 Fish</option>
                    <option value="🌊">🌊 Wave</option>
                    <option value="🌴">🌴 Palm</option>
                    <option value="🐙">🐙 Octopus</option>
                  </select>
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Banner Theme Style</label>
                  <select 
                    className={styles.select} 
                    value={config.cardStyle} 
                    onChange={e => updateField('cardStyle', e.target.value as 'dark' | 'light')}
                  >
                    <option value="dark">Deep Blue Dark</option>
                    <option value="light">Warm Apricot Light</option>
                  </select>
                </div>
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Banner Greeting Title</label>
                <input 
                  className={styles.input} 
                  type="text" 
                  value={config.welcomeText} 
                  onChange={e => updateField('welcomeText', e.target.value)} 
                  required 
                />
              </div>
            </div>
          </div>

          {/* Channel Onboarding Greeting settings */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Public Announcement Greeting</h3>
            
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.formGroupRow} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <div style={{ fontSize: 'var(--text-sm)', fontWeight: 'bold' }}>Post Public Greeting Card</div>
                  <div style={{ fontSize: '12px', color: 'var(--color-text-muted)' }}>Send welcoming cards to a target guild channel.</div>
                </div>
                <label className={styles.toggle}>
                  <input 
                    type="checkbox" 
                    checked={config.sendChannel} 
                    onChange={e => updateField('sendChannel', e.target.checked)} 
                  />
                  <span className={styles.slider}></span>
                </label>
              </div>

              {config.sendChannel && (
                <div className={styles.formGroup}>
                  <label className={styles.label}>Announcement Channel Target</label>
                  <select 
                    className={styles.select} 
                    value={config.channelId} 
                    onChange={e => updateField('channelId', e.target.value)}
                  >
                    <option value="">Select a channel...</option>
                    {channels.map(c => (
                      <option key={c.id} value={c.id}>#{c.name}</option>
                    ))}
                  </select>
                </div>
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
                    checked={config.sendDm} 
                    onChange={e => updateField('sendDm', e.target.checked)} 
                  />
                  <span className={styles.slider}></span>
                </label>
              </div>

              {config.sendDm && (
                <div className={styles.formGroup}>
                  <label className={styles.label}>Private Greeting Content</label>
                  <textarea 
                    className={styles.textarea} 
                    rows={4} 
                    value={config.dmText} 
                    onChange={e => updateField('dmText', e.target.value)} 
                    placeholder="Enter welcome message instructions..."
                    required 
                  />
                </div>
              )}
            </div>
          </div>

          <button type="submit" className={styles.submitBtn} disabled={saving}>
            <Save size={16} />
            <span>{saving ? "Saving..." : "Save Welcome Configurations"}</span>
          </button>
        </form>

        {/* Right Column: Rendering Preview card */}
        <div>
          <div className={styles.card} style={{ position: 'sticky', top: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: 'var(--space-2)' }}>
              <Eye size={16} style={{ color: 'var(--color-accent)' }} />
              <h3 className={styles.cardTitle}>Live Card Banner Preview</h3>
            </div>

            {/* Embedded customizer widget (reused styling) */}
            <div 
              style={{
                background: config.cardStyle === "light" ? 'linear-gradient(135deg, #fffbeb 0%, #fff8f2 100%)' : 'linear-gradient(135deg, #1e1b4b 0%, #110f24 100%)',
                border: config.cardStyle === "light" ? '1px solid #fed7aa' : '1px solid #312e81',
                borderRadius: '16px',
                padding: '24px',
                textAlign: 'center',
                boxShadow: 'var(--shadow-sm)',
                position: 'relative',
                overflow: 'hidden'
              }}
            >
              <div 
                style={{
                  position: 'absolute', width: '150px', height: '150px',
                  background: 'radial-gradient(circle, rgba(255, 123, 107, 0.2) 0%, transparent 70%)',
                  top: '-50px', left: '-50px', pointerEvents: 'none'
                }}
              ></div>
              
              <div 
                style={{
                  width: '80px', height: '80px', borderRadius: '50%',
                  border: '4px solid var(--color-primary)', margin: '0 auto 16px',
                  background: '#ffedd5', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: '32px', boxShadow: '0 4px 12px rgba(0,0,0,0.15)'
                }}
              >
                {config.avatarEmoji}
              </div>
              
              <h3 style={{ fontSize: '20px', marginBottom: '8px', color: config.cardStyle === "light" ? '#1a0f1f' : '#ffffff' }}>
                {config.welcomeText}
              </h3>
              
              <p style={{ color: config.cardStyle === "light" ? '#7a5c6e' : '#a5b4fc', fontSize: '14px', marginBottom: '16px' }}>
                We are excited to have you here! Feel free to assign yourself some roles.
              </p>
              
              <div style={{ display: 'flex', justifyContent: 'center', gap: '16px', fontSize: '12px', color: config.cardStyle === "light" ? 'var(--color-primary)' : '#6366f1' }}>
                <span>ID: #99318</span>
                <span>•</span>
                <span>Server: Shrimpy Sandbox</span>
              </div>
            </div>

            <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', display: 'flex', alignItems: 'center', gap: '4px', marginTop: 'var(--space-2)' }}>
              <Sparkles size={12} style={{ color: 'var(--color-accent)' }} />
              <span>Renders instantly inside Discord on join event when active.</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

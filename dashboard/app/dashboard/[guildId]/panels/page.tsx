// dashboard/app/dashboard/[guildId]/panels/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Layers,
  Plus,
  Trash2,
  Eye,
  Ticket
} from "lucide-react";
import styles from "@/app/dashboard/dashboard.module.css";
import { ShrimpyAPI, TicketPanel, TicketCategory, DiscordChannel, DiscordRole } from "@/lib/api";

export default function PanelsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;

  const [panels, setPanels] = useState<TicketPanel[]>([]);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [selectedPanel, setSelectedPanel] = useState<TicketPanel | null>(null);
  const [categories, setCategories] = useState<TicketCategory[]>([]);

  // Form states for new panel
  const [newTitle, setNewTitle] = useState("Contact Support Services");
  const [newDesc, setNewDesc] = useState("Click the button below to open a private ticket.");
  const [newBtnLabel, setNewBtnLabel] = useState("Create Ticket");
  const [newBtnStyle, setNewBtnStyle] = useState<'primary' | 'success' | 'danger' | 'secondary'>('primary');
  const [newChannelId, setNewChannelId] = useState("");

  // Form states for new category
  const [newCatName, setNewCatName] = useState("");
  const [newCatChanId, setNewCatChanId] = useState("");
  const [newCatRoleId, setNewCatRoleId] = useState("");

  useEffect(() => {
    async function loadData() {
      try {
        const [panelsData, chansData, rolesData] = await Promise.all([
          ShrimpyAPI.listPanels(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordRoles(guildId)
        ]);
        setPanels(panelsData);
        setChannels(chansData);
        setRoles(rolesData);
        
        if (chansData.length > 0) {
          setNewChannelId(chansData[0].id);
          setNewCatChanId(chansData[0].id);
        }
        if (rolesData.length > 0) {
          setNewCatRoleId(rolesData[0].id);
        }

        if (panelsData.length > 0) {
          setSelectedPanel(panelsData[0]);
        }
      } catch (err) {
        console.error(err);
      }
    }
    loadData();
  }, [guildId]);

  useEffect(() => {
    if (selectedPanel) {
      ShrimpyAPI.listCategories(guildId, selectedPanel.id).then(setCategories);
    } else {
      const timer = setTimeout(() => {
        setCategories([]);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [selectedPanel, guildId]);

  const handleCreatePanel = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const p = await ShrimpyAPI.createPanel(guildId, {
        channelId: newChannelId,
        title: newTitle,
        description: newDesc,
        buttonLabel: newBtnLabel,
        buttonStyle: newBtnStyle
      });
      setPanels(prev => [...prev, p]);
      setSelectedPanel(p);
    } catch (err) {
      console.error(err);
    }
  };

  const handleDeletePanel = async (panelId: string) => {
    try {
      await ShrimpyAPI.deletePanel(guildId, panelId);
      setPanels(prev => prev.filter(p => p.id !== panelId));
      if (selectedPanel?.id === panelId) {
        setSelectedPanel(null);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPanel || !newCatName) return;
    try {
      const c = await ShrimpyAPI.createCategory(guildId, selectedPanel.id, {
        name: newCatName,
        channelId: newCatChanId,
        supportRoles: [newCatRoleId]
      });
      setCategories(prev => [...prev, c]);
      setNewCatName("");
    } catch (err) {
      console.error(err);
    }
  };

  const handleDeleteCategory = async (catId: string) => {
    if (!selectedPanel) return;
    try {
      await ShrimpyAPI.deleteCategory(guildId, selectedPanel.id, catId);
      setCategories(prev => prev.filter(c => c.id !== catId));
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Ticket Panels Builder</h2>
        <p className={styles.sectionDesc}>Create beautiful, interactive ticket creation desks that post to your channels as Discord embeds.</p>
      </div>

      <div className={styles.grid}>
        {/* Left Column: Creator / Config */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          
          {/* Active Panels List */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Active Support Desks</h3>
            {panels.length === 0 ? (
              <div style={{ color: 'var(--color-text-muted)', fontSize: '13px' }}>No active panels found. Use the creator below to build one.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {panels.map(p => (
                  <div 
                    key={p.id} 
                    className={`${styles.actionBtn}`} 
                    style={{
                      justifyContent: 'space-between',
                      borderColor: selectedPanel?.id === p.id ? 'var(--color-primary)' : '',
                      background: selectedPanel?.id === p.id ? 'var(--primary-muted)' : '',
                    }}
                    onClick={() => setSelectedPanel(p)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <Layers size={14} style={{ color: 'var(--color-primary)' }} />
                      <span style={{ fontWeight: 'bold' }}>{p.title}</span>
                      <span style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>in #{channels.find(c => c.id === p.channelId)?.name || p.channelId}</span>
                    </div>
                    <button 
                      onClick={(e) => { e.stopPropagation(); handleDeletePanel(p.id); }}
                      style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Panel Creator Form */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Create New Ticket Panel</h3>
            <form onSubmit={handleCreatePanel} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.formGroup}>
                <label className={styles.label}>Panel Destination Channel</label>
                <select className={styles.select} value={newChannelId} onChange={e => setNewChannelId(e.target.value)}>
                  {channels.map(c => (
                    <option key={c.id} value={c.id}>#{c.name}</option>
                  ))}
                </select>
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Embed Title</label>
                <input className={styles.input} type="text" value={newTitle} onChange={e => setNewTitle(e.target.value)} required />
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Embed Description</label>
                <textarea className={styles.textarea} rows={3} value={newDesc} onChange={e => setNewDesc(e.target.value)} required />
              </div>

              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Button Text Label</label>
                  <input className={styles.input} type="text" value={newBtnLabel} onChange={e => setNewBtnLabel(e.target.value)} required />
                </div>
                
                <div className={styles.formGroup}>
                  <label className={styles.label}>Button Accent Style</label>
                  <select className={styles.select} value={newBtnStyle} onChange={e => setNewBtnStyle(e.target.value as 'primary' | 'success' | 'danger' | 'secondary')}>
                    <option value="primary">Primary (Blue)</option>
                    <option value="success">Success (Green)</option>
                    <option value="danger">Danger (Red)</option>
                    <option value="secondary">Secondary (Gray)</option>
                  </select>
                </div>
              </div>

              <button type="submit" className={styles.submitBtn}>
                <Plus size={16} />
                <span>Deploy Panel Desk</span>
              </button>
            </form>
          </div>

        </div>

        {/* Right Column: Live Preview & Ticket Categories */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
          
          {/* Real-time Discord Preview Card */}
          <div className={styles.card}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Eye size={16} style={{ color: 'var(--color-accent)' }} />
              <h3 className={styles.cardTitle}>Real-time Discord Preview</h3>
            </div>
            
            <div className="previewPanel" style={{ background: '#36393f', border: '1px solid #202225', padding: '16px', borderRadius: '8px', minHeight: 'auto' }}>
              <div style={{ background: '#2f3136', borderLeft: '4px solid #5865F2', borderRadius: '4px', padding: '16px', width: '100%' }}>
                <div style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '15px', marginBottom: '8px' }}>
                  {selectedPanel ? selectedPanel.title : newTitle}
                </div>
                <div style={{ color: '#dcddde', fontSize: '13px', whiteSpace: 'pre-wrap', lineHeight: '1.4' }}>
                  {selectedPanel ? selectedPanel.description : newDesc}
                </div>
                <div style={{ color: '#72767d', fontSize: '11px', marginTop: '12px' }}>
                  Response time: &lt; 15 mins
                </div>
                
                <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
                  <button 
                    style={{
                      backgroundColor: (selectedPanel ? selectedPanel.buttonStyle : newBtnStyle) === 'primary' ? '#5865F2' :
                                      (selectedPanel ? selectedPanel.buttonStyle : newBtnStyle) === 'success' ? '#3ba55d' :
                                      (selectedPanel ? selectedPanel.buttonStyle : newBtnStyle) === 'danger' ? '#d83c3e' : '#4f545c',
                      color: 'white', border: 'none', padding: '8px 16px', borderRadius: '3px', fontWeight: 500, fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px'
                    }}
                    disabled
                  >
                    <Ticket size={14} />
                    <span>{selectedPanel ? selectedPanel.buttonLabel : newBtnLabel}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Categories inside Selected Panel */}
          {selectedPanel && (
            <div className={styles.card}>
              <h3 className={styles.cardTitle}>Support Categories</h3>
              <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
                Route tickets to different channels and helper roles based on user issue selections.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
                {categories.map(c => (
                  <div key={c.id} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                    <div>
                      <span style={{ fontWeight: 'bold' }}>{c.name}</span>
                      <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginLeft: '6px' }}>
                        routes to #{channels.find(ch => ch.id === c.channelId)?.name || c.channelId}
                      </span>
                    </div>
                    <button 
                      onClick={() => handleDeleteCategory(c.id)}
                      style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))}
              </div>

              <form onSubmit={handleCreateCategory} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-4)', marginTop: 'var(--space-2)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Category Name</label>
                  <input 
                    className={styles.input} 
                    type="text" 
                    placeholder="e.g. Billing Assistance" 
                    value={newCatName} 
                    onChange={e => setNewCatName(e.target.value)} 
                    required 
                  />
                </div>

                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Spawn Thread Channel</label>
                    <select className={styles.select} value={newCatChanId} onChange={e => setNewCatChanId(e.target.value)}>
                      {channels.map(c => (
                        <option key={c.id} value={c.id}>#{c.name}</option>
                      ))}
                    </select>
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Assigned Staff Role</label>
                    <select className={styles.select} value={newCatRoleId} onChange={e => setNewCatRoleId(e.target.value)}>
                      {roles.map(r => (
                        <option key={r.id} value={r.id}>{r.name}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <button type="submit" className={styles.submitBtn} style={{ padding: '10px' }}>
                  <Plus size={14} />
                  <span>Add Category Routing</span>
                </button>
              </form>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// dashboard/app/dashboard/[guildId]/panels/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Layers,
  Plus,
  Trash2,
  Eye,
  Ticket,
  Users
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, TicketPanel, TicketCategory, PanelHandlerRole, CategoryHandlerRole, DiscordChannel, DiscordRole } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import { useToast } from "@/hooks/useToast";

const BUTTON_COLORS: Record<string, string> = {
  primary: '#5865F2',
  success: '#3ba55d',
  danger: '#d83c3e',
  secondary: '#4f545c',
};

function colorToHex(n?: number): string {
  if (n === undefined || n === null) return '#5865F2';
  return '#' + Math.max(0, Math.min(0xffffff, n)).toString(16).padStart(6, '0');
}

function hexToColor(hex: string): number {
  return parseInt(hex.replace('#', ''), 16) || 0;
}

export default function PanelsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast } = useToast();

  const [panels, setPanels] = useState<TicketPanel[]>([]);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [selectedPanel, setSelectedPanel] = useState<TicketPanel | null>(null);
  const [categories, setCategories] = useState<TicketCategory[]>([]);
  const [handlerRoles, setHandlerRoles] = useState<PanelHandlerRole[]>([]);
  const [selectedHandlerRole, setSelectedHandlerRole] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<TicketCategory | null>(null);
  const [categoryHandlerRoles, setCategoryHandlerRoles] = useState<CategoryHandlerRole[]>([]);
  const [selectedCategoryHandlerRole, setSelectedCategoryHandlerRole] = useState("");

  // Form state for new panel
  const [newName, setNewName] = useState("Main Support Desk");
  const [newChannelId, setNewChannelId] = useState("");
  const [newPanelStyle, setNewPanelStyle] = useState<'buttons' | 'select_menu'>('buttons');
  const [newContent, setNewContent] = useState("");
  const [newEmbedTitle, setNewEmbedTitle] = useState("Contact Support Services");
  const [newEmbedDesc, setNewEmbedDesc] = useState("Click a button below to open a private ticket.");
  const [newEmbedColor, setNewEmbedColor] = useState<string>('#5865F2');
  const [newAuthorName, setNewAuthorName] = useState("");
  const [newAuthorIconUrl, setNewAuthorIconUrl] = useState("");
  const [newThumbnailUrl, setNewThumbnailUrl] = useState("");
  const [newImageUrl, setNewImageUrl] = useState("");
  const [newFooterText, setNewFooterText] = useState("");
  const [newFooterIconUrl, setNewFooterIconUrl] = useState("");

  // Form state for new category
  const [newCatName, setNewCatName] = useState("");
  const [newCatButtonLabel, setNewCatButtonLabel] = useState("");
  const [newCatButtonStyle, setNewCatButtonStyle] = useState<'primary' | 'secondary' | 'success' | 'danger'>('primary');
  const [newCatEmoji, setNewCatEmoji] = useState("");
  const [newCatDestination, setNewCatDestination] = useState<'thread' | 'channel'>('thread');
  const [newCatOpenContent, setNewCatOpenContent] = useState("");
  const [newCatOpenTitle, setNewCatOpenTitle] = useState("");
  const [newCatOpenDesc, setNewCatOpenDesc] = useState("");
  const [newCatOpenColor, setNewCatOpenColor] = useState<string>('#5865F2');

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
        }
        if (rolesData.length > 0) {
          setSelectedHandlerRole(rolesData[0].id);
          setSelectedCategoryHandlerRole(rolesData[0].id);
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
      ShrimpyAPI.listPanelHandlerRoles(guildId, selectedPanel.id).then(setHandlerRoles);
    } else {
      const timer = setTimeout(() => {
        setCategories([]);
        setHandlerRoles([]);
      }, 0);
      return () => clearTimeout(timer);
    }
    setSelectedCategory(null);
  }, [selectedPanel, guildId]);

  useEffect(() => {
    if (selectedPanel && selectedCategory) {
      ShrimpyAPI.listCategoryHandlerRoles(guildId, selectedPanel.id, selectedCategory.id).then(setCategoryHandlerRoles);
    } else {
      const timer = setTimeout(() => {
        setCategoryHandlerRoles([]);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [selectedPanel, selectedCategory, guildId]);

  const handleCreatePanel = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const hasMedia = !!(newAuthorName || newThumbnailUrl || newImageUrl || newFooterText);
      const p = await ShrimpyAPI.createPanel(guildId, {
        channelId: newChannelId,
        name: newName,
        panelStyle: newPanelStyle,
        content: newContent || undefined,
        embedTitle: newEmbedTitle || undefined,
        embedDescription: newEmbedDesc || undefined,
        embedColor: (newEmbedTitle || newEmbedDesc) ? hexToColor(newEmbedColor) : undefined,
        embedMedia: hasMedia ? {
          author: newAuthorName ? { name: newAuthorName, iconUrl: newAuthorIconUrl || undefined } : undefined,
          thumbnail: newThumbnailUrl ? { url: newThumbnailUrl } : undefined,
          image: newImageUrl ? { url: newImageUrl } : undefined,
          footer: newFooterText ? { text: newFooterText, iconUrl: newFooterIconUrl || undefined } : undefined,
        } : undefined,
      });
      setPanels(prev => [...prev, p]);
      setSelectedPanel(p);
      showToast("Ticket panel deployed!", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to deploy panel.", "error");
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
    if (!selectedPanel || !newCatName || !newCatButtonLabel) return;
    if (selectedPanel.panelStyle === 'buttons' && categories.length >= 3) {
      showToast("Button layout supports up to 3 categories. Switch to Select Menu for more.", "warning");
      return;
    }
    try {
      const c = await ShrimpyAPI.createCategory(guildId, selectedPanel.id, {
        name: newCatName,
        buttonLabel: newCatButtonLabel,
        buttonStyle: newCatButtonStyle,
        emoji: newCatEmoji || undefined,
        buttonOrder: categories.length,
        ticketDestination: newCatDestination,
        ticketNameTemplate: '{category}-{number}',
        ticketOpenContent: newCatOpenContent || undefined,
        ticketOpenTitle: newCatOpenTitle || undefined,
        ticketOpenMessage: newCatOpenDesc || undefined,
        ticketOpenColor: (newCatOpenTitle || newCatOpenDesc) ? hexToColor(newCatOpenColor) : undefined,
        maxTicketsPerUser: 1,
        allowUserClose: true,
      });
      setCategories(prev => [...prev, c]);
      setNewCatName("");
      setNewCatButtonLabel("");
      setNewCatEmoji("");
      setNewCatOpenContent("");
      setNewCatOpenTitle("");
      setNewCatOpenDesc("");
    } catch (err) {
      console.error(err);
      showToast("Failed to add category.", "error");
    }
  };

  const handleDeleteCategory = async (catId: string) => {
    if (!selectedPanel) return;
    try {
      await ShrimpyAPI.deleteCategory(guildId, selectedPanel.id, catId);
      setCategories(prev => prev.filter(c => c.id !== catId));
      if (selectedCategory?.id === catId) {
        setSelectedCategory(null);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleAddCategoryHandlerRole = async () => {
    if (!selectedPanel || !selectedCategory || !selectedCategoryHandlerRole) return;
    if (categoryHandlerRoles.some(r => r.roleId === selectedCategoryHandlerRole)) {
      showToast("Role is already a ticket handler for this category!", "warning");
      return;
    }
    try {
      await ShrimpyAPI.addCategoryHandlerRole(guildId, selectedPanel.id, selectedCategory.id, selectedCategoryHandlerRole);
      const refreshed = await ShrimpyAPI.listCategoryHandlerRoles(guildId, selectedPanel.id, selectedCategory.id);
      setCategoryHandlerRoles(refreshed);
    } catch (err) {
      console.error(err);
      showToast("Failed to add category handler role.", "error");
    }
  };

  const handleRemoveCategoryHandlerRole = async (roleId: string) => {
    if (!selectedPanel || !selectedCategory) return;
    try {
      await ShrimpyAPI.removeCategoryHandlerRole(guildId, selectedPanel.id, selectedCategory.id, roleId);
      setCategoryHandlerRoles(prev => prev.filter(r => r.roleId !== roleId));
    } catch (err) {
      console.error(err);
    }
  };

  const handleAddHandlerRole = async () => {
    if (!selectedPanel || !selectedHandlerRole) return;
    if (handlerRoles.some(r => r.roleId === selectedHandlerRole)) {
      showToast("Role is already a ticket handler for this panel!", "warning");
      return;
    }
    try {
      await ShrimpyAPI.addPanelHandlerRole(guildId, selectedPanel.id, selectedHandlerRole);
      const refreshed = await ShrimpyAPI.listPanelHandlerRoles(guildId, selectedPanel.id);
      setHandlerRoles(refreshed);
    } catch (err) {
      console.error(err);
      showToast("Failed to add panel handler role.", "error");
    }
  };

  const handleRemoveHandlerRole = async (roleId: string) => {
    if (!selectedPanel) return;
    try {
      await ShrimpyAPI.removePanelHandlerRole(guildId, selectedPanel.id, roleId);
      setHandlerRoles(prev => prev.filter(r => r.roleId !== roleId));
    } catch (err) {
      console.error(err);
    }
  };

  const previewContent = selectedPanel ? selectedPanel.content : newContent;
  const previewEmbedTitle = selectedPanel ? selectedPanel.embedTitle : newEmbedTitle;
  const previewEmbedDesc = selectedPanel ? selectedPanel.embedDescription : newEmbedDesc;
  const previewEmbedColor = colorToHex(selectedPanel ? selectedPanel.embedColor : hexToColor(newEmbedColor));
  const previewMedia = selectedPanel ? selectedPanel.embedMedia : (newAuthorName || newThumbnailUrl || newImageUrl || newFooterText) ? {
    author: newAuthorName ? { name: newAuthorName, iconUrl: newAuthorIconUrl || undefined } : undefined,
    thumbnail: newThumbnailUrl ? { url: newThumbnailUrl } : undefined,
    image: newImageUrl ? { url: newImageUrl } : undefined,
    footer: newFooterText ? { text: newFooterText, iconUrl: newFooterIconUrl || undefined } : undefined,
  } : undefined;
  const hasPreviewEmbed = !!(previewEmbedTitle || previewEmbedDesc || previewMedia);
  const previewCategories = selectedPanel ? categories : [];

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Ticket Panels Builder</h2>
        <p className={styles.sectionDesc}>Create interactive ticket creation desks that post plain text and/or an embed to your channels, with one button per category.</p>
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
                      <span style={{ fontWeight: 'bold' }}>{p.name}</span>
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
              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Panel Name (internal)</label>
                  <input className={styles.input} type="text" value={newName} onChange={e => setNewName(e.target.value)} required />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Destination Channel</label>
                  <Dropdown
                    value={newChannelId}
                    onChange={setNewChannelId}
                    options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                  />
                </div>
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Plain Text Message (optional)</label>
                <textarea className={styles.textarea} rows={2} value={newContent} onChange={e => setNewContent(e.target.value)} placeholder="Sent as the message's own text, above any embed." />
              </div>

              <div style={{ borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-3)', fontSize: '12px', color: 'var(--color-text-muted)', fontWeight: 'bold' }}>
                Embed (optional — leave title &amp; description empty to send plain text only)
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Embed Title</label>
                <input className={styles.input} type="text" value={newEmbedTitle} onChange={e => setNewEmbedTitle(e.target.value)} />
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Embed Description</label>
                <textarea className={styles.textarea} rows={3} value={newEmbedDesc} onChange={e => setNewEmbedDesc(e.target.value)} />
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Embed Color</label>
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                  <input type="color" value={newEmbedColor} onChange={e => setNewEmbedColor(e.target.value)} style={{ width: '40px', height: '36px', padding: '2px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', background: 'none', cursor: 'pointer' }} />
                  <input className={styles.input} type="text" value={newEmbedColor} onChange={e => setNewEmbedColor(e.target.value)} style={{ flex: 1 }} />
                </div>
              </div>

              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Author Name</label>
                  <input className={styles.input} type="text" value={newAuthorName} onChange={e => setNewAuthorName(e.target.value)} />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Author Icon URL</label>
                  <input className={styles.input} type="text" value={newAuthorIconUrl} onChange={e => setNewAuthorIconUrl(e.target.value)} />
                </div>
              </div>

              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Thumbnail URL</label>
                  <input className={styles.input} type="text" value={newThumbnailUrl} onChange={e => setNewThumbnailUrl(e.target.value)} />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Main Image URL</label>
                  <input className={styles.input} type="text" value={newImageUrl} onChange={e => setNewImageUrl(e.target.value)} />
                </div>
              </div>

              <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Footer Text</label>
                  <input className={styles.input} type="text" value={newFooterText} onChange={e => setNewFooterText(e.target.value)} />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Footer Icon URL</label>
                  <input className={styles.input} type="text" value={newFooterIconUrl} onChange={e => setNewFooterIconUrl(e.target.value)} />
                </div>
              </div>

              <div className={styles.formGroup}>
                <label className={styles.label}>Button Layout</label>
                <Dropdown
                  value={newPanelStyle}
                  onChange={val => setNewPanelStyle(val as 'buttons' | 'select_menu')}
                  options={[
                    { value: "buttons", label: "Buttons (up to 3 categories)" },
                    { value: "select_menu", label: "Select Menu (up to 25 categories)" },
                  ]}
                />
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

            <div style={{ background: '#36393f', border: '1px solid #202225', padding: '16px', borderRadius: '8px', minHeight: 'auto' }}>
              {previewContent && (
                <div style={{ color: '#dcddde', fontSize: '14px', whiteSpace: 'pre-wrap', lineHeight: '1.4', marginBottom: hasPreviewEmbed ? '10px' : 0 }}>
                  {previewContent}
                </div>
              )}

              {hasPreviewEmbed && (
                <div style={{ background: '#2f3136', borderLeft: `4px solid ${previewEmbedColor}`, borderRadius: '4px', padding: '16px', width: '100%', display: 'flex', gap: '12px' }}>
                  <div style={{ flex: 1 }}>
                    {previewMedia?.author?.name && (
                      <div style={{ color: '#ffffff', fontSize: '12px', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                        {previewMedia.author.iconUrl && (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img src={previewMedia.author.iconUrl} alt="" style={{ width: '20px', height: '20px', borderRadius: '50%' }} />
                        )}
                        <span>{previewMedia.author.name}</span>
                      </div>
                    )}
                    {previewEmbedTitle && (
                      <div style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '15px', marginBottom: '8px' }}>
                        {previewEmbedTitle}
                      </div>
                    )}
                    {previewEmbedDesc && (
                      <div style={{ color: '#dcddde', fontSize: '13px', whiteSpace: 'pre-wrap', lineHeight: '1.4' }}>
                        {previewEmbedDesc}
                      </div>
                    )}
                    {previewMedia?.image?.url && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={previewMedia.image.url} alt="" style={{ maxWidth: '100%', borderRadius: '4px', marginTop: '10px' }} />
                    )}
                    {previewMedia?.footer?.text && (
                      <div style={{ color: '#72767d', fontSize: '11px', marginTop: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                        {previewMedia.footer.iconUrl && (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img src={previewMedia.footer.iconUrl} alt="" style={{ width: '16px', height: '16px', borderRadius: '50%' }} />
                        )}
                        <span>{previewMedia.footer.text}</span>
                      </div>
                    )}
                  </div>
                  {previewMedia?.thumbnail?.url && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={previewMedia.thumbnail.url} alt="" style={{ width: '64px', height: '64px', borderRadius: '4px', objectFit: 'cover', flexShrink: 0 }} />
                  )}
                </div>
              )}

              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginTop: '14px' }}>
                {previewCategories.length > 0 ? (
                  previewCategories.map(c => (
                    <button
                      key={c.id}
                      style={{
                        backgroundColor: BUTTON_COLORS[c.buttonStyle] || BUTTON_COLORS.primary,
                        color: 'white', border: 'none', padding: '8px 16px', borderRadius: '3px', fontWeight: 500, fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px'
                      }}
                      disabled
                    >
                      <Ticket size={14} />
                      <span>{c.emoji ? `${c.emoji} ` : ''}{c.buttonLabel}</span>
                    </button>
                  ))
                ) : (
                  <button
                    style={{
                      backgroundColor: BUTTON_COLORS[newCatButtonStyle] || BUTTON_COLORS.primary,
                      color: 'white', border: 'none', padding: '8px 16px', borderRadius: '3px', fontWeight: 500, fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px', opacity: 0.6
                    }}
                    disabled
                  >
                    <Ticket size={14} />
                    <span>Add a category to see buttons</span>
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* Categories inside Selected Panel */}
          {selectedPanel && (
            <div className={styles.card}>
              <h3 className={styles.cardTitle}>Support Categories</h3>
              <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
                Each category becomes one button on the panel. Click a category to manage its ticket handler roles.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
                {categories.map(c => (
                  <div
                    key={c.id}
                    className={styles.actionBtn}
                    style={{
                      justifyContent: 'space-between',
                      borderColor: selectedCategory?.id === c.id ? 'var(--color-primary)' : '',
                      background: selectedCategory?.id === c.id ? 'var(--primary-muted)' : '',
                    }}
                    onClick={() => setSelectedCategory(c)}
                  >
                    <div>
                      <span style={{ fontWeight: 'bold' }}>{c.emoji ? `${c.emoji} ` : ''}{c.name}</span>
                      <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginLeft: '6px' }}>
                        opens a {c.ticketDestination}
                      </span>
                    </div>
                    <button
                      onClick={(e) => { e.stopPropagation(); handleDeleteCategory(c.id); }}
                      style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))}
              </div>

              {selectedPanel.panelStyle === 'buttons' && categories.length >= 3 ? (
                <div style={{ borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-4)', marginTop: 'var(--space-2)', fontSize: '12px', color: 'var(--color-text-muted)' }}>
                  This panel uses Button layout, which supports up to 3 categories. Delete a category to add another, or create a new panel with Select Menu layout to support more.
                </div>
              ) : (
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
                    <label className={styles.label}>Button Label</label>
                    <input
                      className={styles.input}
                      type="text"
                      placeholder="e.g. Billing Help"
                      value={newCatButtonLabel}
                      onChange={e => setNewCatButtonLabel(e.target.value)}
                      required
                    />
                  </div>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Button Emoji (optional)</label>
                    <input
                      className={styles.input}
                      type="text"
                      placeholder="🎫"
                      value={newCatEmoji}
                      onChange={e => setNewCatEmoji(e.target.value)}
                    />
                  </div>
                </div>

                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Button Style</label>
                    <Dropdown
                      value={newCatButtonStyle}
                      onChange={val => setNewCatButtonStyle(val as 'primary' | 'secondary' | 'success' | 'danger')}
                      options={[
                        { value: "primary", label: "Primary (Blue)" },
                        { value: "success", label: "Success (Green)" },
                        { value: "danger", label: "Danger (Red)" },
                        { value: "secondary", label: "Secondary (Gray)" },
                      ]}
                    />
                  </div>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Opens As</label>
                    <Dropdown
                      value={newCatDestination}
                      onChange={val => setNewCatDestination(val as 'thread' | 'channel')}
                      options={[
                        { value: "thread", label: "Private Thread" },
                        { value: "channel", label: "Dedicated Channel" },
                      ]}
                    />
                  </div>
                </div>

                <div style={{ borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-3)', fontSize: '12px', color: 'var(--color-text-muted)', fontWeight: 'bold' }}>
                  Ticket greeting (sent inside the opened ticket)
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Plain Text Greeting (optional)</label>
                  <textarea className={styles.textarea} rows={2} value={newCatOpenContent} onChange={e => setNewCatOpenContent(e.target.value)} placeholder="Sent below the automatic &quot;Welcome @user&quot; line." />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Greeting Embed Title (optional)</label>
                  <input className={styles.input} type="text" value={newCatOpenTitle} onChange={e => setNewCatOpenTitle(e.target.value)} />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Greeting Embed Description (optional)</label>
                  <textarea className={styles.textarea} rows={2} value={newCatOpenDesc} onChange={e => setNewCatOpenDesc(e.target.value)} />
                </div>

                {(newCatOpenTitle || newCatOpenDesc) && (
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Greeting Embed Color</label>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <input type="color" value={newCatOpenColor} onChange={e => setNewCatOpenColor(e.target.value)} style={{ width: '40px', height: '36px', padding: '2px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', background: 'none', cursor: 'pointer' }} />
                      <input className={styles.input} type="text" value={newCatOpenColor} onChange={e => setNewCatOpenColor(e.target.value)} style={{ flex: 1 }} />
                    </div>
                  </div>
                )}

                <button type="submit" className={styles.submitBtn} style={{ padding: '10px' }}>
                  <Plus size={14} />
                  <span>Add Category</span>
                </button>
              </form>
              )}
            </div>
          )}

          {/* Ticket Handler Roles for Selected Category */}
          {selectedPanel && selectedCategory && (
            <div className={styles.card}>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Users size={18} style={{ color: 'var(--color-primary)' }} />
                <h3 className={styles.cardTitle}>&quot;{selectedCategory.name}&quot; Handler Roles</h3>
              </div>
              <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
                Roles invited into tickets opened from this category specifically, in addition to the panel&apos;s handler roles above. This does not grant dashboard access.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
                {categoryHandlerRoles.length === 0 ? (
                  <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No category-specific handler roles. Only the panel&apos;s handler roles will be invited.</div>
                ) : (
                  categoryHandlerRoles.map(hr => {
                    const matched = roles.find(r => r.id === hr.roleId);
                    return (
                      <div key={hr.id} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                        <span style={{ fontWeight: 'bold' }}>{matched?.name || hr.roleId}</span>
                        <button
                          type="button"
                          onClick={() => handleRemoveCategoryHandlerRole(hr.roleId)}
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
                  value={selectedCategoryHandlerRole}
                  onChange={setSelectedCategoryHandlerRole}
                  options={roles.map(r => ({ value: r.id, label: r.name }))}
                  style={{ flex: 1 }}
                />
                <button onClick={handleAddCategoryHandlerRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                  <Plus size={14} />
                  <span>Add</span>
                </button>
              </div>
            </div>
          )}

          {/* Ticket Handler Roles for Selected Panel */}
          {selectedPanel && (
            <div className={styles.card}>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                <Users size={18} style={{ color: 'var(--color-primary)' }} />
                <h3 className={styles.cardTitle}>Ticket Handler Roles</h3>
              </div>
              <p className={styles.sectionDesc} style={{ fontSize: '12px' }}>
                Roles invited into the Discord channel or thread created for a ticket opened from this panel, so they can handle it. This does not grant dashboard access.
              </p>

              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
                {handlerRoles.length === 0 ? (
                  <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No handler roles added. Only the ticket opener and the bot will see the channel.</div>
                ) : (
                  handlerRoles.map(hr => {
                    const matched = roles.find(r => r.id === hr.roleId);
                    return (
                      <div key={hr.id} className={styles.actionBtn} style={{ justifyContent: 'space-between', cursor: 'default' }}>
                        <span style={{ fontWeight: 'bold' }}>{matched?.name || hr.roleId}</span>
                        <button
                          type="button"
                          onClick={() => handleRemoveHandlerRole(hr.roleId)}
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
                  value={selectedHandlerRole}
                  onChange={setSelectedHandlerRole}
                  options={roles.map(r => ({ value: r.id, label: r.name }))}
                  style={{ flex: 1 }}
                />
                <button onClick={handleAddHandlerRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                  <Plus size={14} />
                  <span>Add</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

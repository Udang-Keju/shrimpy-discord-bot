// dashboard/lib/api.ts

export interface Guild {
  id: string;
  name: string;
  icon?: string;
  nickname?: string;
  prefix?: string;
  ticketLimit?: number;
  logChannelId?: string;
  autoRoles: string[];
  staffRoles: string[];
  bot_joined?: boolean;
}

export interface PublicConfig {
  client_id: string;
  redirect_uri: string;
}

export interface Ticket {
  id: string;
  guildId: string;
  channelId: string;
  creatorId: string;
  creatorUsername: string;
  status: 'open' | 'claimed' | 'resolved' | 'closed' | 'archived';
  assignedTo?: string;
  categoryName: string;
  createdAt: string;
  closedAt?: string;
}

export interface EmbedMediaAuthor {
  name: string;
  iconUrl?: string;
  url?: string;
}

export interface EmbedMediaFooter {
  text: string;
  iconUrl?: string;
}

export interface EmbedMedia {
  author?: EmbedMediaAuthor;
  thumbnail?: { url: string };
  image?: { url: string };
  footer?: EmbedMediaFooter;
}

export interface TicketPanel {
  id: string;
  guildId: string;
  channelId: string;
  messageId?: string;
  name: string;
  panelStyle: 'buttons' | 'select_menu';
  content?: string;
  embedTitle?: string;
  embedDescription?: string;
  embedColor?: number;
  embedMedia?: EmbedMedia;
  // Write-only: included in create/update payloads so the backend can reconcile
  // panel_handler_roles to match exactly. Omitted (not just empty) on reads.
  handlerRoleIds?: string[];
}

export interface TicketCategory {
  id: string;
  panelId: string;
  name: string;
  emoji?: string;
  buttonLabel: string;
  buttonStyle: 'primary' | 'secondary' | 'success' | 'danger';
  buttonDescription?: string;
  buttonOrder: number;
  ticketDestination: 'thread' | 'channel';
  // Thread destination: the parent text channel a private thread is started from.
  // Empty/omitted ⇒ the panel's channel is used.
  threadParentChannelId?: string;
  // Channel destination: the Discord channel group the dedicated channel is placed under.
  // Empty/omitted ⇒ no group (guild root).
  channelCategoryId?: string;
  ticketNameTemplate: string;
  ticketOpenTitle?: string;
  ticketOpenMessage?: string;
  ticketOpenColor?: number;
  ticketOpenMedia?: EmbedMedia;
  ticketOpenContent?: string;
  maxTicketsPerUser: number;
  autoCloseHours?: number;
  transcriptChannelId?: string;
  allowUserClose: boolean;
  // Write-only: included in create/update payloads so the backend can reconcile
  // category_handler_roles to match exactly. Omitted (not just empty) on reads.
  handlerRoleIds?: string[];
}

// Roles invited into the Discord channel/thread created for a ticket, so they
// can handle it. Distinct from Guild.staffRoles, which gate dashboard access.
export interface PanelHandlerRole {
  id: string;
  panelId: string;
  roleId: string;
}

// Same concept as PanelHandlerRole, but scoped to one category — additive to
// the panel's handler roles when a ticket is opened from that category.
export interface CategoryHandlerRole {
  id: string;
  categoryId: string;
  roleId: string;
}

export interface WelcomeConfig {
  guildId: string;
  sendDm: boolean;
  dmText: string;
  sendChannel: boolean;
  channelId: string;
  welcomeText: string;
  cardStyle: 'dark' | 'light';
  avatarEmoji: string;
}

export interface TranslationChannelConfig {
  channelId: string;
  targetLangOverride: string | null;
}

export interface TranslationEmojiConfig {
  emoji: string;
  targetLangOverride: string | null;
}

export interface TranslationConfig {
  guildId: string;
  enabled: boolean;
  autoEnabled: boolean;
  reactionEnabled: boolean;
  reactionDelivery: 'channel' | 'dm'; // where reaction-triggered translations are sent
  provider: string;            // 'deepl' | 'google' | 'libretranslate'
  apiKey: string;              // masked '***' when a key is stored, '' otherwise
  hasApiKey: boolean;
  endpointUrl: string | null;  // for self-hosted engines (LibreTranslate)
  targetLang: string | null;   // null = fall back to the server language
  channels: TranslationChannelConfig[];
  emojis: TranslationEmojiConfig[];
}

export interface ReactionRoleMapping {
  emoji: string;
  roleId: string;
  roleName: string;
}

export interface ReactionRole {
  id: string; // Message ID
  guildId: string;
  channelId: string;
  title: string;
  description: string;
  mappings: ReactionRoleMapping[];
}

export interface DiscordChannel {
  id: string;
  name: string;
  type: string;
}

export interface DiscordRole {
  id: string;
  name: string;
  color: string;
}

// A guild's custom emoji, as surfaced to the emoji picker. `mention` is the
// canonical Discord form (<:name:id> / <a:name:id>) stored/sent for this emoji;
// `url` is the CDN image for preview.
export interface DiscordEmoji {
  id: string;
  name: string;
  animated: boolean;
  mention: string;
  url: string;
}

export interface DiscordUser {
  id: string;
  username: string;
  globalName?: string;
  avatar?: string;
}

const API_BASE = process.env.NEXT_PUBLIC_SHRIMPY_API_URL || 'http://localhost:8080';

// In-memory cache for guild custom emojis. The picker reopens frequently but a
// server's emoji set rarely changes, so a short TTL avoids refetching on every open
// while still picking up edits within a few minutes.
const EMOJI_CACHE_TTL_MS = 5 * 60 * 1000;
const emojiCache = new Map<string, { emojis: DiscordEmoji[]; at: number }>();

// Simulated Mock Database for Standalone/Offline mode
const mockGuilds: Guild[] = [
  {
    id: '123456789012345678',
    name: 'Shrimpy Sandbox',
    icon: '🦐',
    nickname: 'Shrimpy Helper',
    prefix: '!',
    ticketLimit: 3,
    logChannelId: 'logs',
    autoRoles: ['member-role-id'],
    staffRoles: ['mod-role-id'],
  },
  {
    id: '876543210987654321',
    name: 'Gamer Guild',
    icon: '🎮',
    nickname: 'Game Shrimpy',
    prefix: '?',
    ticketLimit: 5,
    logChannelId: 'bot-logs',
    autoRoles: [],
    staffRoles: [],
  }
];

const mockTickets: Ticket[] = [
  {
    id: 'ticket-1001',
    guildId: '123456789012345678',
    channelId: 'ticket-0001',
    creatorId: 'user-1',
    creatorUsername: 'ShrimpLover42',
    status: 'open',
    categoryName: 'Billing Support',
    createdAt: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: 'ticket-1002',
    guildId: '123456789012345678',
    channelId: 'ticket-0002',
    creatorId: 'user-2',
    creatorUsername: 'OceanMan',
    status: 'claimed',
    assignedTo: 'ModStaff',
    categoryName: 'Technical Bug',
    createdAt: new Date(Date.now() - 7200000).toISOString(),
  },
  {
    id: 'ticket-1003',
    guildId: '123456789012345678',
    channelId: 'ticket-0003',
    creatorId: 'user-3',
    creatorUsername: 'CoralReef',
    status: 'closed',
    assignedTo: 'AdminBob',
    categoryName: 'General Help',
    createdAt: new Date(Date.now() - 86400000).toISOString(),
    closedAt: new Date(Date.now() - 82000000).toISOString(),
  }
];

const mockWelcomeConfigs: Record<string, WelcomeConfig> = {
  '123456789012345678': {
    guildId: '123456789012345678',
    sendDm: true,
    dmText: 'Thanks for joining Shrimpy Sandbox server! Make sure to read the rules.',
    sendChannel: true,
    channelId: 'general',
    welcomeText: 'Welcome to Shrimpy Server!',
    cardStyle: 'dark',
    avatarEmoji: '🦐'
  }
};

const mockTranslationConfigs: Record<string, TranslationConfig> = {};

function defaultTranslationConfig(guildId: string): TranslationConfig {
  return {
    guildId,
    enabled: false,
    autoEnabled: false,
    reactionEnabled: false,
    reactionDelivery: 'channel',
    provider: 'deepl',
    apiKey: '',
    hasApiKey: false,
    endpointUrl: null,
    targetLang: null,
    channels: [],
    emojis: []
  };
}

let mockReactionRoles: ReactionRole[] = [
  {
    id: 'msg-9901',
    guildId: '123456789012345678',
    channelId: 'roles-picker',
    title: 'Roles Picker Desk',
    description: 'Click on the reaction badges below to self-assign tags in this guild.',
    mappings: [
      { emoji: '🦐', roleId: 'member', roleName: 'Server Member' },
      { emoji: '🛠️', roleId: 'developer', roleName: 'Guild Developer' },
      { emoji: '🎮', roleId: 'gamer', roleName: 'Community Gamer' }
    ]
  }
];

let mockPanels: TicketPanel[] = [
  {
    id: 'panel-001',
    guildId: '123456789012345678',
    channelId: 'support-desk',
    name: 'Main Support Desk',
    panelStyle: 'buttons',
    embedTitle: 'Contact Support Services',
    embedDescription: 'Click the button below to open a private support ticket. Our staff is available 24/7.'
  }
];

const mockCategories: Record<string, TicketCategory[]> = {
  'panel-001': [
    {
      id: 'cat-001', panelId: 'panel-001', name: 'General Help', buttonLabel: 'General Help',
      buttonStyle: 'primary', buttonOrder: 0, ticketDestination: 'thread',
      ticketNameTemplate: '{category}-{number}', maxTicketsPerUser: 1, allowUserClose: true
    },
    {
      id: 'cat-002', panelId: 'panel-001', name: 'Billing', buttonLabel: 'Billing',
      buttonStyle: 'secondary', buttonOrder: 1, ticketDestination: 'thread',
      ticketNameTemplate: '{category}-{number}', maxTicketsPerUser: 1, allowUserClose: true
    }
  ]
};

const mockPanelHandlerRoles: Record<string, PanelHandlerRole[]> = {
  'panel-001': [
    { id: 'phr-001', panelId: 'panel-001', roleId: 'mod-role-id' }
  ]
};

const mockCategoryHandlerRoles: Record<string, CategoryHandlerRole[]> = {
  'cat-001': [
    { id: 'chr-001', categoryId: 'cat-001', roleId: 'mod-role-id' }
  ],
  'cat-002': [
    { id: 'chr-002', categoryId: 'cat-002', roleId: 'admin-role-id' }
  ]
};

const mockChannels: DiscordChannel[] = [
  { id: 'general', name: 'general-chat', type: 'text' },
  { id: 'rules', name: 'rules-and-info', type: 'text' },
  { id: 'announcements', name: 'announcements', type: 'text' },
  { id: 'support-desk', name: 'support-desk', type: 'text' },
  { id: 'logs', name: 'bot-logs', type: 'text' }
];

const mockChannelGroups: DiscordChannel[] = [
  { id: 'group-support', name: 'Support', type: 'category' },
  { id: 'group-community', name: 'Community', type: 'category' }
];

const mockEmojis: DiscordEmoji[] = [
  { id: '1001', name: 'shrimpy', animated: false, mention: '<:shrimpy:1001>', url: 'https://cdn.discordapp.com/emojis/1001.png' },
  { id: '1002', name: 'pog', animated: false, mention: '<:pog:1002>', url: 'https://cdn.discordapp.com/emojis/1002.png' },
  { id: '1003', name: 'party', animated: true, mention: '<a:party:1003>', url: 'https://cdn.discordapp.com/emojis/1003.gif' }
];

const mockRoles: DiscordRole[] = [
  { id: 'member', name: 'Server Member', color: '#7289da' },
  { id: 'developer', name: 'Guild Developer', color: '#4f545c' },
  { id: 'gamer', name: 'Community Gamer', color: '#43b581' },
  { id: 'mod-role-id', name: 'Moderator', color: '#faa61a' },
  { id: 'admin-role-id', name: 'Administrator', color: '#f04747' }
];

// Demo mode (USER_JOURNEY §14.1): explicit /demo route only — never a silent
// fallback for authenticated sessions. Real requests below throw on failure.
const DEMO_FLAG_KEY = 'shrimpy_demo';

export function isDemoMode(): boolean {
  if (typeof window === 'undefined') return false;
  return window.sessionStorage.getItem(DEMO_FLAG_KEY) === '1';
}

function enterDemoMode(): void {
  window.sessionStorage.setItem(DEMO_FLAG_KEY, '1');
}

function exitDemoMode(): void {
  window.sessionStorage.removeItem(DEMO_FLAG_KEY);
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${url}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers || {}),
    },
  });
  if (!res.ok) {
    throw new Error(`API Error: ${res.status} ${res.statusText}`);
  }
  return await res.json() as T;
}

export const ShrimpyAPI = {
  enterDemoMode,
  exitDemoMode,

  // Auth
  getCurrentUser: async (): Promise<DiscordUser> => {
    if (isDemoMode()) {
      return {
        id: 'discord-user-123456',
        username: 'shrimp_commander',
        globalName: 'Shrimp Commander',
        avatar: 'https://images.unsplash.com/photo-1553753861-267865544b20?w=150'
      };
    }
    return fetchJSON<DiscordUser>('/api/v1/auth/me');
  },

  getPublicConfig: async (): Promise<PublicConfig> => {
    if (isDemoMode()) {
      return {
        client_id: '123456789012345678',
        redirect_uri: 'http://localhost:8080/api/v1/auth/callback'
      };
    }
    return fetchJSON<PublicConfig>('/api/v1/config');
  },

  logout: async (): Promise<void> => {
    await fetch(`${API_BASE}/api/v1/auth/logout`, { method: 'DELETE', credentials: 'include' });
  },

  // Re-fetches the user's current Discord guilds/permissions and re-issues the session
  // cookie, so newly joined guilds or newly granted permissions show up without a re-login.
  refreshSession: async (): Promise<void> => {
    if (isDemoMode()) return;
    await fetch(`${API_BASE}/api/v1/auth/refresh`, { method: 'POST', credentials: 'include' });
  },

  // Guilds
  listGuilds: async (): Promise<Guild[]> => {
    if (isDemoMode()) return mockGuilds;
    return fetchJSON<Guild[]>('/api/v1/guilds');
  },

  getGuildConfig: async (guildId: string): Promise<Guild> => {
    if (isDemoMode()) {
      return mockGuilds.find(g => g.id === guildId) || mockGuilds[0];
    }
    return fetchJSON<Guild>(`/api/v1/guilds/${guildId}`);
  },

  updateGuildConfig: async (guildId: string, updates: Partial<Guild>): Promise<Guild> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      Object.assign(g, updates);
      return g;
    }
    return fetchJSON<Guild>(`/api/v1/guilds/${guildId}`, {
      method: 'PATCH',
      body: JSON.stringify(updates)
    });
  },

  updateNickname: async (guildId: string, nickname: string | null): Promise<void> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      g.nickname = nickname || undefined;
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/nickname`, {
      method: 'PATCH',
      body: JSON.stringify({ nickname })
    });
  },

  addAutoRole: async (guildId: string, roleId: string): Promise<void> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      g.autoRoles.push(roleId);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/auto-roles`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId })
    });
  },

  removeAutoRole: async (guildId: string, roleId: string): Promise<void> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      g.autoRoles = g.autoRoles.filter(r => r !== roleId);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/auto-roles/${roleId}`, { method: 'DELETE' });
  },

  addStaffRole: async (guildId: string, roleId: string): Promise<void> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      g.staffRoles.push(roleId);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/staff-roles`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId })
    });
  },

  removeStaffRole: async (guildId: string, roleId: string): Promise<void> => {
    if (isDemoMode()) {
      const g = mockGuilds.find(x => x.id === guildId) || mockGuilds[0];
      g.staffRoles = g.staffRoles.filter(r => r !== roleId);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/staff-roles/${roleId}`, { method: 'DELETE' });
  },

  // Discord API proxies (for dropdown selects)
  getDiscordChannels: async (guildId: string): Promise<DiscordChannel[]> => {
    if (isDemoMode()) return mockChannels;
    return fetchJSON<DiscordChannel[]>(`/api/v1/guilds/${guildId}/discord/channels`);
  },

  // Channel groups (Discord category channels) — used to place a ticket's dedicated channel.
  getDiscordChannelGroups: async (guildId: string): Promise<DiscordChannel[]> => {
    if (isDemoMode()) return mockChannelGroups;
    return fetchJSON<DiscordChannel[]>(`/api/v1/guilds/${guildId}/discord/categories`);
  },

  getDiscordRoles: async (guildId: string): Promise<DiscordRole[]> => {
    if (isDemoMode()) return mockRoles;
    return fetchJSON<DiscordRole[]>(`/api/v1/guilds/${guildId}/discord/roles`);
  },

  // The guild's custom emojis, for the emoji picker's "Server" tab. Cached per
  // guild for EMOJI_CACHE_TTL_MS since the picker is opened often but a server's
  // emoji set rarely changes; pass force=true to bypass (e.g. a manual refresh).
  getDiscordEmojis: async (guildId: string, force = false): Promise<DiscordEmoji[]> => {
    if (isDemoMode()) return mockEmojis;
    const cached = emojiCache.get(guildId);
    if (!force && cached && Date.now() - cached.at < EMOJI_CACHE_TTL_MS) {
      return cached.emojis;
    }
    const emojis = await fetchJSON<DiscordEmoji[]>(`/api/v1/guilds/${guildId}/discord/emojis`);
    emojiCache.set(guildId, { emojis, at: Date.now() });
    return emojis;
  },

  // Welcome Config
  getWelcomeConfig: async (guildId: string): Promise<WelcomeConfig> => {
    if (isDemoMode()) {
      return mockWelcomeConfigs[guildId] || {
        guildId,
        sendDm: false,
        dmText: '',
        sendChannel: false,
        channelId: '',
        welcomeText: 'Welcome!',
        cardStyle: 'dark',
        avatarEmoji: '🦐'
      };
    }
    return fetchJSON<WelcomeConfig>(`/api/v1/guilds/${guildId}/welcome`);
  },

  saveWelcomeConfig: async (guildId: string, config: WelcomeConfig): Promise<WelcomeConfig> => {
    if (isDemoMode()) {
      mockWelcomeConfigs[guildId] = config;
      return config;
    }
    return fetchJSON<WelcomeConfig>(`/api/v1/guilds/${guildId}/welcome`, {
      method: 'PUT',
      body: JSON.stringify(config)
    });
  },

  // Translation Config
  getTranslationConfig: async (guildId: string): Promise<TranslationConfig> => {
    if (isDemoMode()) {
      return mockTranslationConfigs[guildId] || defaultTranslationConfig(guildId);
    }
    return fetchJSON<TranslationConfig>(`/api/v1/guilds/${guildId}/translation`);
  },

  saveTranslationConfig: async (guildId: string, config: TranslationConfig): Promise<TranslationConfig> => {
    if (isDemoMode()) {
      const stored = { ...config, hasApiKey: config.hasApiKey || (config.apiKey !== '' && config.apiKey !== '***'), apiKey: '***' };
      mockTranslationConfigs[guildId] = stored;
      return stored;
    }
    return fetchJSON<TranslationConfig>(`/api/v1/guilds/${guildId}/translation`, {
      method: 'PUT',
      body: JSON.stringify({
        enabled: config.enabled,
        autoEnabled: config.autoEnabled,
        reactionEnabled: config.reactionEnabled,
        reactionDelivery: config.reactionDelivery,
        provider: config.provider,
        apiKey: config.apiKey,
        endpointUrl: config.endpointUrl,
        targetLang: config.targetLang
      })
    });
  },

  addTranslationChannel: async (guildId: string, channelId: string, targetLangOverride: string | null = null): Promise<void> => {
    if (isDemoMode()) {
      const cfg = mockTranslationConfigs[guildId] || defaultTranslationConfig(guildId);
      if (!cfg.channels.some(c => c.channelId === channelId)) {
        cfg.channels = [...cfg.channels, { channelId, targetLangOverride }];
      }
      mockTranslationConfigs[guildId] = cfg;
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/translation/channels`, {
      method: 'POST',
      body: JSON.stringify({ channelId, targetLangOverride })
    });
  },

  removeTranslationChannel: async (guildId: string, channelId: string): Promise<void> => {
    if (isDemoMode()) {
      const cfg = mockTranslationConfigs[guildId];
      if (cfg) cfg.channels = cfg.channels.filter(c => c.channelId !== channelId);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/translation/channels/${channelId}`, { method: 'DELETE' });
  },

  addTranslationEmoji: async (guildId: string, emoji: string, targetLangOverride: string | null = null): Promise<void> => {
    if (isDemoMode()) {
      const cfg = mockTranslationConfigs[guildId] || defaultTranslationConfig(guildId);
      if (!cfg.emojis.some(e => e.emoji === emoji)) {
        cfg.emojis = [...cfg.emojis, { emoji, targetLangOverride }];
      }
      mockTranslationConfigs[guildId] = cfg;
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/translation/emojis`, {
      method: 'POST',
      body: JSON.stringify({ emoji, targetLangOverride })
    });
  },

  removeTranslationEmoji: async (guildId: string, emoji: string): Promise<void> => {
    if (isDemoMode()) {
      const cfg = mockTranslationConfigs[guildId];
      if (cfg) cfg.emojis = cfg.emojis.filter(e => e.emoji !== emoji);
      return;
    }
    await fetchJSON(`/api/v1/guilds/${guildId}/translation/emojis?emoji=${encodeURIComponent(emoji)}`, { method: 'DELETE' });
  },

  // Ticket Panels
  listPanels: async (guildId: string): Promise<TicketPanel[]> => {
    if (isDemoMode()) return mockPanels.filter(p => p.guildId === guildId);
    return fetchJSON<TicketPanel[]>(`/api/v1/guilds/${guildId}/panels`);
  },

  createPanel: async (guildId: string, panel: Omit<TicketPanel, 'id' | 'guildId'>): Promise<TicketPanel> => {
    if (isDemoMode()) {
      const newPanel: TicketPanel = {
        ...panel,
        id: `panel-${Math.floor(Math.random() * 1000)}`,
        guildId
      };
      mockPanels.push(newPanel);
      if (panel.handlerRoleIds) {
        mockPanelHandlerRoles[newPanel.id] = panel.handlerRoleIds.map(roleId => ({
          id: `phr-${Math.floor(Math.random() * 1000)}`, panelId: newPanel.id, roleId
        }));
      }
      return newPanel;
    }
    return fetchJSON<TicketPanel>(`/api/v1/guilds/${guildId}/panels`, {
      method: 'POST',
      body: JSON.stringify(panel)
    });
  },

  updatePanel: async (guildId: string, panelId: string, panel: Omit<TicketPanel, 'id' | 'guildId'>): Promise<TicketPanel> => {
    if (isDemoMode()) {
      const updated: TicketPanel = { ...panel, id: panelId, guildId };
      mockPanels = mockPanels.map(p => p.id === panelId ? updated : p);
      if (panel.handlerRoleIds) {
        mockPanelHandlerRoles[panelId] = panel.handlerRoleIds.map(roleId => ({
          id: `phr-${Math.floor(Math.random() * 1000)}`, panelId, roleId
        }));
      }
      return updated;
    }
    return fetchJSON<TicketPanel>(`/api/v1/guilds/${guildId}/panels/${panelId}`, {
      method: 'PATCH',
      body: JSON.stringify(panel)
    });
  },

  deletePanel: async (guildId: string, panelId: string): Promise<void> => {
    if (isDemoMode()) {
      mockPanels = mockPanels.filter(p => p.id !== panelId);
      return;
    }
    await fetch(`${API_BASE}/api/v1/guilds/${guildId}/panels/${panelId}`, { method: 'DELETE', credentials: 'include' });
  },

  listCategories: async (guildId: string, panelId: string): Promise<TicketCategory[]> => {
    if (isDemoMode()) return mockCategories[panelId] || [];
    return fetchJSON<TicketCategory[]>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories`);
  },

  createCategory: async (guildId: string, panelId: string, cat: Omit<TicketCategory, 'id' | 'panelId'>): Promise<TicketCategory> => {
    if (isDemoMode()) {
      const newCat: TicketCategory = {
        ...cat,
        id: `cat-${Math.floor(Math.random() * 1000)}`,
        panelId
      };
      if (!mockCategories[panelId]) mockCategories[panelId] = [];
      mockCategories[panelId].push(newCat);
      if (cat.handlerRoleIds) {
        mockCategoryHandlerRoles[newCat.id] = cat.handlerRoleIds.map(roleId => ({
          id: `chr-${Math.floor(Math.random() * 1000)}`, categoryId: newCat.id, roleId
        }));
      }
      return newCat;
    }
    return fetchJSON<TicketCategory>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories`, {
      method: 'POST',
      body: JSON.stringify(cat)
    });
  },

  updateCategory: async (guildId: string, panelId: string, catId: string, cat: Omit<TicketCategory, 'id' | 'panelId'>): Promise<TicketCategory> => {
    if (isDemoMode()) {
      const updated: TicketCategory = { ...cat, id: catId, panelId };
      if (mockCategories[panelId]) {
        mockCategories[panelId] = mockCategories[panelId].map(c => c.id === catId ? updated : c);
      }
      if (cat.handlerRoleIds) {
        mockCategoryHandlerRoles[catId] = cat.handlerRoleIds.map(roleId => ({
          id: `chr-${Math.floor(Math.random() * 1000)}`, categoryId: catId, roleId
        }));
      }
      return updated;
    }
    return fetchJSON<TicketCategory>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories/${catId}`, {
      method: 'PATCH',
      body: JSON.stringify(cat)
    });
  },

  deleteCategory: async (guildId: string, panelId: string, catId: string): Promise<void> => {
    if (isDemoMode()) {
      if (mockCategories[panelId]) {
        mockCategories[panelId] = mockCategories[panelId].filter(c => c.id !== catId);
      }
      return;
    }
    await fetch(`${API_BASE}/api/v1/guilds/${guildId}/panels/${panelId}/categories/${catId}`, { method: 'DELETE', credentials: 'include' });
  },

  // Per-panel ticket handler roles (who gets invited into the created ticket channel/thread).
  // Writes happen via createPanel/updatePanel's handlerRoleIds field instead of dedicated
  // endpoints — the backend reconciles the set from the panel payload.
  listPanelHandlerRoles: async (guildId: string, panelId: string): Promise<PanelHandlerRole[]> => {
    if (isDemoMode()) return mockPanelHandlerRoles[panelId] || [];
    return fetchJSON<PanelHandlerRole[]>(`/api/v1/guilds/${guildId}/panels/${panelId}/handler-roles`);
  },

  // Per-category ticket handler roles (additive to the panel's handler roles). Writes
  // happen via createCategory/updateCategory's handlerRoleIds field.
  listCategoryHandlerRoles: async (guildId: string, panelId: string, catId: string): Promise<CategoryHandlerRole[]> => {
    if (isDemoMode()) return mockCategoryHandlerRoles[catId] || [];
    return fetchJSON<CategoryHandlerRole[]>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories/${catId}/handler-roles`);
  },

  // Tickets management
  listTickets: async (guildId: string): Promise<Ticket[]> => {
    if (isDemoMode()) return mockTickets.filter(t => t.guildId === guildId);
    const res = await fetchJSON<{ tickets: Ticket[] }>(`/api/v1/guilds/${guildId}/tickets`);
    return res.tickets;
  },

  updateTicket: async (guildId: string, ticketId: string, updates: Partial<Ticket>): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      Object.assign(t, updates);
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}`, {
      method: 'PATCH',
      body: JSON.stringify(updates)
    });
  },

  claimTicket: async (guildId: string, ticketId: string, username: string): Promise<Ticket> => {
    return ShrimpyAPI.updateTicket(guildId, ticketId, { status: 'claimed', assignedTo: username });
  },

  resolveTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      t.status = 'resolved';
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/resolve`, { method: 'POST' });
  },

  unresolveTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      t.status = t.assignedTo ? 'claimed' : 'open';
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/unresolve`, { method: 'POST' });
  },

  closeTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      t.status = 'closed';
      t.closedAt = new Date().toISOString();
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/close`, { method: 'POST' });
  },

  reopenTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      t.status = 'open';
      t.closedAt = undefined;
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/reopen`, { method: 'POST' });
  },

  archiveTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    if (isDemoMode()) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (!t) throw new Error(`Demo ticket ${ticketId} not found`);
      t.status = 'archived';
      return t;
    }
    return fetchJSON<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/archive`, { method: 'POST' });
  },

  // Reaction Roles
  listReactionRoles: async (guildId: string): Promise<ReactionRole[]> => {
    if (isDemoMode()) return mockReactionRoles.filter(r => r.guildId === guildId);
    return fetchJSON<ReactionRole[]>(`/api/v1/guilds/${guildId}/reaction-roles`);
  },

  createReactionRole: async (guildId: string, rr: Omit<ReactionRole, 'id' | 'guildId'>): Promise<ReactionRole> => {
    if (isDemoMode()) {
      const newRr: ReactionRole = {
        ...rr,
        id: `msg-${Math.floor(Math.random() * 10000)}`,
        guildId
      };
      mockReactionRoles.push(newRr);
      return newRr;
    }
    return fetchJSON<ReactionRole>(`/api/v1/guilds/${guildId}/reaction-roles`, {
      method: 'POST',
      body: JSON.stringify(rr)
    });
  },

  // Persists a single emoji→role mapping onto an existing reaction role message and
  // makes the bot add the reaction on Discord. Called per mapping after the panel is
  // created (the create endpoint itself only posts the embed).
  addReactionRoleEmoji: async (guildId: string, msgId: string, emoji: string, roleId: string): Promise<void> => {
    if (isDemoMode()) return;
    await fetchJSON(`/api/v1/guilds/${guildId}/reaction-roles/${msgId}/emojis`, {
      method: 'POST',
      body: JSON.stringify({ emoji, role_id: roleId })
    });
  },

  deleteReactionRole: async (guildId: string, msgId: string): Promise<void> => {
    if (isDemoMode()) {
      mockReactionRoles = mockReactionRoles.filter(r => r.id !== msgId);
      return;
    }
    await fetch(`${API_BASE}/api/v1/guilds/${guildId}/reaction-roles/${msgId}`, { method: 'DELETE', credentials: 'include' });
  }
};

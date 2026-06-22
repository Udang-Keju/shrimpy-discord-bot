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
  status: 'open' | 'claimed' | 'closed' | 'archived';
  assignedTo?: string;
  categoryName: string;
  createdAt: string;
  closedAt?: string;
}

export interface TicketPanel {
  id: string;
  guildId: string;
  channelId: string;
  title: string;
  description: string;
  buttonLabel: string;
  buttonStyle: 'primary' | 'success' | 'danger' | 'secondary';
}

export interface TicketCategory {
  id: string;
  panelId: string;
  name: string;
  channelId: string;
  logChannelId?: string;
  supportRoles: string[];
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

export interface DiscordUser {
  id: string;
  username: string;
  globalName?: string;
  avatar?: string;
}

const API_BASE = process.env.NEXT_PUBLIC_SHRIMPY_API_URL || 'http://localhost:8080';

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
    title: 'Contact Support Services',
    description: 'Click the button below to open a private support ticket. Our staff is available 24/7.',
    buttonLabel: 'Create Ticket',
    buttonStyle: 'primary'
  }
];

const mockCategories: Record<string, TicketCategory[]> = {
  'panel-001': [
    { id: 'cat-001', panelId: 'panel-001', name: 'General Help', channelId: 'tickets-gen', supportRoles: ['mod-role-id'] },
    { id: 'cat-002', panelId: 'panel-001', name: 'Billing', channelId: 'tickets-bill', supportRoles: ['admin-role-id'] }
  ]
};

const mockChannels: DiscordChannel[] = [
  { id: 'general', name: 'general-chat', type: 'text' },
  { id: 'rules', name: 'rules-and-info', type: 'text' },
  { id: 'announcements', name: 'announcements', type: 'text' },
  { id: 'support-desk', name: 'support-desk', type: 'text' },
  { id: 'logs', name: 'bot-logs', type: 'text' }
];

const mockRoles: DiscordRole[] = [
  { id: 'member', name: 'Server Member', color: '#7289da' },
  { id: 'developer', name: 'Guild Developer', color: '#4f545c' },
  { id: 'gamer', name: 'Community Gamer', color: '#43b581' },
  { id: 'mod-role-id', name: 'Moderator', color: '#faa61a' },
  { id: 'admin-role-id', name: 'Administrator', color: '#f04747' }
];

// Helper to perform safe fetch, falling back to mocks on error or offline
async function safeFetch<T>(url: string, options?: RequestInit, fallbackData?: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${url}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(options?.headers || {}),
      },
    });
    if (!res.ok) {
      if (fallbackData !== undefined) return fallbackData;
      throw new Error(`API Error: ${res.status} ${res.statusText}`);
    }
    return await res.json() as T;
  } catch (e) {
    console.warn(`Fetch to ${url} failed. Using mock data.`, e);
    if (fallbackData !== undefined) return fallbackData;
    throw e;
  }
}

export const ShrimpyAPI = {
  // Auth
  getCurrentUser: async (): Promise<DiscordUser> => {
    return safeFetch<DiscordUser>('/api/v1/auth/me', {}, {
      id: 'discord-user-123456',
      username: 'shrimp_commander',
      globalName: 'Shrimp Commander',
      avatar: 'https://images.unsplash.com/photo-1553753861-267865544b20?w=150'
    });
  },

  getPublicConfig: async (): Promise<PublicConfig> => {
    return safeFetch<PublicConfig>('/api/v1/config', {}, {
      client_id: '123456789012345678',
      redirect_uri: 'http://localhost:8080/api/v1/auth/callback'
    });
  },

  logout: async (): Promise<void> => {
    try {
      await fetch(`${API_BASE}/api/v1/auth/logout`, { method: 'DELETE' });
    } catch {
      console.warn('Logout request failed (offline)');
    }
  },

  // Guilds
  listGuilds: async (): Promise<Guild[]> => {
    return safeFetch<Guild[]>('/api/v1/guilds', {}, mockGuilds);
  },

  getGuildConfig: async (guildId: string): Promise<Guild> => {
    return safeFetch<Guild>(
      `/api/v1/guilds/${guildId}`,
      {},
      mockGuilds.find(g => g.id === guildId) || mockGuilds[0]
    );
  },

  updateGuildConfig: async (guildId: string, updates: Partial<Guild>): Promise<Guild> => {
    try {
      return await safeFetch<Guild>(`/api/v1/guilds/${guildId}`, {
        method: 'PATCH',
        body: JSON.stringify(updates)
      });
    } catch {
      const g = mockGuilds.find(x => x.id === guildId);
      if (g) {
        Object.assign(g, updates);
        return g;
      }
      return mockGuilds[0];
    }
  },

  // Discord Discord API proxies (for dropdown selects)
  getDiscordChannels: async (guildId: string): Promise<DiscordChannel[]> => {
    return safeFetch<DiscordChannel[]>(`/api/v1/guilds/${guildId}/discord/channels`, {}, mockChannels);
  },

  getDiscordRoles: async (guildId: string): Promise<DiscordRole[]> => {
    return safeFetch<DiscordRole[]>(`/api/v1/guilds/${guildId}/discord/roles`, {}, mockRoles);
  },

  // Welcome Config
  getWelcomeConfig: async (guildId: string): Promise<WelcomeConfig> => {
    return safeFetch<WelcomeConfig>(
      `/api/v1/guilds/${guildId}/welcome`,
      {},
      mockWelcomeConfigs[guildId] || {
        guildId,
        sendDm: false,
        dmText: '',
        sendChannel: false,
        channelId: '',
        welcomeText: 'Welcome!',
        cardStyle: 'dark',
        avatarEmoji: '🦐'
      }
    );
  },

  saveWelcomeConfig: async (guildId: string, config: WelcomeConfig): Promise<WelcomeConfig> => {
    try {
      return await safeFetch<WelcomeConfig>(`/api/v1/guilds/${guildId}/welcome`, {
        method: 'PUT',
        body: JSON.stringify(config)
      });
    } catch {
      mockWelcomeConfigs[guildId] = config;
      return config;
    }
  },

  // Ticket Panels
  listPanels: async (guildId: string): Promise<TicketPanel[]> => {
    return safeFetch<TicketPanel[]>(`/api/v1/guilds/${guildId}/panels`, {}, mockPanels.filter(p => p.guildId === guildId));
  },

  createPanel: async (guildId: string, panel: Omit<TicketPanel, 'id' | 'guildId'>): Promise<TicketPanel> => {
    try {
      return await safeFetch<TicketPanel>(`/api/v1/guilds/${guildId}/panels`, {
        method: 'POST',
        body: JSON.stringify(panel)
      });
    } catch {
      const newPanel: TicketPanel = {
        ...panel,
        id: `panel-${Math.floor(Math.random() * 1000)}`,
        guildId
      };
      mockPanels.push(newPanel);
      return newPanel;
    }
  },

  deletePanel: async (guildId: string, panelId: string): Promise<void> => {
    try {
      await fetch(`${API_BASE}/api/v1/guilds/${guildId}/panels/${panelId}`, { method: 'DELETE' });
    } catch {
      mockPanels = mockPanels.filter(p => p.id !== panelId);
    }
  },

  listCategories: async (guildId: string, panelId: string): Promise<TicketCategory[]> => {
    return safeFetch<TicketCategory[]>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories`, {}, mockCategories[panelId] || []);
  },

  createCategory: async (guildId: string, panelId: string, cat: Omit<TicketCategory, 'id' | 'panelId'>): Promise<TicketCategory> => {
    try {
      return await safeFetch<TicketCategory>(`/api/v1/guilds/${guildId}/panels/${panelId}/categories`, {
        method: 'POST',
        body: JSON.stringify(cat)
      });
    } catch {
      const newCat: TicketCategory = {
        ...cat,
        id: `cat-${Math.floor(Math.random() * 1000)}`,
        panelId
      };
      if (!mockCategories[panelId]) mockCategories[panelId] = [];
      mockCategories[panelId].push(newCat);
      return newCat;
    }
  },

  deleteCategory: async (guildId: string, panelId: string, catId: string): Promise<void> => {
    try {
      await fetch(`${API_BASE}/api/v1/guilds/${guildId}/panels/${panelId}/categories/${catId}`, { method: 'DELETE' });
    } catch {
      if (mockCategories[panelId]) {
        mockCategories[panelId] = mockCategories[panelId].filter(c => c.id !== catId);
      }
    }
  },

  // Tickets management
  listTickets: async (guildId: string): Promise<Ticket[]> => {
    return safeFetch<Ticket[]>(`/api/v1/guilds/${guildId}/tickets`, {}, mockTickets.filter(t => t.guildId === guildId));
  },

  updateTicket: async (guildId: string, ticketId: string, updates: Partial<Ticket>): Promise<Ticket> => {
    try {
      return await safeFetch<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}`, {
        method: 'PATCH',
        body: JSON.stringify(updates)
      });
    } catch (e) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (t) {
        Object.assign(t, updates);
        return t;
      }
      throw e;
    }
  },

  claimTicket: async (guildId: string, ticketId: string, username: string): Promise<Ticket> => {
    try {
      // The endpoint is PATCH /tickets/{id} with status claimed or similar
      return await ShrimpyAPI.updateTicket(guildId, ticketId, { status: 'claimed', assignedTo: username });
    } catch (e) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (t) {
        t.status = 'claimed';
        t.assignedTo = username;
        return t;
      }
      throw e;
    }
  },

  closeTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    try {
      return await safeFetch<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/close`, { method: 'POST' });
    } catch (e) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (t) {
        t.status = 'closed';
        t.closedAt = new Date().toISOString();
        return t;
      }
      throw e;
    }
  },

  reopenTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    try {
      return await safeFetch<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/reopen`, { method: 'POST' });
    } catch (e) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (t) {
        t.status = 'open';
        t.closedAt = undefined;
        return t;
      }
      throw e;
    }
  },

  archiveTicket: async (guildId: string, ticketId: string): Promise<Ticket> => {
    try {
      return await safeFetch<Ticket>(`/api/v1/guilds/${guildId}/tickets/${ticketId}/archive`, { method: 'POST' });
    } catch (e) {
      const t = mockTickets.find(x => x.id === ticketId);
      if (t) {
        t.status = 'archived';
        return t;
      }
      throw e;
    }
  },

  // Reaction Roles
  listReactionRoles: async (guildId: string): Promise<ReactionRole[]> => {
    return safeFetch<ReactionRole[]>(`/api/v1/guilds/${guildId}/reaction-roles`, {}, mockReactionRoles.filter(r => r.guildId === guildId));
  },

  createReactionRole: async (guildId: string, rr: Omit<ReactionRole, 'id' | 'guildId'>): Promise<ReactionRole> => {
    try {
      return await safeFetch<ReactionRole>(`/api/v1/guilds/${guildId}/reaction-roles`, {
        method: 'POST',
        body: JSON.stringify(rr)
      });
    } catch {
      const newRr: ReactionRole = {
        ...rr,
        id: `msg-${Math.floor(Math.random() * 10000)}`,
        guildId
      };
      mockReactionRoles.push(newRr);
      return newRr;
    }
  },

  deleteReactionRole: async (guildId: string, msgId: string): Promise<void> => {
    try {
      await fetch(`${API_BASE}/api/v1/guilds/${guildId}/reaction-roles/${msgId}`, { method: 'DELETE' });
    } catch {
      mockReactionRoles = mockReactionRoles.filter(r => r.id !== msgId);
    }
  }
};

// dashboard/app/dashboard/[guildId]/layout.tsx
"use client";

import { useEffect, useState } from "react";
import {
  useParams,
  usePathname,
  useRouter
} from "next/navigation";
import Link from "next/link";
import {
  LogOut,
  Sun,
  Moon
} from "lucide-react";
import styles from "./dashboard.module.css";
import { getNavigationGroups } from "./navConfig";
import { ShrimpyAPI, Guild, DiscordUser, PublicConfig, isDemoMode } from "@/lib/api";
import { getSavedTheme, applyTheme, Theme } from "@/lib/theme";
import DemoBanner from "@/components/DemoBanner";
import Dropdown from "@/components/Dropdown";
import InviteGate from "@/components/InviteGate";

// Scoped invite permissions — kept in sync with the servers page (USER_JOURNEY §14.6).
const INVITE_PERMISSIONS = "17448660048";

export default function DashboardLayout({
  children,
}: { 
  children: React.ReactNode;
}) {
  const params = useParams();
  const pathname = usePathname();
  const router = useRouter();
  
  const guildId = (params?.guildId as string) || "123456789012345678";
  
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");
  const [user, setUser] = useState<DiscordUser | null>(null);
  const [guilds, setGuilds] = useState<Guild[]>([]);
  const [activeGuild, setActiveGuild] = useState<Guild | null>(null);
  const [config, setConfig] = useState<PublicConfig | null>(null);

  // Fetch guilds and user info
  useEffect(() => {
    const timer = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);

    async function loadData() {
      try {
        const [userData, guildList, configData] = await Promise.all([
          ShrimpyAPI.getCurrentUser(),
          ShrimpyAPI.listGuilds(),
          ShrimpyAPI.getPublicConfig()
        ]);
        setUser(userData);
        setGuilds(guildList);
        setConfig(configData);
        const current = guildList.find(g => g.id === guildId) || guildList[0];
        setActiveGuild(current || null);
      } catch (err) {
        console.error("Failed to load dashboard resources", err);
      }
    }
    loadData();

    return () => clearTimeout(timer);
  }, [guildId]);

  const toggleTheme = () => {
    const nextTheme = theme === "dark" ? "light" : "dark";
    setTheme(nextTheme);
    applyTheme(nextTheme);
  };

  const handleGuildChange = (selectedId: string) => {
    if (pathname === `/dashboard/${guildId}`) {
      router.push(`/dashboard/${selectedId}`);
      return;
    }
    const currentTab = pathname.split("/").pop() || "tickets";
    router.push(`/dashboard/${selectedId}/${currentTab}`);
  };

  const isIconUrl = (icon?: string) => !!icon && icon.startsWith("http");

  // Returns null (rather than a fake link) when config hasn't loaded, so we never
  // send users to Discord's "Unknown Application" error with a bogus client_id.
  const getInviteUrl = (targetGuildId: string): string | null => {
    if (!config?.client_id) return null;
    return `https://discord.com/api/oauth2/authorize?client_id=${config.client_id}&permissions=${INVITE_PERMISSIONS}&scope=bot+applications.commands&guild_id=${targetGuildId}`;
  };

  const handleLogout = async () => {
    if (isDemoMode()) {
      ShrimpyAPI.exitDemoMode();
      router.push("/login");
      return;
    }
    await ShrimpyAPI.logout();
    router.push("/login");
  };

  if (!mounted) {
    return null;
  }

  const navigationGroups = getNavigationGroups(guildId);
  const navigationItems = navigationGroups.flatMap(g => g.items);

  // Gate the config pages when the bot isn't in this guild. The Discord-backed
  // config endpoints would otherwise return nothing and the pages render empty.
  // Only block on an explicit `false` so demo guilds (bot_joined undefined) pass through.
  const currentGuild = guilds.find(g => g.id === guildId);
  const needsInvite = currentGuild?.bot_joined === false;

  const getPageTitle = () => {
    if (pathname === `/dashboard/${guildId}`) return "Overview";
    const current = navigationItems.find(item => pathname.startsWith(item.href));
    return current ? current.name : "Dashboard";
  };

  return (
    <div className={styles.container}>
      {/* SIDEBAR */}
      <aside className={styles.sidebar}>
        <div className={styles.sidebarHeader}>
          <span className={styles.brandIcon}>🦐</span>
          <span className={styles.brandText}>Shrimpy</span>
        </div>

        <div className={styles.guildSelectorWrapper}>
          <div className={styles.label} style={{ marginBottom: '6px' }}>Active Guild</div>
          <Dropdown
            value={guildId}
            onChange={handleGuildChange}
            placeholder="Select a server..."
            options={[
              ...guilds.filter(g => g.bot_joined).map(g => ({ value: g.id, label: g.name, icon: g.icon, group: "Invited" })),
              ...guilds.filter(g => !g.bot_joined).map(g => ({ value: g.id, label: g.name, icon: g.icon, group: "Not Invited" })),
            ]}
          />
        </div>

        <nav className={styles.sidebarNav}>
          {navigationGroups.map(group => (
            <div key={group.label} className={styles.navGroup}>
              <div className={styles.navGroupLabel}>{group.label}</div>
              {group.items.map(item => {
                const Icon = item.icon;
                const isActive = pathname.startsWith(item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`${styles.navItem} ${isActive ? styles.navItemActive : ""}`}
                  >
                    <Icon size={18} />
                    <span>{item.name}</span>
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>

        <div className={styles.sidebarFooter}>
          <button onClick={handleLogout} className={`${styles.navItem}`} style={{ width: '100%', background: 'none', border: 'none', textAlign: 'left', cursor: 'pointer' }}>
            <LogOut size={18} />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>

      {/* MAIN VIEW AREA */}
      <div className={styles.mainContent}>
        {isDemoMode() && <DemoBanner />}

        {/* TOPBAR NAVBAR */}
        <header className={styles.topbar}>
          <div className={styles.topbarLeft}>
            <div className={styles.guildBadge}>
              {isIconUrl(activeGuild?.icon) ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={activeGuild!.icon} alt="" className={styles.guildBadgeImg} />
              ) : (
                activeGuild?.icon || "🦐"
              )}
            </div>
            <div>
              <h1 className={styles.pageTitle}>{getPageTitle()}</h1>
            </div>
          </div>

          <div className={styles.topbarRight}>
            <button 
              onClick={toggleTheme} 
              className={styles.logoutBtn} 
              style={{ padding: '6px', color: 'var(--color-text)' }}
              title="Toggle Theme"
            >
              {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
            </button>

            {user && (
              <div className={styles.userProfile}>
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img 
                  src={user.avatar} 
                  alt={user.username} 
                  className={styles.userAvatar} 
                />
                <span className={styles.username}>{user.globalName || user.username}</span>
              </div>
            )}
          </div>
        </header>

        {/* DASHBOARD PAGE PANEL BODY */}
        <main className={styles.contentBody}>
          {needsInvite ? (
            <InviteGate
              guildName={currentGuild?.name || "this server"}
              inviteUrl={getInviteUrl(guildId)}
            />
          ) : (
            children
          )}
        </main>
      </div>
    </div>
  );
}

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
  Ticket,
  UserPlus,
  Tags,
  Settings,
  Layers,
  LogOut,
  Sun,
  Moon
} from "lucide-react";
import styles from "./dashboard.module.css";
import { ShrimpyAPI, Guild, DiscordUser, isDemoMode } from "@/lib/api";
import { getSavedTheme, applyTheme, Theme } from "@/lib/theme";
import DemoBanner from "@/components/DemoBanner";

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

  // Fetch guilds and user info
  useEffect(() => {
    const timer = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);

    async function loadData() {
      try {
        const [userData, guildList] = await Promise.all([
          ShrimpyAPI.getCurrentUser(),
          ShrimpyAPI.listGuilds()
        ]);
        setUser(userData);
        setGuilds(guildList);
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

  const handleGuildChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedId = e.target.value;
    const currentTab = pathname.split("/").pop() || "tickets";
    router.push(`/dashboard/${selectedId}/${currentTab}`);
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

  const navigationGroups = [
    {
      label: "Operate",
      items: [
        { name: "Support Tickets", href: `/dashboard/${guildId}/tickets`, icon: Ticket },
      ],
    },
    {
      label: "Server Management",
      items: [
        { name: "Ticket Panels", href: `/dashboard/${guildId}/panels`, icon: Layers },
        { name: "Welcome Greetings", href: `/dashboard/${guildId}/welcome`, icon: UserPlus },
        { name: "Reaction Roles", href: `/dashboard/${guildId}/roles`, icon: Tags },
      ],
    },
    {
      label: "Settings",
      items: [
        { name: "General Settings", href: `/dashboard/${guildId}/settings`, icon: Settings },
      ],
    },
  ];
  const navigationItems = navigationGroups.flatMap(g => g.items);

  const getPageTitle = () => {
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
          <select 
            className={styles.guildSelect} 
            value={guildId} 
            onChange={handleGuildChange}
          >
            {guilds.map(g => (
              <option key={g.id} value={g.id}>
                {g.icon} {g.name}
              </option>
            ))}
          </select>
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
              {activeGuild?.icon || "🦐"}
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
          {children}
        </main>
      </div>
    </div>
  );
}

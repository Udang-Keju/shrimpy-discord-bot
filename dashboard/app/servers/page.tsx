// dashboard/app/servers/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ShrimpyAPI, Guild, DiscordUser, PublicConfig } from "@/lib/api";
import { LogOut, Plus, Server, Bot, ArrowRight, ExternalLink } from "lucide-react";

// Scoped invite permissions (USER_JOURNEY §14.6): View/Manage Channels, Manage Roles,
// Manage Threads, Send Messages, Embed Links, Read Message History, Add Reactions,
// Use External Emojis, Manage Messages. Replaces the previous Administrator (8) grant.
const INVITE_PERMISSIONS = "17448660048";

export default function ServersPage() {
  const router = useRouter();
  const [user, setUser] = useState<DiscordUser | null>(null);
  const [guilds, setGuilds] = useState<Guild[]>([]);
  const [config, setConfig] = useState<PublicConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [hoveredCard, setHoveredCard] = useState<string | null>(null);
  const [hoveredInvite, setHoveredInvite] = useState(false);

  useEffect(() => {
    async function loadData() {
      const [userResult, guildResult, configResult] = await Promise.allSettled([
        ShrimpyAPI.getCurrentUser(),
        ShrimpyAPI.listGuilds(),
        ShrimpyAPI.getPublicConfig()
      ]);
      if (userResult.status === "fulfilled") {
        setUser(userResult.value);
      } else {
        console.error("Failed to load current user", userResult.reason);
      }
      if (guildResult.status === "fulfilled") {
        setGuilds(guildResult.value);
      } else {
        console.error("Failed to load guild list", guildResult.reason);
      }
      if (configResult.status === "fulfilled") {
        setConfig(configResult.value);
      } else {
        console.error("Failed to load public config", configResult.reason);
      }
      setLoading(false);
    }
    loadData();
  }, []);

  const handleLogout = async () => {
    await ShrimpyAPI.logout();
    router.push("/login");
  };

  // No fallback here: an invite link built from a fake client_id silently
  // sends users to Discord's "Unknown Application" error after redirect.
  const getInviteUrl = (guildId?: string) => {
    if (!config?.client_id) return null;
    let url = `https://discord.com/api/oauth2/authorize?client_id=${config.client_id}&permissions=${INVITE_PERMISSIONS}&scope=bot+applications.commands`;
    if (guildId) {
      url += `&guild_id=${guildId}`;
    }
    return url;
  };

  const getServerInitials = (name: string) => {
    return name
      .split(" ")
      .map((word) => word[0])
      .join("")
      .slice(0, 3)
      .toUpperCase();
  };

  if (loading) {
    return (
      <div style={{ display: "flex", height: "100vh", width: "100vw", justifyContent: "center", alignItems: "center", backgroundColor: "var(--color-background)", color: "var(--color-text)", fontFamily: "var(--font-display)" }}>
        <div style={{ textAlign: "center" }}>
          <div style={{ fontSize: "36px", marginBottom: "16px", animation: "float 2s ease-in-out infinite" }}>🦐</div>
          <p style={{ color: "var(--color-text-muted)", fontSize: "14px", letterSpacing: "0.5px" }}>Loading Shrimpy Console...</p>
        </div>
        <style jsx global>{`
          @keyframes float {
            0%, 100% { transform: translateY(0); }
            50% { transform: translateY(-10px); }
          }
        `}</style>
      </div>
    );
  }

  return (
    <div style={{ minHeight: "100vh", width: "100%", backgroundColor: "var(--color-background)", color: "var(--color-text)", fontFamily: "var(--font-display)", position: "relative", overflowX: "hidden" }}>
      {/* Dynamic Radial Background Glow */}
      <div style={{ position: "absolute", top: 0, left: "50%", transform: "translateX(-50%)", width: "100%", maxWidth: "1200px", height: "500px", background: "radial-gradient(circle at top, var(--primary-muted), transparent 70%)", pointerEvents: "none", zIndex: 1 }} />

      {/* HEADER */}
      <header style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "20px 40px", borderBottom: "1px solid var(--color-border)", position: "relative", zIndex: 10, backdropFilter: "blur(12px)", backgroundColor: "var(--color-surface)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <span style={{ fontSize: "28px" }}>🦐</span>
          <span style={{ fontWeight: 800, fontSize: "22px", letterSpacing: "-0.5px", background: "linear-gradient(to right, var(--primary), var(--accent))", WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>Shrimpy Console</span>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "24px" }}>
          {user && (
            <div style={{ display: "flex", alignItems: "center", gap: "12px", padding: "6px 14px", borderRadius: "100px", backgroundColor: "var(--bg-surface-hover)", border: "1px solid var(--color-border)" }}>
              {user.avatar ? (
                <img src={user.avatar} alt={user.username} style={{ width: "28px", height: "28px", borderRadius: "50%" }} />
              ) : (
                <div style={{ width: "28px", height: "28px", borderRadius: "50%", backgroundColor: "var(--color-primary)", display: "flex", justifyContent: "center", alignItems: "center", fontSize: "12px" }}>👤</div>
              )}
              <span style={{ fontSize: "14px", fontWeight: 600 }}>{user.globalName || user.username}</span>
            </div>
          )}
          <button
            onClick={handleLogout}
            style={{ display: "flex", alignItems: "center", gap: "8px", background: "none", border: "none", color: "var(--color-text-muted)", cursor: "pointer", fontSize: "14px", fontWeight: 500, transition: "color 0.2s" }}
            onMouseEnter={(e) => e.currentTarget.style.color = "var(--color-danger)"}
            onMouseLeave={(e) => e.currentTarget.style.color = "var(--color-text-muted)"}
          >
            <LogOut size={16} />
            <span>Sign Out</span>
          </button>
        </div>
      </header>

      {/* CONTENT BODY */}
      <main style={{ maxWidth: "1200px", margin: "0 auto", padding: "60px 40px", position: "relative", zIndex: 10 }}>
        <div style={{ textAlign: "center", marginBottom: "48px" }}>
          <h1 style={{ fontSize: "40px", fontWeight: 800, letterSpacing: "-1px", marginBottom: "12px", color: "var(--color-text)" }}>Select a Server</h1>
          <p style={{ color: "var(--color-text-muted)", fontSize: "16px", maxWidth: "600px", margin: "0 auto" }}>Choose a server to configure support tickets, welcome messages, and reaction roles. Only servers where you have Administrator permissions are listed.</p>
        </div>

        {guilds.length === 0 ? (
          /* EMPTY STATE */
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", padding: "60px 40px", borderRadius: "20px", backgroundColor: "var(--color-surface)", border: "1px solid var(--color-border)", backdropFilter: "blur(12px)", textAlign: "center", maxWidth: "550px", margin: "0 auto" }}>
            <div style={{ width: "64px", height: "64px", borderRadius: "16px", backgroundColor: "var(--primary-muted)", display: "flex", justifyContent: "center", alignItems: "center", color: "var(--color-primary)", marginBottom: "20px" }}>
              <Server size={32} />
            </div>
            <h3 style={{ fontSize: "20px", fontWeight: 700, marginBottom: "8px" }}>No Servers Found</h3>
            <p style={{ color: "var(--color-text-muted)", fontSize: "14px", marginBottom: "28px", lineHeight: "1.5" }}>We couldn&apos;t detect any servers where you possess Administrator permissions. Make sure you are logged into the correct Discord account.</p>
            {getInviteUrl() ? (
              <a
                href={getInviteUrl()!}
                target="_blank"
                rel="noopener noreferrer"
                style={{ display: "flex", alignItems: "center", gap: "10px", padding: "12px 24px", borderRadius: "10px", backgroundColor: "var(--color-primary)", color: "var(--color-primary-fg)", textDecoration: "none", fontSize: "14px", fontWeight: 600, transition: "background-color 0.2s" }}
                onMouseEnter={(e) => e.currentTarget.style.backgroundColor = "var(--primary-hover)"}
                onMouseLeave={(e) => e.currentTarget.style.backgroundColor = "var(--color-primary)"}
              >
                <Bot size={18} />
                <span>Invite Bot to a Server</span>
                <ExternalLink size={14} />
              </a>
            ) : (
              <p style={{ color: "var(--color-danger)", fontSize: "13px" }}>Invite link unavailable — couldn&apos;t load app configuration. Try refreshing the page.</p>
            )}
          </div>
        ) : (
          /* GUILD LIST GRID */
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: "24px" }}>
            {guilds.map((g) => {
              const isHovered = hoveredCard === g.id;
              const hasJoined = g.bot_joined === true;

              return (
                <div
                  key={g.id}
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "space-between",
                    padding: "24px",
                    borderRadius: "16px",
                    backgroundColor: "var(--color-surface)",
                    border: isHovered ? "1px solid var(--border-focus)" : "1px solid var(--color-border)",
                    backdropFilter: "blur(12px)",
                    transform: isHovered ? "translateY(-4px)" : "translateY(0)",
                    boxShadow: isHovered ? "var(--shadow-primary)" : "none",
                    transition: "all 0.25s cubic-bezier(0.4, 0, 0.2, 1)",
                    cursor: "default"
                  }}
                  onMouseEnter={() => setHoveredCard(g.id)}
                  onMouseLeave={() => setHoveredCard(null)}
                >
                  <div>
                    {/* Icon and Title */}
                    <div style={{ display: "flex", alignItems: "center", gap: "16px", marginBottom: "20px" }}>
                      {g.icon ? (
                        <div style={{ width: "52px", height: "52px", borderRadius: "12px", overflow: "hidden", display: "flex", justifyContent: "center", alignItems: "center", border: "1px solid var(--color-border)" }}>
                          {g.icon.length > 4 ? (
                            <img src={g.icon} alt={g.name} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                          ) : (
                            <span style={{ fontSize: "22px" }}>{g.icon}</span>
                          )}
                        </div>
                      ) : (
                        <div style={{ width: "52px", height: "52px", borderRadius: "12px", backgroundColor: "var(--color-surface-raised)", display: "flex", justifyContent: "center", alignItems: "center", fontSize: "14px", fontWeight: 700, color: "var(--color-text)", border: "1px solid var(--color-border)" }}>
                          {getServerInitials(g.name)}
                        </div>
                      )}

                      <div style={{ flex: 1, minWidth: 0 }}>
                        <h3 style={{ fontSize: "16px", fontWeight: 700, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", margin: 0 }}>{g.name}</h3>
                        <div style={{ display: "flex", alignItems: "center", gap: "6px", marginTop: "4px" }}>
                          <span style={{ width: "6px", height: "6px", borderRadius: "50%", backgroundColor: hasJoined ? "var(--color-success)" : "var(--color-text-muted)" }} />
                          <span style={{ fontSize: "12px", color: hasJoined ? "var(--color-success)" : "var(--color-text-muted)", fontWeight: 500 }}>
                            {hasJoined ? "Active" : "Invite Needed"}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div>
                    {hasJoined ? (
                      <button
                        onClick={() => router.push(`/dashboard/${g.id}/tickets`)}
                        style={{
                          width: "100%",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          gap: "8px",
                          padding: "11px",
                          borderRadius: "10px",
                          backgroundColor: isHovered ? "var(--color-primary)" : "var(--primary-muted)",
                          border: isHovered ? "1px solid var(--color-primary)" : "1px solid var(--primary-muted)",
                          color: isHovered ? "var(--color-primary-fg)" : "var(--color-primary)",
                          fontSize: "14px",
                          fontWeight: 600,
                          cursor: "pointer",
                          transition: "all 0.2s"
                        }}
                      >
                        <span>Configure Dashboard</span>
                        <ArrowRight size={16} />
                      </button>
                    ) : getInviteUrl(g.id) ? (
                      <a
                        href={getInviteUrl(g.id)!}
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{
                          width: "100%",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          gap: "8px",
                          padding: "10px",
                          borderRadius: "10px",
                          backgroundColor: "var(--color-surface-raised)",
                          border: "1px solid var(--color-border)",
                          color: "var(--color-text)",
                          fontSize: "14px",
                          fontWeight: 600,
                          textDecoration: "none",
                          textAlign: "center",
                          boxSizing: "border-box",
                          transition: "all 0.2s"
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = "var(--primary-muted)";
                          e.currentTarget.style.borderColor = "var(--color-primary)";
                          e.currentTarget.style.color = "var(--color-primary)";
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = "var(--color-surface-raised)";
                          e.currentTarget.style.borderColor = "var(--color-border)";
                          e.currentTarget.style.color = "var(--color-text)";
                        }}
                      >
                        <Bot size={16} />
                        <span>Setup Shrimpy</span>
                        <ExternalLink size={14} />
                      </a>
                    ) : (
                      <div
                        style={{
                          width: "100%",
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                          gap: "8px",
                          padding: "10px",
                          borderRadius: "10px",
                          border: "1px solid var(--color-border)",
                          color: "var(--color-text-muted)",
                          fontSize: "13px",
                          boxSizing: "border-box"
                        }}
                      >
                        <span>Config unavailable</span>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}

            {/* Permanent "Invite Bot to New Server" Card */}
            {getInviteUrl() ? (
              <a
                href={getInviteUrl()!}
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  justifyContent: "center",
                  padding: "24px",
                  borderRadius: "16px",
                  border: hoveredInvite ? "1px dashed var(--color-primary)" : "1px dashed var(--color-border)",
                  backgroundColor: hoveredInvite ? "var(--primary-muted)" : "transparent",
                  color: hoveredInvite ? "var(--color-primary)" : "var(--color-text-muted)",
                  textDecoration: "none",
                  cursor: "pointer",
                  transition: "all 0.25s cubic-bezier(0.4, 0, 0.2, 1)",
                  minHeight: "155px",
                  boxSizing: "border-box"
                }}
                onMouseEnter={() => setHoveredInvite(true)}
                onMouseLeave={() => setHoveredInvite(false)}
              >
                <div style={{ width: "40px", height: "40px", borderRadius: "50%", backgroundColor: hoveredInvite ? "var(--primary-muted)" : "var(--color-surface-raised)", display: "flex", justifyContent: "center", alignItems: "center", marginBottom: "12px", transition: "all 0.2s" }}>
                  <Plus size={20} />
                </div>
                <span style={{ fontSize: "15px", fontWeight: 700 }}>Invite to New Server</span>
                <span style={{ fontSize: "12px", color: "var(--color-text-muted)", marginTop: "4px", textAlign: "center" }}>Add Shrimpy to another community</span>
              </a>
            ) : (
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  justifyContent: "center",
                  padding: "24px",
                  borderRadius: "16px",
                  border: "1px dashed var(--color-border)",
                  color: "var(--color-text-muted)",
                  minHeight: "155px",
                  boxSizing: "border-box"
                }}
              >
                <span style={{ fontSize: "13px" }}>Invite unavailable — refresh to retry</span>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}

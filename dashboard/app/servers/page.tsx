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
      <div style={{ display: "flex", height: "100vh", width: "100vw", justifyContent: "center", alignItems: "center", backgroundColor: "#06070a", color: "#fff", fontFamily: "'Outfit', sans-serif" }}>
        <div style={{ textAlign: "center" }}>
          <div style={{ fontSize: "36px", marginBottom: "16px", animation: "float 2s ease-in-out infinite" }}>🦐</div>
          <p style={{ color: "#717f96", fontSize: "14px", letterSpacing: "0.5px" }}>Loading Shrimpy Console...</p>
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
    <div style={{ minHeight: "100vh", width: "100%", backgroundColor: "#06070a", color: "#f8fafc", fontFamily: "'Outfit', sans-serif", position: "relative", overflowX: "hidden" }}>
      {/* Dynamic Radial Background Glow */}
      <div style={{ position: "absolute", top: 0, left: "50%", transform: "translateX(-50%)", width: "100%", maxWidth: "1200px", height: "500px", background: "radial-gradient(circle at top, rgba(79, 70, 229, 0.15), transparent 70%)", pointerEvents: "none", zIndex: 1 }} />

      {/* HEADER */}
      <header style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "20px 40px", borderBottom: "1px solid rgba(255, 255, 255, 0.05)", position: "relative", zIndex: 10, backdropFilter: "blur(12px)", backgroundColor: "rgba(6, 7, 10, 0.8)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <span style={{ fontSize: "28px" }}>🦐</span>
          <span style={{ fontWeight: 800, fontSize: "22px", letterSpacing: "-0.5px", background: "linear-gradient(to right, #818cf8, #c084fc)", WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>Shrimpy Console</span>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "24px" }}>
          {user && (
            <div style={{ display: "flex", alignItems: "center", gap: "12px", padding: "6px 14px", borderRadius: "100px", backgroundColor: "rgba(255, 255, 255, 0.03)", border: "1px solid rgba(255, 255, 255, 0.05)" }}>
              {user.avatar ? (
                <img src={user.avatar} alt={user.username} style={{ width: "28px", height: "28px", borderRadius: "50%" }} />
              ) : (
                <div style={{ width: "28px", height: "28px", borderRadius: "50%", backgroundColor: "#4f46e5", display: "flex", justifyContent: "center", alignItems: "center", fontSize: "12px" }}>👤</div>
              )}
              <span style={{ fontSize: "14px", fontWeight: 600 }}>{user.globalName || user.username}</span>
            </div>
          )}
          <button
            onClick={handleLogout}
            style={{ display: "flex", alignItems: "center", gap: "8px", background: "none", border: "none", color: "#94a3b8", cursor: "pointer", fontSize: "14px", fontWeight: 500, transition: "color 0.2s" }}
            onMouseEnter={(e) => e.currentTarget.style.color = "#ef4444"}
            onMouseLeave={(e) => e.currentTarget.style.color = "#94a3b8"}
          >
            <LogOut size={16} />
            <span>Sign Out</span>
          </button>
        </div>
      </header>

      {/* CONTENT BODY */}
      <main style={{ maxWidth: "1200px", margin: "0 auto", padding: "60px 40px", position: "relative", zIndex: 10 }}>
        <div style={{ textAlign: "center", marginBottom: "48px" }}>
          <h1 style={{ fontSize: "40px", fontWeight: 800, letterSpacing: "-1px", marginBottom: "12px", background: "linear-gradient(to right, #f8fafc, #cbd5e1)", WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>Select a Server</h1>
          <p style={{ color: "#94a3b8", fontSize: "16px", maxWidth: "600px", margin: "0 auto" }}>Choose a server to configure support tickets, welcome messages, and reaction roles. Only servers where you have Administrator permissions are listed.</p>
        </div>

        {guilds.length === 0 ? (
          /* EMPTY STATE */
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", padding: "60px 40px", borderRadius: "20px", backgroundColor: "rgba(17, 18, 25, 0.65)", border: "1px solid rgba(255, 255, 255, 0.08)", backdropFilter: "blur(12px)", textAlign: "center", maxWidth: "550px", margin: "0 auto" }}>
            <div style={{ width: "64px", height: "64px", borderRadius: "16px", backgroundColor: "rgba(79, 70, 229, 0.1)", display: "flex", justifyContent: "center", alignItems: "center", color: "#6366f1", marginBottom: "20px" }}>
              <Server size={32} />
            </div>
            <h3 style={{ fontSize: "20px", fontWeight: 700, marginBottom: "8px" }}>No Servers Found</h3>
            <p style={{ color: "#94a3b8", fontSize: "14px", marginBottom: "28px", lineHeight: "1.5" }}>We couldn&apos;t detect any servers where you possess Administrator permissions. Make sure you are logged into the correct Discord account.</p>
            {getInviteUrl() ? (
              <a
                href={getInviteUrl()!}
                target="_blank"
                rel="noopener noreferrer"
                style={{ display: "flex", alignItems: "center", gap: "10px", padding: "12px 24px", borderRadius: "10px", backgroundColor: "#4f46e5", color: "#fff", textDecoration: "none", fontSize: "14px", fontWeight: 600, transition: "background-color 0.2s" }}
                onMouseEnter={(e) => e.currentTarget.style.backgroundColor = "#4338ca"}
                onMouseLeave={(e) => e.currentTarget.style.backgroundColor = "#4f46e5"}
              >
                <Bot size={18} />
                <span>Invite Bot to a Server</span>
                <ExternalLink size={14} />
              </a>
            ) : (
              <p style={{ color: "#ef4444", fontSize: "13px" }}>Invite link unavailable — couldn&apos;t load app configuration. Try refreshing the page.</p>
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
                    backgroundColor: "rgba(17, 18, 25, 0.65)",
                    border: isHovered ? "1px solid rgba(99, 102, 241, 0.4)" : "1px solid rgba(255, 255, 255, 0.08)",
                    backdropFilter: "blur(12px)",
                    transform: isHovered ? "translateY(-4px)" : "translateY(0)",
                    boxShadow: isHovered ? "0 12px 24px -10px rgba(79, 70, 229, 0.2)" : "none",
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
                        <div style={{ width: "52px", height: "52px", borderRadius: "12px", overflow: "hidden", display: "flex", justifyContent: "center", alignItems: "center", border: "1px solid rgba(255, 255, 255, 0.05)" }}>
                          {g.icon.length > 4 ? (
                            <img src={g.icon} alt={g.name} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                          ) : (
                            <span style={{ fontSize: "22px" }}>{g.icon}</span>
                          )}
                        </div>
                      ) : (
                        <div style={{ width: "52px", height: "52px", borderRadius: "12px", backgroundColor: "#1e293b", display: "flex", justifyContent: "center", alignItems: "center", fontSize: "14px", fontWeight: 700, color: "#cbd5e1", border: "1px solid rgba(255, 255, 255, 0.05)" }}>
                          {getServerInitials(g.name)}
                        </div>
                      )}

                      <div style={{ flex: 1, minWidth: 0 }}>
                        <h3 style={{ fontSize: "16px", fontWeight: 700, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", margin: 0 }}>{g.name}</h3>
                        <div style={{ display: "flex", alignItems: "center", gap: "6px", marginTop: "4px" }}>
                          <span style={{ width: "6px", height: "6px", borderRadius: "50%", backgroundColor: hasJoined ? "#10b981" : "#64748b" }} />
                          <span style={{ fontSize: "12px", color: hasJoined ? "#34d399" : "#94a3b8", fontWeight: 500 }}>
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
                          backgroundColor: isHovered ? "#4f46e5" : "rgba(79, 70, 229, 0.1)",
                          border: isHovered ? "1px solid #6366f1" : "1px solid rgba(99, 102, 241, 0.2)",
                          color: isHovered ? "#fff" : "#818cf8",
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
                          backgroundColor: "rgba(255, 255, 255, 0.03)",
                          border: "1px solid rgba(255, 255, 255, 0.08)",
                          color: "#fff",
                          fontSize: "14px",
                          fontWeight: 600,
                          textDecoration: "none",
                          textAlign: "center",
                          boxSizing: "border-box",
                          transition: "all 0.2s"
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.backgroundColor = "rgba(79, 70, 229, 0.15)";
                          e.currentTarget.style.borderColor = "rgba(99, 102, 241, 0.3)";
                          e.currentTarget.style.color = "#818cf8";
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.backgroundColor = "rgba(255, 255, 255, 0.03)";
                          e.currentTarget.style.borderColor = "rgba(255, 255, 255, 0.08)";
                          e.currentTarget.style.color = "#fff";
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
                          border: "1px solid rgba(255, 255, 255, 0.05)",
                          color: "#64748b",
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
                  border: hoveredInvite ? "1px dashed rgba(99, 102, 241, 0.6)" : "1px dashed rgba(255, 255, 255, 0.15)",
                  backgroundColor: hoveredInvite ? "rgba(79, 70, 229, 0.03)" : "transparent",
                  color: hoveredInvite ? "#818cf8" : "#94a3b8",
                  textDecoration: "none",
                  cursor: "pointer",
                  transition: "all 0.25s cubic-bezier(0.4, 0, 0.2, 1)",
                  minHeight: "155px",
                  boxSizing: "border-box"
                }}
                onMouseEnter={() => setHoveredInvite(true)}
                onMouseLeave={() => setHoveredInvite(false)}
              >
                <div style={{ width: "40px", height: "40px", borderRadius: "50%", backgroundColor: hoveredInvite ? "rgba(79, 70, 229, 0.1)" : "rgba(255, 255, 255, 0.03)", display: "flex", justifyContent: "center", alignItems: "center", marginBottom: "12px", transition: "all 0.2s" }}>
                  <Plus size={20} />
                </div>
                <span style={{ fontSize: "15px", fontWeight: 700 }}>Invite to New Server</span>
                <span style={{ fontSize: "12px", color: "#64748b", marginTop: "4px", textAlign: "center" }}>Add Shrimpy to another community</span>
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
                  border: "1px dashed rgba(255, 255, 255, 0.08)",
                  color: "#64748b",
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

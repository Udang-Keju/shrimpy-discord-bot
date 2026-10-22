// dashboard/app/admin/apps/page.tsx
// USER_JOURNEY A.12 (status + reconnect only): bot-wide screen for the Discord applications the
// backend runs. Viewing status is open to the bot owner and to Administrators of a guild with the
// bot; restarting a bot (which affects every server) is bot-owner only. Both are enforced
// server-side — a 403 on the list renders the access-denied state.
"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ShrimpyAPI, DiscordApp } from "@/lib/api";
import { useToast } from "@/hooks/useToast";
import { ArrowLeft, Bot, RefreshCw, ShieldAlert } from "lucide-react";

const STATUS_POLL_MS = 30_000;
const RECONNECT_SETTLE_MS = 1500;
const RECONNECT_MAX_CHECKS = 4;

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

type LoadState = "loading" | "ready" | "denied" | "error";

export default function AdminAppsPage() {
  const router = useRouter();
  const { showToast, updateToast } = useToast();
  const [apps, setApps] = useState<DiscordApp[]>([]);
  const [state, setState] = useState<LoadState>("loading");
  const [reconnecting, setReconnecting] = useState<Set<string>>(new Set());
  const [isOwner, setIsOwner] = useState(false);
  // Synchronous double-click guard; `reconnecting` state alone lags a render behind.
  const inFlight = useRef<Set<string>>(new Set());

  const loadApps = useCallback(async (): Promise<DiscordApp[] | null> => {
    try {
      const list = await ShrimpyAPI.listApps();
      setApps(list);
      setState("ready");
      return list;
    } catch (err) {
      const message = err instanceof Error ? err.message : "";
      if (message.includes("403")) {
        setState("denied");
      } else {
        console.error("Failed to load bot applications", err);
        // Keep showing the last good list on a transient failure during background polling.
        setState((prev) => (prev === "ready" ? prev : "error"));
      }
      return null;
    }
  }, []);

  useEffect(() => {
    async function load() {
      // The server enforces owner-only; this just decides whether to show the Reconnect button.
      ShrimpyAPI.getCurrentUser()
        .then((user) => setIsOwner(user.isOwner === true))
        .catch((err) => console.error("Failed to load current user", err));
      await loadApps();
    }
    load();
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") loadApps();
    }, STATUS_POLL_MS);
    return () => clearInterval(timer);
  }, [loadApps]);

  const handleReconnect = async (app: DiscordApp) => {
    if (inFlight.current.has(app.id)) return;
    inFlight.current.add(app.id);
    setReconnecting((prev) => new Set(prev).add(app.id));

    const toastId = showToast(`Reconnecting ${app.name}…`, "loading");
    try {
      await ShrimpyAPI.reconnectApp(app.id);

      // The gateway needs a moment to start heartbeating; poll briefly for a healthy state.
      let connected = false;
      for (let i = 0; i < RECONNECT_MAX_CHECKS && !connected; i++) {
        await sleep(RECONNECT_SETTLE_MS);
        const list = await loadApps();
        connected = list?.find((a) => a.id === app.id)?.connected ?? false;
      }

      if (connected) {
        updateToast(toastId, `${app.name} reconnected.`, "success");
      } else {
        updateToast(toastId, `Reconnect sent, but ${app.name} isn't reporting healthy yet. Check again shortly.`, "warning");
      }
    } catch (err) {
      console.error("Failed to reconnect bot application", err);
      updateToast(toastId, `Failed to reconnect ${app.name}.`, "error");
    } finally {
      inFlight.current.delete(app.id);
      setReconnecting((prev) => {
        const next = new Set(prev);
        next.delete(app.id);
        return next;
      });
    }
  };

  const shell = (children: React.ReactNode) => (
    <div style={{ minHeight: "100vh", width: "100%", backgroundColor: "var(--color-background)", color: "var(--color-text)", fontFamily: "var(--font-display)" }}>
      <header style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "20px 40px", borderBottom: "1px solid var(--color-border)", backgroundColor: "var(--color-surface)" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <span style={{ fontSize: "28px" }}>🦐</span>
          <span style={{ fontWeight: 800, fontSize: "22px", letterSpacing: "-0.5px" }}>Shrimpy · Admin</span>
        </div>
        <button
          onClick={() => router.push("/servers")}
          style={{ display: "flex", alignItems: "center", gap: "8px", background: "none", border: "none", color: "var(--color-text-muted)", cursor: "pointer", fontSize: "14px", fontWeight: 500 }}
        >
          <ArrowLeft size={16} />
          <span>Back to servers</span>
        </button>
      </header>
      <main style={{ maxWidth: "820px", margin: "0 auto", padding: "48px 24px" }}>{children}</main>
    </div>
  );

  if (state === "loading") {
    return shell(<p style={{ color: "var(--color-text-muted)", textAlign: "center" }}>Loading applications…</p>);
  }

  if (state === "denied") {
    return shell(
      <div style={{ textAlign: "center", padding: "48px 24px", borderRadius: "16px", backgroundColor: "var(--color-surface)", border: "1px solid var(--color-border)" }}>
        <ShieldAlert size={36} style={{ color: "var(--color-warning)", marginBottom: "12px" }} />
        <h1 style={{ fontSize: "22px", fontWeight: 700, margin: "0 0 8px" }}>Access denied</h1>
        <p style={{ color: "var(--color-text-muted)", fontSize: "14px", margin: 0, lineHeight: 1.5 }}>
          This page is for the bot owner and for administrators of a server where Shrimpy is installed.
        </p>
      </div>
    );
  }

  if (state === "error") {
    return shell(
      <div style={{ textAlign: "center", padding: "48px 24px", borderRadius: "16px", backgroundColor: "var(--color-surface)", border: "1px solid var(--color-border)" }}>
        <p style={{ color: "var(--color-danger)", fontSize: "14px", marginTop: 0 }}>Couldn&apos;t load bot applications.</p>
        <button
          onClick={() => { setState("loading"); loadApps(); }}
          style={{ padding: "10px 20px", borderRadius: "10px", border: "1px solid var(--color-border)", backgroundColor: "var(--color-surface-raised)", color: "var(--color-text)", fontWeight: 600, cursor: "pointer" }}
        >
          Retry
        </button>
      </div>
    );
  }

  return shell(
    <>
      <h1 style={{ fontSize: "32px", fontWeight: 800, letterSpacing: "-0.5px", margin: "0 0 8px" }}>Discord Applications</h1>
      <p style={{ color: "var(--color-text-muted)", fontSize: "14px", margin: "0 0 32px", lineHeight: 1.5 }}>
        The backend automatically reconnects a bot that stays offline for a few minutes.{" "}
        {isOwner
          ? "If a bot still shows as offline in Discord, or a server is stuck on “Invite Needed” after inviting it, you can reconnect it here right away. This briefly interrupts the bot in every server."
          : "A bot is one connection shared by every server, so only the bot owner can reconnect it manually."}
      </p>

      <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
        {apps.map((app) => {
          const busy = reconnecting.has(app.id);
          const color = app.connected ? "var(--color-success)" : "var(--color-danger)";
          return (
            <div key={app.id} style={{ display: "flex", alignItems: "center", gap: "16px", padding: "20px 24px", borderRadius: "16px", backgroundColor: "var(--color-surface)", border: "1px solid var(--color-border)" }}>
              <Bot size={28} style={{ color: "var(--color-text-muted)", flexShrink: 0 }} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                  <span style={{ width: "8px", height: "8px", borderRadius: "50%", backgroundColor: color, flexShrink: 0 }} />
                  <span style={{ fontSize: "12px", fontWeight: 600, color }}>{app.connected ? "Connected" : "Disconnected"}</span>
                </div>
                <h2 style={{ fontSize: "17px", fontWeight: 700, margin: "4px 0 2px", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{app.name}</h2>
                <span style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>Client ID {app.discordClientId}</span>
              </div>
              {isOwner && (
                <button
                  onClick={() => handleReconnect(app)}
                  disabled={busy}
                  style={{ display: "flex", alignItems: "center", gap: "8px", padding: "10px 16px", borderRadius: "10px", border: "1px solid var(--color-border)", backgroundColor: "var(--color-surface-raised)", color: "var(--color-text)", fontSize: "14px", fontWeight: 600, cursor: busy ? "default" : "pointer", opacity: busy ? 0.7 : 1, flexShrink: 0 }}
                >
                  <RefreshCw size={16} style={busy ? { animation: "spin 0.8s linear infinite" } : undefined} />
                  <span>{busy ? "Reconnecting…" : "Reconnect"}</span>
                </button>
              )}
            </div>
          );
        })}
        {apps.length === 0 && (
          <p style={{ color: "var(--color-text-muted)", textAlign: "center" }}>No bot applications are configured.</p>
        )}
      </div>

      <style jsx global>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </>
  );
}

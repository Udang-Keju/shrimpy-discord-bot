// dashboard/components/InviteGate/InviteGate.tsx
"use client";

import Link from "next/link";
import { Bot, ExternalLink, ArrowLeft } from "lucide-react";

interface InviteGateProps {
  guildName: string;
  inviteUrl: string | null;
}

// Shown in place of a guild's configuration pages when Shrimpy has not yet been
// invited to that server. Without the bot in the guild, the Discord-backed config
// endpoints (channels, roles, etc.) can't resolve, so the underlying pages would
// render empty. This prompts the user to invite the bot before configuring.
export default function InviteGate({ guildName, inviteUrl }: InviteGateProps) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "60vh",
        padding: "var(--space-6)",
      }}
    >
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          textAlign: "center",
          maxWidth: "480px",
          width: "100%",
          padding: "var(--space-10) var(--space-8)",
          borderRadius: "var(--radius-lg)",
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          boxShadow: "var(--shadow-md)",
        }}
      >
        <div
          style={{
            width: "64px",
            height: "64px",
            borderRadius: "16px",
            backgroundColor: "var(--primary-muted)",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            color: "var(--color-primary)",
            marginBottom: "var(--space-5)",
          }}
        >
          <Bot size={32} />
        </div>

        <h2
          style={{
            fontFamily: "var(--font-display)",
            fontSize: "var(--text-xl)",
            fontWeight: 700,
            color: "var(--color-text)",
            margin: 0,
          }}
        >
          Invite Shrimpy first
        </h2>

        <p
          style={{
            color: "var(--color-text-muted)",
            fontSize: "var(--text-sm)",
            lineHeight: 1.6,
            margin: "var(--space-3) 0 var(--space-6)",
          }}
        >
          Shrimpy isn&apos;t in <strong style={{ color: "var(--color-text)" }}>{guildName}</strong> yet.
          Add the bot to this server to unlock its dashboard and start configuring tickets,
          welcome greetings, and reaction roles.
        </p>

        {inviteUrl ? (
          <a
            href={inviteUrl}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              display: "flex",
              alignItems: "center",
              gap: "10px",
              padding: "12px 24px",
              borderRadius: "var(--radius-md)",
              backgroundColor: "var(--color-primary)",
              color: "var(--color-primary-fg)",
              textDecoration: "none",
              fontSize: "var(--text-sm)",
              fontWeight: 600,
              transition: "background-color var(--transition-fast)",
            }}
            onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = "var(--primary-hover)")}
            onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = "var(--color-primary)")}
          >
            <Bot size={18} />
            <span>Invite Shrimpy to {guildName}</span>
            <ExternalLink size={14} />
          </a>
        ) : (
          <p style={{ color: "var(--color-danger)", fontSize: "var(--text-sm)" }}>
            Invite link unavailable — couldn&apos;t load app configuration. Try refreshing the page.
          </p>
        )}

        <Link
          href="/servers"
          style={{
            display: "flex",
            alignItems: "center",
            gap: "6px",
            marginTop: "var(--space-5)",
            color: "var(--color-text-muted)",
            fontSize: "var(--text-sm)",
            textDecoration: "none",
            fontWeight: 500,
          }}
        >
          <ArrowLeft size={14} />
          <span>Back to server selection</span>
        </Link>
      </div>
    </div>
  );
}

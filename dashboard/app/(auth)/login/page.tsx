// dashboard/app/(auth)/login/page.tsx
"use client";

import { Sparkles, ChevronRight } from "lucide-react";
import styles from "@/app/page.module.css";
import Link from "next/link";

export default function LoginPage() {
  const loginUrl = `${process.env.NEXT_PUBLIC_SHRIMPY_API_URL || "http://localhost:8080"}/api/v1/auth/login`;

  return (
    <div className={styles.wrapper} style={{ justifyContent: 'center', alignItems: 'center' }}>
      <div className={styles.gridOverlay}></div>

      <div className={`${styles.welcomeCard} glass-effect`} style={{ width: '100%', maxWidth: '440px', padding: '40px var(--space-6)', zIndex: 10 }}>
        <div className={styles.cardBlob}></div>
        
        <div className={styles.welcomeUserAvatar} style={{ animation: 'float 3s ease-in-out infinite' }}>
          🦐
        </div>
        
        <h2 className={styles.welcomeTitle} style={{ fontSize: 'var(--text-2xl)', fontWeight: 800 }}>
          Welcome to <span className="gradient-text">Shrimpy</span> Console
        </h2>
        
        <p className={styles.welcomeDesc} style={{ marginBottom: 'var(--space-6)' }}>
          Authorize your account to configure ticketing systems, welcome onboarding, and self-assignable reaction roles.
        </p>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', width: '100%' }}>
          <a 
            href={loginUrl}
            className={`${styles.btn} ${styles.discordBtn}`} 
            style={{ width: '100%', justifyContent: 'center', padding: '12px' }}
          >
            <Sparkles size={16} />
            <span>Login with Discord</span>
          </a>

          <Link 
            href="/dashboard/123456789012345678/tickets"
            className={`${styles.btn} ${styles.btnSecondary}`} 
            style={{ width: '100%', justifyContent: 'center', padding: '12px' }}
          >
            <span>Enter Sandbox Demo Preview</span>
            <ChevronRight size={16} />
          </Link>
        </div>

        <div style={{ marginTop: 'var(--space-6)', fontSize: '11px', color: 'var(--color-text-muted)' }}>
          Secure server-side session management. Shrimpy never stores your Discord credentials in plaintext.
        </div>
      </div>
    </div>
  );
}

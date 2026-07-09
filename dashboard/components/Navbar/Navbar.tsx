// dashboard/components/Navbar.tsx
"use client";

import Link from "next/link";
import { Sun, Moon, Sparkles, LayoutDashboard } from "lucide-react";
import styles from "@/app/page.module.css";
import { Theme } from "@/lib/theme";
import { DiscordUser } from "@/lib/api";

interface NavbarProps {
  theme: Theme;
  toggleTheme: () => void;
  // undefined = auth state still resolving; null = signed out; object = signed in.
  user?: DiscordUser | null;
}

export default function Navbar({ theme, toggleTheme, user }: NavbarProps) {
  const loginUrl = `${process.env.NEXT_PUBLIC_SHRIMPY_API_URL || "http://localhost:8080"}/api/v1/auth/login`;

  return (
    <header className={`${styles.header} glass-effect`}>
      <div className={styles.headerInner}>
        <div className={styles.brand}>
          <span className={styles.brandIcon}>🦐</span>
          <span>Shrimpy</span>
        </div>

        <nav className={styles.nav}>
          <a href="#features" className={styles.navLink}>Features</a>
          <a href="#demo" className={styles.navLink}>Interactive Demo</a>
          <a href="#about" className={styles.navLink}>About</a>
          <div className={styles.navActions}>
            <button
              onClick={toggleTheme}
              className={styles.themeBtn}
              title={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} Mode`}
              aria-label="Toggle Theme"
            >
              {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
            </button>
            {user ? (
              <Link
                href="/servers"
                className={`${styles.btn} ${styles.discordBtn}`}
                title={`Signed in as ${user.globalName || user.username}`}
              >
                {user.avatar ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={user.avatar}
                    alt=""
                    width={20}
                    height={20}
                    style={{ borderRadius: "50%" }}
                  />
                ) : (
                  <LayoutDashboard size={16} />
                )}
                <span>Open Dashboard</span>
              </Link>
            ) : (
              <a
                href={loginUrl}
                className={`${styles.btn} ${styles.discordBtn}`}
                // Keep the button non-interactive while the session check is in flight
                // so a valid session isn't clobbered by an accidental re-login click.
                aria-disabled={user === undefined}
                style={user === undefined ? { opacity: 0.6, pointerEvents: "none" } : undefined}
              >
                <Sparkles size={16} />
                <span>Login with Discord</span>
              </a>
            )}
          </div>
        </nav>
      </div>
    </header>
  );
}

// dashboard/components/Navbar.tsx
"use client";

import { Sun, Moon, Sparkles } from "lucide-react";
import styles from "@/app/page.module.css";
import { Theme } from "@/lib/theme";

interface NavbarProps {
  theme: Theme;
  toggleTheme: () => void;
}

export default function Navbar({ theme, toggleTheme }: NavbarProps) {
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
            <a 
              href={`${process.env.NEXT_PUBLIC_SHRIMPY_API_URL || "http://localhost:8080"}/api/v1/auth/login`}
              className={`${styles.btn} ${styles.discordBtn}`}
            >
              <Sparkles size={16} />
              <span>Login with Discord</span>
            </a>
          </div>
        </nav>
      </div>
    </header>
  );
}

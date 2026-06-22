// dashboard/components/Hero.tsx
"use client";

import { ChevronRight } from "lucide-react";
import styles from "@/app/page.module.css";

export default function Hero() {
  return (
    <section className={`${styles.hero} hero-gradient`}>
      <div className={styles.heroInner}>
        <div className={styles.heroContent}>
          <div className={styles.badge}>
            <span>🤖 Multi-Bot Configurable Dashboard</span>
          </div>
          <h1 className={styles.title}>
            Manage Your Guilds with <span className="gradient-text">Shrimpy</span>
          </h1>
          <p className={styles.subtitle}>
            An ultra-modern, high-performance Discord companion offering dynamic support ticket panels, customizable greetings, and responsive reaction roles.
          </p>
          <div className={styles.actions}>
            <button className={`${styles.btn} ${styles.btnPrimary}`}>
              <span>Add to Discord</span>
              <ChevronRight size={16} />
            </button>
            <a href="#demo" className={`${styles.btn} ${styles.btnSecondary}`}>
              <span>Try Interactive Demo</span>
            </a>
          </div>

          <div className={styles.statsRow}>
            <div className={styles.statItem}>
              <span className={styles.statVal}>99.9%</span>
              <span className={styles.statLbl}>Uptime</span>
            </div>
            <div className={styles.statItem}>
              <span className={styles.statVal}>&lt; 50ms</span>
              <span className={styles.statLbl}>Response Time</span>
            </div>
            <div className={styles.statItem}>
              <span className={styles.statVal}>100k+</span>
              <span className={styles.statLbl}>Members Assisted</span>
            </div>
          </div>
        </div>

        <div className={styles.heroVisual}>
          <div className={styles.glowBlob}></div>
          <div className={styles.welcomeCard}>
            <div className={styles.cardBlob}></div>
            <div className={styles.welcomeUserAvatar}>🦐</div>
            <h3 className={styles.welcomeTitle}>Welcome, New Member!</h3>
            <p className={styles.welcomeDesc}>Say hi in #general and get started!</p>
            <div className={styles.welcomeMeta}>
              <span>ID: #89139</span>
              <span>•</span>
              <span>Member count: 1,248</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

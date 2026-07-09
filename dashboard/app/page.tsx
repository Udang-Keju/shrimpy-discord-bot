// dashboard/app/page.tsx
"use client";

import { useEffect, useState } from "react";
import styles from "./page.module.css";
import { getSavedTheme, applyTheme, Theme } from "../lib/theme";
import { ShrimpyAPI, DiscordUser } from "@/lib/api";
import Navbar from "@/components/Navbar";
import Hero from "@/components/Hero";
import Features from "@/components/Features";
import InteractiveDemo from "@/components/InteractiveDemo";
import Footer from "@/components/Footer";

export default function Home() {
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");
  const [activeTab, setActiveTab] = useState<"tickets" | "welcome" | "roles">("tickets");
  // undefined = still checking the existing session; null = signed out; object = signed in.
  const [user, setUser] = useState<DiscordUser | null | undefined>(undefined);

  // Handle Theme switching
  useEffect(() => {
    const t = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);
    return () => clearTimeout(t);
  }, []);

  // Reuse the existing 7-day session cookie: if /auth/me succeeds the user is
  // already logged in, so the navbar can offer "Open Dashboard" instead of a
  // fresh Discord OAuth round-trip. A 401 (signed out) is expected, not an error.
  useEffect(() => {
    let cancelled = false;
    ShrimpyAPI.getCurrentUser()
      .then((u) => { if (!cancelled) setUser(u); })
      .catch(() => { if (!cancelled) setUser(null); });
    return () => { cancelled = true; };
  }, []);

  const toggleTheme = () => {
    const nextTheme = theme === "dark" ? "light" : "dark";
    setTheme(nextTheme);
    applyTheme(nextTheme);
  };

  if (!mounted) {
    return null;
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.gridOverlay}></div>

      {/* NAVBAR */}
      <Navbar theme={theme} toggleTheme={toggleTheme} user={user} />

      {/* HERO SECTION */}
      <Hero />

      {/* FEATURES SECTION */}
      <Features setActiveTab={setActiveTab} />

      {/* INTERACTIVE DEMO */}
      <InteractiveDemo activeTab={activeTab} setActiveTab={setActiveTab} />

      {/* FOOTER SECTION */}
      <Footer />
    </div>
  );
}

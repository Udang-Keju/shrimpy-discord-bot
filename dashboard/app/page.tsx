// dashboard/app/page.tsx
"use client";

import { useEffect, useState } from "react";
import styles from "./page.module.css";
import { getSavedTheme, applyTheme, Theme } from "../lib/theme";
import Navbar from "@/components/Navbar";
import Hero from "@/components/Hero";
import Features from "@/components/Features";
import InteractiveDemo from "@/components/InteractiveDemo";
import Footer from "@/components/Footer";

export default function Home() {
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");
  const [activeTab, setActiveTab] = useState<"tickets" | "welcome" | "roles">("tickets");

  // Handle Theme switching
  useEffect(() => {
    const t = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);
    return () => clearTimeout(t);
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
      <Navbar theme={theme} toggleTheme={toggleTheme} />

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

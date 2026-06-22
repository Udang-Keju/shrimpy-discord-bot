// dashboard/components/Features.tsx
"use client";

import { Ticket, UserPlus, Tags, Settings } from "lucide-react";
import styles from "@/app/page.module.css";

interface FeaturesProps {
  setActiveTab?: (tab: "tickets" | "welcome" | "roles") => void;
}

export default function Features({ setActiveTab }: FeaturesProps) {
  const handleCardClick = (tab: "tickets" | "welcome" | "roles") => {
    if (setActiveTab) {
      setActiveTab(tab);
      document.getElementById("demo")?.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <section id="features" className={styles.features}>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Everything You Need for Community Management</h2>
        <p className={styles.sectionDesc}>
          Ditch complex setup commands. Configure support pipelines, interactive role assignment panels, and custom announcements directly in our web-dashboard.
        </p>
      </div>

      <div className={styles.featuresGrid}>
        <div className={styles.card} onClick={() => handleCardClick("tickets")}>
          <div className={styles.iconWrapper}>
            <Ticket size={24} />
          </div>
          <h3 className={styles.cardTitle}>Support Ticket Panels</h3>
          <p className={styles.cardDesc}>
            Create rich, button-based ticket builders. When members click to get help, Shrimpy spawns private, temporary threads with custom greeting overlays.
          </p>
        </div>

        <div className={styles.card} onClick={() => handleCardClick("welcome")}>
          <div className={styles.iconWrapper}>
            <UserPlus size={24} />
          </div>
          <h3 className={styles.cardTitle}>Custom Welcome Messages</h3>
          <p className={styles.cardDesc}>
            Set up personalized direct messages and public greetings to welcome new users, automatically applying roles on server join.
          </p>
        </div>

        <div className={styles.card} onClick={() => handleCardClick("roles")}>
          <div className={styles.iconWrapper}>
            <Tags size={24} />
          </div>
          <h3 className={styles.cardTitle}>Interactive Reaction Roles</h3>
          <p className={styles.cardDesc}>
            Enable self-assignable roles. Users simply click message reaction buttons to toggle server tags, avoiding manual staff moderation.
          </p>
        </div>

        <div className={styles.card}>
          <div className={styles.iconWrapper}>
            <Settings size={24} />
          </div>
          <h3 className={styles.cardTitle}>Multi-Bot Registration</h3>
          <p className={styles.cardDesc}>
            Hosting multiple configurations? Manage all your custom Discord applications under a single GORM relational backend in pgxpool.
          </p>
        </div>
      </div>
    </section>
  );
}

// dashboard/app/dashboard/[guildId]/page.tsx
"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { getNavigationGroups } from "./navConfig";

export default function OverviewPage() {
  const params = useParams();
  const guildId = params?.guildId as string;

  const groups = getNavigationGroups(guildId).filter(g => g.label !== "Settings");

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Overview</h2>
        <p className={styles.sectionDesc}>Quick access to everything Shrimpy can do for this server.</p>
      </div>

      {groups.map(group => (
        <div key={group.label}>
          <h3 className={styles.featureGroupLabel}>{group.label}</h3>
          <div className={styles.featureGrid}>
            {group.items.map(item => {
              const Icon = item.icon;
              return (
                <Link key={item.href} href={item.href} className={styles.featureCard}>
                  <div className={styles.featureIconWrap}>
                    <Icon size={20} />
                  </div>
                  <div className={styles.featureCardTitle}>{item.name}</div>
                  <div className={styles.featureCardDesc}>{item.description}</div>
                </Link>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}

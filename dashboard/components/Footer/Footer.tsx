// dashboard/components/Footer.tsx
"use client";

import Link from "next/link";
import { Heart } from "lucide-react";
import styles from "@/app/page.module.css";

export default function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.footerInner}>
        <div className={styles.footerBrand}>
          <div className={styles.brand}>
            <span>🦐</span>
            <span>Shrimpy</span>
          </div>
          <p className={styles.footerDesc}>
            A highly performant Discord utility system engineered in Go with Next.js dashboards.
          </p>
        </div>

        <div className={styles.footerCol}>
          <span className={styles.footerTitle}>Resources</span>
          <ul className={styles.footerLinks}>
            <li><a href="#" className={styles.footerLink}>Documentation</a></li>
            <li><a href="#" className={styles.footerLink}>Commands list</a></li>
            <li><a href="#" className={styles.footerLink}>API Guides</a></li>
          </ul>
        </div>

        <div className={styles.footerCol}>
          <span className={styles.footerTitle}>Support</span>
          <ul className={styles.footerLinks}>
            <li><a href="#" className={styles.footerLink}>Join Discord</a></li>
            <li><a href="#" className={styles.footerLink}>Report Bug</a></li>
            <li><a href="#" className={styles.footerLink}>Status Page</a></li>
          </ul>
        </div>

        <div className={styles.footerCol}>
          <span className={styles.footerTitle}>Legal</span>
          <ul className={styles.footerLinks}>
            <li><Link href="/privacy" className={styles.footerLink}>Privacy Policy</Link></li>
            <li><Link href="/terms" className={styles.footerLink}>Terms of Service</Link></li>
          </ul>
        </div>
      </div>

      <div className={styles.footerBottom}>
        <span>&copy; {new Date().getFullYear()} Shrimpy Bot. All rights reserved.</span>
        <span style={{display: 'flex', alignItems: 'center', gap: '4px'}}>
          Made with <Heart size={12} style={{color: 'var(--color-primary)'}} /> by the Engineering Team
        </span>
      </div>
    </footer>
  );
}

// dashboard/app/privacy/page.tsx
"use client";

import { useEffect, useState } from "react";
import page from "@/app/page.module.css";
import legal from "@/app/legal.module.css";
import { getSavedTheme, applyTheme, Theme } from "@/lib/theme";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

const CONTACT_EMAIL = "vns05081208@gmail.com";

export default function PrivacyPolicyPage() {
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");

  useEffect(() => {
    const t = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);
    document.title = "Privacy Policy | Shrimpy";
    return () => clearTimeout(t);
  }, []);

  const toggleTheme = () => {
    const next = theme === "dark" ? "light" : "dark";
    setTheme(next);
    applyTheme(next);
  };

  if (!mounted) return null;

  return (
    <div className={page.wrapper}>
      <div className={page.gridOverlay}></div>

      <Navbar theme={theme} toggleTheme={toggleTheme} />

      <main className={legal.page}>
        <div className={legal.container}>
          <h1 className={legal.title}>Privacy Policy</h1>
          <p className={legal.updated}>Last updated: July 1, 2026</p>

          <p className={legal.intro}>
            This Privacy Policy explains how Shrimpy (&quot;Shrimpy&quot;, &quot;we&quot;,
            &quot;us&quot;, or &quot;our&quot;) collects, uses, and safeguards information when
            you add our Discord bot to a server or use our web dashboard. By using Shrimpy, you
            agree to the practices described below.
          </p>

          <section className={legal.section}>
            <h2 className={legal.heading}>1. Information We Collect</h2>
            <p className={legal.paragraph}>
              Shrimpy only collects the data needed to provide its server-management and
              help-desk features:
            </p>
            <ul className={legal.list}>
              <li>
                <strong>Account &amp; authentication data.</strong> When you log in to the
                dashboard with Discord OAuth2, we store your Discord user ID, username,
                discriminator, and avatar hash. Discord OAuth access and refresh tokens are
                stored <strong>encrypted</strong> and used only to identify you and read the
                servers you manage.
              </li>
              <li>
                <strong>Server configuration.</strong> Settings you configure per server —
                command prefix, bot nickname, log channel, welcome messages, auto-roles, staff
                roles, reaction-role mappings, and ticket panels/categories.
              </li>
              <li>
                <strong>Ticket data &amp; transcripts.</strong> For the ticketing system we
                store ticket messages, their content, author IDs and usernames, attachments,
                and lifecycle metadata (who opened, claimed, or closed a ticket and when).
              </li>
              <li>
                <strong>Discord identifiers.</strong> Numeric Discord IDs (snowflakes) for
                guilds, channels, roles, messages, and users that are referenced by the features
                above.
              </li>
            </ul>
            <p className={legal.paragraph}>
              We do not collect message content from your server beyond the ticket
              conversations Shrimpy is explicitly used to manage.
            </p>
            <p className={legal.paragraph}>
              As Shrimpy&apos;s features evolve, we may collect additional information
              necessary to provide new functionality. Any such changes will be reflected in
              this policy and in its &quot;Last updated&quot; date.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>2. How We Use Information</h2>
            <p className={legal.paragraph}>We use the information we collect to:</p>
            <ul className={legal.list}>
              <li>Operate, maintain, and provide Shrimpy&apos;s features.</li>
              <li>Authenticate dashboard users and verify their permissions on a server.</li>
              <li>Generate and store ticket transcripts for server staff.</li>
              <li>Apply the configuration you set for each server.</li>
            </ul>
            <p className={legal.paragraph}>
              We do not use your data for advertising and we do not sell it.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>3. Data Storage &amp; Security</h2>
            <p className={legal.paragraph}>
              Data is stored in a PostgreSQL database. Sensitive credentials — including bot
              tokens and Discord OAuth tokens — are encrypted at rest using AES-256-GCM before
              being written to the database. We take reasonable measures to protect your
              information, though no method of transmission or storage is completely secure.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>4. Data Sharing</h2>
            <p className={legal.paragraph}>
              We do not sell, rent, or trade your information. Data is shared only with the
              Discord API as required to deliver Shrimpy&apos;s functionality (for example, to
              post messages, create ticket channels, or assign roles), and with the
              infrastructure providers that host our database and services. We may disclose
              information if required by law.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>5. Data Retention &amp; Deletion</h2>
            <p className={legal.paragraph}>
              We retain data for as long as Shrimpy is active on your server. Removing the bot
              from a server stops further data collection for that server. To request deletion
              of your stored data — including server configuration, tickets, transcripts, or
              your account record — contact us at the email below and we will process the
              request within a reasonable period.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>6. Children&apos;s Privacy</h2>
            <p className={legal.paragraph}>
              Shrimpy is intended for use on Discord, which requires users to be at least 13
              years old (or the minimum age of digital consent in their country). We do not
              knowingly collect data from anyone who does not meet Discord&apos;s age
              requirements.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>7. Changes to This Policy</h2>
            <p className={legal.paragraph}>
              We may update this Privacy Policy from time to time. Material changes will be
              reflected by updating the &quot;Last updated&quot; date at the top of this page.
              Continued use of Shrimpy after changes take effect constitutes acceptance of the
              revised policy.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>8. Contact</h2>
            <p className={legal.paragraph}>
              If you have questions about this Privacy Policy or wish to make a data request,
              contact us at{" "}
              <a className={legal.link} href={`mailto:${CONTACT_EMAIL}`}>
                {CONTACT_EMAIL}
              </a>
              .
            </p>
          </section>
        </div>
      </main>

      <Footer />
    </div>
  );
}

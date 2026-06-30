// dashboard/app/terms/page.tsx
"use client";

import { useEffect, useState } from "react";
import page from "@/app/page.module.css";
import legal from "@/app/legal.module.css";
import { getSavedTheme, applyTheme, Theme } from "@/lib/theme";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

const CONTACT_EMAIL = "vns05081208@gmail.com";

export default function TermsOfServicePage() {
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");

  useEffect(() => {
    const t = setTimeout(() => {
      setMounted(true);
      setTheme(getSavedTheme());
    }, 0);
    document.title = "Terms of Service | Shrimpy";
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
          <h1 className={legal.title}>Terms of Service</h1>
          <p className={legal.updated}>Last updated: July 1, 2026</p>

          <p className={legal.intro}>
            These Terms of Service (&quot;Terms&quot;) govern your use of Shrimpy
            (&quot;Shrimpy&quot;, &quot;we&quot;, &quot;us&quot;, or &quot;our&quot;), including
            the Discord bot and the web dashboard. By adding Shrimpy to a Discord server or using
            the dashboard, you agree to these Terms.
          </p>

          <section className={legal.section}>
            <h2 className={legal.heading}>1. Acceptance of Terms</h2>
            <p className={legal.paragraph}>
              By inviting Shrimpy to a server, configuring it, or otherwise using the service,
              you acknowledge that you have read, understood, and agree to be bound by these
              Terms. If you do not agree, do not use Shrimpy.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>2. Description of Service</h2>
            <p className={legal.paragraph}>
              Shrimpy is a Discord server management and help-desk system that provides
              ticketing, welcome messages, auto-roles, reaction roles, and related
              configuration tools through a Discord bot and a web dashboard. We continually
              develop Shrimpy, and may add, change, or remove features over time.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>3. Eligibility</h2>
            <p className={legal.paragraph}>
              You must comply with Discord&apos;s{" "}
              <a
                className={legal.link}
                href="https://discord.com/terms"
                target="_blank"
                rel="noopener noreferrer"
              >
                Terms of Service
              </a>{" "}
              and{" "}
              <a
                className={legal.link}
                href="https://discord.com/guidelines"
                target="_blank"
                rel="noopener noreferrer"
              >
                Community Guidelines
              </a>{" "}
              to use Shrimpy. You must have the appropriate permissions on a Discord server to
              add the bot and change its configuration.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>4. Acceptable Use</h2>
            <p className={legal.paragraph}>You agree not to:</p>
            <ul className={legal.list}>
              <li>Use Shrimpy for any unlawful purpose or in violation of Discord&apos;s policies.</li>
              <li>Attempt to disrupt, abuse, overload, or reverse-engineer the service.</li>
              <li>Use Shrimpy to harass other users or distribute prohibited content.</li>
              <li>Attempt to gain unauthorized access to other servers&apos; data or to our systems.</li>
            </ul>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>5. Service Availability</h2>
            <p className={legal.paragraph}>
              Shrimpy is provided on an &quot;as is&quot; and &quot;as available&quot; basis. We
              do not guarantee that the service will be uninterrupted, error-free, or available
              at any particular time. We may modify, suspend, or discontinue any part of the
              service at any time without notice.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>6. Limitation of Liability</h2>
            <p className={legal.paragraph}>
              To the maximum extent permitted by law, Shrimpy and its operators shall not be
              liable for any indirect, incidental, special, consequential, or punitive damages,
              or any loss of data, profits, or revenue, arising out of or related to your use of
              the service.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>7. Termination</h2>
            <p className={legal.paragraph}>
              You may stop using Shrimpy at any time by removing the bot from your server. We may
              suspend or terminate access to the service for any user or server that violates
              these Terms or Discord&apos;s policies.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>8. Changes to These Terms</h2>
            <p className={legal.paragraph}>
              We may update these Terms from time to time. Material changes will be reflected by
              updating the &quot;Last updated&quot; date at the top of this page. Continued use of
              Shrimpy after changes take effect constitutes acceptance of the revised Terms.
            </p>
          </section>

          <section className={legal.section}>
            <h2 className={legal.heading}>9. Contact</h2>
            <p className={legal.paragraph}>
              If you have questions about these Terms, contact us at{" "}
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

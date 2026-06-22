import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import localFont from "next/font/local";
import "./globals.css";

const ggSans = localFont({
  src: [
    {
      path: "../public/fonts/ggsans-normal-400.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../public/fonts/ggsans-normal-500.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../public/fonts/ggsans-normal-600.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../public/fonts/ggsans-normal-700.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../public/fonts/ggsans-normal-800.woff2",
      weight: "800",
      style: "normal",
    },
  ],
  variable: "--font-ggsans",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  weight: ["400", "500"],
});

export const metadata: Metadata = {
  title: "Shrimpy 🦐 — The Ultimate Discord Bot Dashboard",
  description: "Manage your Discord server tickets, welcome messages, auto-roles, and reaction roles easily with Shrimpy's premium dashboard.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${ggSans.variable} ${jetbrainsMono.variable}`}>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                var theme = localStorage.getItem('Shrimpy-theme') ||
                  (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
                document.documentElement.setAttribute('data-theme', theme);
              })();
            `,
          }}
        />
      </head>
      <body>
        {children}
      </body>
    </html>
  );
}

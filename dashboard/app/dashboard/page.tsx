"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { ShrimpyAPI } from "@/lib/api";

export default function DashboardIndex() {
  const router = useRouter();

  useEffect(() => {
    async function redirectUser() {
      try {
        const guildList = await ShrimpyAPI.listGuilds();
        if (guildList && guildList.length > 0) {
          router.replace(`/dashboard/${guildList[0].id}/tickets`);
        } else {
          // If the user has no guilds, default to the sandbox demo guild
          router.replace(`/dashboard/123456789012345678/tickets`);
        }
      } catch (err) {
        console.error("Failed to load user guilds, falling back to demo", err);
        router.replace(`/dashboard/123456789012345678/tickets`);
      }
    }
    redirectUser();
  }, [router]);

  return (
    <div style={{ display: 'flex', height: '100vh', width: '100vw', justifyContent: 'center', alignItems: 'center', backgroundColor: '#0d0e12', color: '#fff' }}>
      <div style={{ textAlign: 'center' }}>
        <div style={{ fontSize: '32px', marginBottom: '16px' }}>🦐</div>
        <p style={{ color: '#8f9cae', fontSize: '14px' }}>Loading your dashboard...</p>
      </div>
    </div>
  );
}

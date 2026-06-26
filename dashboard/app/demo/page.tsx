// dashboard/app/demo/page.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { ShrimpyAPI } from "@/lib/api";

const DEMO_GUILD_ID = "123456789012345678";

export default function DemoEntryPage() {
  const router = useRouter();

  useEffect(() => {
    ShrimpyAPI.enterDemoMode();
    router.replace(`/dashboard/${DEMO_GUILD_ID}/tickets`);
  }, [router]);

  return (
    <div style={{ display: "flex", height: "100vh", width: "100vw", justifyContent: "center", alignItems: "center", backgroundColor: "#06070a", color: "#fff", fontFamily: "'Outfit', sans-serif" }}>
      <div style={{ textAlign: "center" }}>
        <div style={{ fontSize: "36px", marginBottom: "16px" }}>🦐</div>
        <p style={{ color: "#717f96", fontSize: "14px", letterSpacing: "0.5px" }}>Entering Shrimpy Sandbox Demo...</p>
      </div>
    </div>
  );
}

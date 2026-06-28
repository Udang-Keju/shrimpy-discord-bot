// dashboard/components/DemoBanner.tsx
"use client";

import { useRouter } from "next/navigation";
import { FlaskConical } from "lucide-react";
import { ShrimpyAPI } from "@/lib/api";

export default function DemoBanner() {
  const router = useRouter();

  const handleExit = () => {
    ShrimpyAPI.exitDemoMode();
    router.push("/login");
  };

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: "10px",
        padding: "8px 16px",
        backgroundColor: "rgba(99, 102, 241, 0.12)",
        borderBottom: "1px solid rgba(99, 102, 241, 0.3)",
        fontSize: "13px",
        fontWeight: 600,
        color: "#818cf8",
      }}
    >
      <FlaskConical size={14} />
      <span>You&apos;re viewing the Sandbox Demo. Changes here aren&apos;t saved to a real Discord server.</span>
      <button
        onClick={handleExit}
        style={{
          background: "none",
          border: "none",
          color: "#818cf8",
          textDecoration: "underline",
          cursor: "pointer",
          fontSize: "13px",
          fontWeight: 700,
          padding: 0,
        }}
      >
        Exit Demo
      </button>
    </div>
  );
}

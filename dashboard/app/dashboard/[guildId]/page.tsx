// dashboard/app/dashboard/[guildId]/page.tsx
import { redirect } from "next/navigation";
import { getDefaultDashboardHref } from "./navConfig";

export default async function GuildIndexPage({
  params,
}: {
  params: Promise<{ guildId: string }>;
}) {
  const { guildId } = await params;
  redirect(getDefaultDashboardHref(guildId));
}

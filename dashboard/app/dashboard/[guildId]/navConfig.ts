// dashboard/app/dashboard/[guildId]/navConfig.ts
import {
  Ticket,
  UserPlus,
  Tags,
  Settings,
  Layers,
  Languages,
  LucideIcon
} from "lucide-react";

export interface NavItem {
  name: string;
  href: string;
  icon: LucideIcon;
  description: string;
}

export interface NavGroup {
  label: string;
  items: NavItem[];
}

// Group order is Settings, Server Management, Tickets (see docs/v1/USER_JOURNEY.md §5.3 note).
export function getNavigationGroups(guildId: string): NavGroup[] {
  return [
    {
      label: "Settings",
      items: [
        {
          name: "General Settings",
          href: `/dashboard/${guildId}/settings`,
          icon: Settings,
          description: "Manage bot naming, prefix, log channel, and dashboard access roles."
        },
      ],
    },
    {
      label: "Server Management",
      items: [
        {
          name: "Welcome Greetings",
          href: `/dashboard/${guildId}/welcome`,
          icon: UserPlus,
          description: "Send automatic welcome banners and DMs when new members join."
        },
        {
          name: "Reaction Roles",
          href: `/dashboard/${guildId}/roles`,
          icon: Tags,
          description: "Let members self-assign roles by reacting to a message."
        },
        {
          name: "Message Translation",
          href: `/dashboard/${guildId}/translation`,
          icon: Languages,
          description: "Auto-translate member messages by channel or emoji reaction."
        },
      ],
    },
    {
      label: "Tickets",
      items: [
        {
          name: "Tickets Manager",
          href: `/dashboard/${guildId}/tickets`,
          icon: Ticket,
          description: "Review active support threads, claim conversations, and save transcripts."
        },
        {
          name: "Ticket Panels",
          href: `/dashboard/${guildId}/panels`,
          icon: Layers,
          description: "Build interactive embeds that let members open tickets with one click."
        },
      ],
    },
  ];
}

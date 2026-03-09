"use client";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

import {
  IconBell,
  IconBook,
  IconCalendarEvent,
  IconCheckbox,
  IconClipboardCheck,
  IconClipboardList,
  IconClock,
  IconLayoutDashboard,
  IconPackage,
  IconSpeakerphone,
  IconUser,
  IconUsers,
  IconUsersGroup,
  IconActivity,
  IconScale,
} from "@tabler/icons-react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { Role } from "@/app/(auth)/components/login-form";
import {
  Bell,
  BookOpen,
  ClipboardCheck,
  CreditCard,
  LayoutDashboard,
  MessageSquare,
  Settings,
} from "lucide-react";
import { NavUser } from "../../components/nav-user";

// Grouped navigation items by category
const navGroups = [
  //   {
  //     label: "Dashboard",
  //     items: [
  //       {
  //         title: "Overview",
  //         url: "/dashboard",
  //         icon: IconLayoutDashboard,
  //         roles: [Role.ADMIN, Role.MANAGER, Role.TUTOR], // instructor => overview
  //       },
  //       {
  //         title: "My Workspace",
  //         url: "/dashboard/my-workspace",
  //         icon: IconClipboardCheck,
  //         roles: [Role.ADMIN, Role.MANAGER, Role.TUTOR], // instructor => my tasks
  //       },
  //     ],
  //   },

  {
    label: "Student",
    items: [
      {
        url: "/dashboard",
        title: "Dashboard",
        icon: LayoutDashboard,
        roles: [Role.ADMIN, Role.MANAGER, Role.TUTOR],
      },
      { url: "/dashboard/my-courses", title: "My Courses", icon: BookOpen },
      {
        url: "/dashboard/peer-reviews",
        title: "Peer Reviews",
        icon: ClipboardCheck,
      },
      {
        url: "/dashboard/subscription",
        title: "Subscription",
        icon: CreditCard,
      },
      { url: "/dashboard/messages", title: "Messages", icon: MessageSquare },
      { url: "/dashboard/notifications", title: "Notifications", icon: Bell },
      { url: "/dashboard/settings", title: "Settings", icon: Settings },
    ],
  },
];

interface AppSidebarProps {
  user: {
    name: string;
    email: string;
    role: any;
    avatar: string;
  } | null;
}

export function DashboardSidebar({ user, ...props }: AppSidebarProps) {
  const pathname = usePathname();

  return (
    <Sidebar {...props}>
      <SidebarHeader className="border-b border-sidebar-border p-3">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Image src={`/logo.png`} alt="Logo" width={32} height={32} />
          </div>
          <span className="text-base font-semibold">Next Level Management</span>
        </div>
      </SidebarHeader>

      <SidebarContent>
        {navGroups.map((group) => {
          // Filter items based on user role
          const filteredItems = group.items.filter(
            (item) => user,
            // (item) => user && item.roles.includes(user.role),
          );

          // Don't render empty groups
          if (filteredItems.length === 0) return null;

          return (
            <SidebarGroup key={group.label}>
              <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {filteredItems.map((item) => {
                    const isActive =
                      pathname === item.url ||
                      (item.url !== "/dashboard" &&
                        pathname.startsWith(item.url));
                    return (
                      <SidebarMenuItem key={item.url}>
                        <SidebarMenuButton
                          isActive={isActive}
                          render={<Link href={item.url} prefetch={true} />}
                        >
                          <item.icon className="h-4 w-4" />
                          <span>{item.title}</span>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          );
        })}
      </SidebarContent>

      {user && <NavUser user={user} />}
    </Sidebar>
  );
}

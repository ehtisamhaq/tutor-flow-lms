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

// Navigation items for different roles
const navGroups = [
  {
    label: "Admin",
    roles: ["admin", "manager"],
    items: [
      { title: "Dashboard", url: "/admin", icon: LayoutDashboard },
      { title: "Users", url: "/admin/users", icon: IconUsers },
      { title: "Courses", url: "/admin/courses", icon: BookOpen },
      { title: "Bundles", url: "/admin/bundles", icon: IconPackage },
      {
        title: "Subscriptions",
        url: "/admin/subscription-plans",
        icon: CreditCard,
      },
      { title: "Refunds", url: "/admin/refunds", icon: IconScale },
      { title: "Revenue", url: "/admin/revenue", icon: CreditCard },
      { title: "Analytics", url: "/admin/analytics", icon: IconActivity },
      { title: "Settings", url: "/admin/settings", icon: Settings },
    ],
  },
  {
    label: "Instructor",
    roles: ["tutor"],
    items: [
      { title: "Dashboard", url: "/tutor", icon: LayoutDashboard },
      { title: "My Courses", url: "/tutor/courses", icon: BookOpen },
      { title: "Earnings", url: "/tutor/earnings", icon: CreditCard },
      { title: "Reviews", url: "/tutor/reviews", icon: IconClipboardCheck },
      { title: "Messages", url: "/tutor/messages", icon: MessageSquare },
      { title: "Settings", url: "/tutor/settings", icon: Settings },
    ],
  },
  {
    label: "Student",
    roles: ["student"],
    items: [
      { title: "Dashboard", url: "/dashboard", icon: LayoutDashboard },
      { title: "My Courses", url: "/dashboard/my-courses", icon: BookOpen },
      {
        title: "Peer Reviews",
        url: "/dashboard/peer-reviews",
        icon: ClipboardCheck,
      },
      {
        title: "Subscription",
        url: "/dashboard/subscription",
        icon: CreditCard,
      },
      { title: "Messages", url: "/dashboard/messages", icon: MessageSquare },
      { title: "Notifications", url: "/dashboard/notifications", icon: Bell },
      { title: "Settings", url: "/dashboard/settings", icon: Settings },
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
          // Filter groups based on user role
          const isAllowed = user && group.roles.includes(user.role);

          // Don't render groups the user doesn't have access to
          if (!isAllowed) return null;

          const filteredItems = group.items;

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

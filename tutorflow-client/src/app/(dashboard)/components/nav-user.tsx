"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  IconCreditCard,
  IconDotsVertical,
  IconLoader2,
  IconLogout,
  IconNotification,
  IconUserCircle,
} from "@tabler/icons-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { logout } from "@/app/(auth)/lib/actions";

/** Derive 1-2 letter initials from a full name */
function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}

/** Resolve role-aware page paths */
function getRolePaths(role: string) {
  if (role === "admin" || role === "manager") {
    return {
      account: "/admin/settings",
      billing: "/admin/subscription-plans",
      notifications: "/admin/settings",
    };
  }
  if (role === "tutor") {
    return {
      account: "/tutor/settings",
      billing: "/tutor/earnings",
      notifications: "/tutor/settings",
    };
  }
  // student (default)
  return {
    account: "/dashboard/settings",
    billing: "/dashboard/subscription",
    notifications: "/dashboard/notifications",
  };
}

export function NavUser({
  user,
}: {
  user: {
    name: string;
    email: string;
    avatar: string;
    role?: string;
  };
}) {
  const { isMobile } = useSidebar();
  const router = useRouter();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const initials = getInitials(user.name);
  const paths = getRolePaths(user.role ?? "student");

  const handleLogout = async () => {
    setIsLoggingOut(true);
    await logout();
  };

  const handleNavigate = (path: string) => {
    router.push(path);
  };

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              />
            }
          >
            <Avatar className="h-8 w-8 rounded-lg grayscale">
              <AvatarImage src={user.avatar} alt={user.name} />
              <AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium">{user.name}</span>
              <span className="truncate text-xs text-muted-foreground">
                {user.email}
              </span>
            </div>
            <IconDotsVertical className="ml-auto size-4" />
          </DropdownMenuTrigger>

          <DropdownMenuContent
            className="w-(--base-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? "top" : "right"}
            align="end"
            sideOffset={4}
          >
            {/* User info header */}
            <DropdownMenuGroup>
              <DropdownMenuLabel className="p-0 font-normal">
                <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarImage src={user.avatar} alt={user.name} />
                    <AvatarFallback className="rounded-lg">
                      {initials}
                    </AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user.name}</span>
                    <span className="truncate text-xs text-muted-foreground capitalize">
                      {user.role ?? "student"}
                    </span>
                  </div>
                </div>
              </DropdownMenuLabel>
            </DropdownMenuGroup>

            <DropdownMenuSeparator />

            {/* Navigation links */}
            <DropdownMenuGroup>
              <DropdownMenuItem onClick={() => handleNavigate(paths.account)}>
                <IconUserCircle className="mr-2 size-4" />
                Account
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleNavigate(paths.billing)}>
                <IconCreditCard className="mr-2 size-4" />
                Billing
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => handleNavigate(paths.notifications)}
              >
                <IconNotification className="mr-2 size-4" />
                Notifications
              </DropdownMenuItem>
            </DropdownMenuGroup>

            <DropdownMenuSeparator />

            {/* Logout */}
            <DropdownMenuGroup>
              <DropdownMenuItem
                onClick={handleLogout}
                disabled={isLoggingOut}
                className="text-destructive focus:text-destructive"
              >
                {isLoggingOut ? (
                  <IconLoader2 className="mr-2 size-4 animate-spin" />
                ) : (
                  <IconLogout className="mr-2 size-4" />
                )}
                {isLoggingOut ? "Signing out…" : "Log out"}
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}

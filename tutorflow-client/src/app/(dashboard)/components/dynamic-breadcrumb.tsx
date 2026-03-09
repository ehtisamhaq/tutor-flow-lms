"use client";

import * as React from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

const routeConfig: Record<string, string> = {
  dashboard: "Dashboard",
  "my-workspace": "My Workspace",
  instructors: "Members",
  teams: "Teams",
  batches: "Batches",
  tasks: "Tasks",
  assignments: "Assignments",
  "extension-requests": "Extension Requests",
  meetings: "Meetings",
  schedule: "Work Schedules",
  materials: "Materials",
  notifications: "Notifications",
  announcements: "Announcements",
  settings: "Settings",
  config: "Configuration",
  analytics: "Analytics",
};

export function DynamicBreadcrumb() {
  const pathname = usePathname();

  // Split path and filter out empty segments
  const segments = pathname.split("/").filter(Boolean);

  // If we're at the root of dashboard or less, maybe show something default
  // but usually we are at least in /dashboard

  return (
    <Breadcrumb className="min-w-0">
      <BreadcrumbList className="flex-nowrap">
        {segments.map((segment, index) => {
          const isLast = index === segments.length - 1;
          const href = `/${segments.slice(0, index + 1).join("/")}`;
          const title =
            routeConfig[segment] ||
            segment.charAt(0).toUpperCase() +
              segment.slice(1).replace(/-/g, " ");

          return (
            <React.Fragment key={href}>
              <BreadcrumbItem
                className={!isLast ? "hidden md:block" : "min-w-0"}
              >
                {isLast ? (
                  <BreadcrumbPage className="truncate max-w-37.5 sm:max-w-none">
                    {title}
                  </BreadcrumbPage>
                ) : (
                  <BreadcrumbLink render={<Link href={href} />}>
                    {title}
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
              {!isLast && <BreadcrumbSeparator className="hidden md:block" />}
            </React.Fragment>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}

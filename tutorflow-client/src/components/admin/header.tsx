"use client";

import Link from "next/link";
import { Bell, Menu, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";

export function AdminHeader() {
  return (
    <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur">
      <div className="flex h-16 items-center justify-between px-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" className="lg:hidden">
            <Menu className="h-5 w-5" />
          </Button>
          <h1 className="text-lg font-semibold hidden sm:block">Admin Panel</h1>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" asChild>
            <Link href="/" target="_blank">
              <ExternalLink className="h-4 w-4 mr-2" />
              View Site
            </Link>
          </Button>
          <Button variant="ghost" size="icon">
            <Bell className="h-5 w-5" />
          </Button>
        </div>
      </div>
    </header>
  );
}

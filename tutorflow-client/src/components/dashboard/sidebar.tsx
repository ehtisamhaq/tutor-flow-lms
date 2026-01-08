"use client";

import { useAuthStore } from "@/store/auth-store";
import { Button } from "@/components/ui/button";
import { LogOut, User } from "lucide-react";
import { useRouter } from "next/navigation";

export function DashboardSidebar() {
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push("/login");
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3 px-3 py-2">
        <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
          <User className="h-5 w-5 text-primary" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">
            {user?.first_name} {user?.last_name}
          </p>
          <p className="text-xs text-muted-foreground truncate">
            {user?.email}
          </p>
        </div>
      </div>
      <Button
        variant="ghost"
        className="w-full justify-start gap-3 text-muted-foreground hover:text-foreground"
        onClick={handleLogout}
      >
        <LogOut className="h-5 w-5" />
        Sign Out
      </Button>
    </div>
  );
}

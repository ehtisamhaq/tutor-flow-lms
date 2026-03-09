import Link from "next/link";
import { redirect } from "next/navigation";
import { cookies } from "next/headers";
import {
  GraduationCap,
  LayoutDashboard,
  BookOpen,
  MessageSquare,
  Bell,
  Settings,
  User,
  LogOut,
  ShoppingCart,
  ClipboardCheck,
  CreditCard,
} from "lucide-react";
import { Button } from "@/components/ui/button";
// import { DashboardSidebar } from "@/components/dashboard/sidebar";
import { DashboardHeader } from "@/components/dashboard/header";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { AppSidebar } from "../components/app-sidebar";
import { SiteHeader } from "../components/site-header";
import { SectionCards } from "../components/section-cards";
import { ChartAreaInteractive } from "../components/chart-area-interactive";
import { DataTable } from "../components/data-table";
import { getSession } from "@/lib/session";
import { Separator } from "@/components/ui/separator";
import { DynamicBreadcrumb } from "../components/dynamic-breadcrumb";
import { ScrollArea } from "@/components/ui/scroll-area";
import { DashboardSidebar } from "./components/dashboard-sidebar";

// Server-side auth check
async function checkAuth() {
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken");
  if (!token) {
    redirect("/login");
  }
  return token.value;
}

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  await checkAuth();

  const session = await getSession();

  if (!session) {
    redirect("/login");
  }

  const user = {
    name: `${session.first_name} ${session.last_name}`,
    email: session.email,
    role: session.role,
    avatar: session.avatar_url || "",
  };

  return (
    <SidebarProvider>
      <DashboardSidebar user={user} />
      <SidebarInset className="flex h-screen flex-col overflow-hidden">
        <header className="flex h-12 sm:h-10 shrink-0 items-center justify-between gap-2 border-b px-2 sm:px-4">
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <SidebarTrigger className="-ml-1 shrink-0" />
            <Separator
              orientation="vertical"
              className="mr-2 hidden sm:block"
            />
            <div className="min-w-0 flex-1">
              <DynamicBreadcrumb />
            </div>
          </div>
          <div className="flex items-center gap-1 sm:gap-2 shrink-0">-</div>
        </header>

        <main className="flex-1 min-h-0 min-w-0 flex flex-col overflow-hidden">
          <ScrollArea className="h-full w-full">
            <div className="w-full overflow-x-hidden">
              {/* <FadeIn key="dashboard-content"> */}
              <div className="flex-1 space-y-4 p-2 md:p-4 lg:p-8">
                {children}
              </div>
              {/* </FadeIn> */}
            </div>
          </ScrollArea>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}

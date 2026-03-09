import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/session";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { DynamicBreadcrumb } from "../components/dynamic-breadcrumb";
import { ScrollArea } from "@/components/ui/scroll-area";
import { DashboardSidebar } from "../dashboard/components/dashboard-sidebar";

export default async function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await getSession();

  if (!session) {
    redirect("/login");
  }

  // Only allow admin and manager roles
  if (session.role !== "admin" && session.role !== "manager") {
    redirect("/dashboard?error=unauthorized");
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

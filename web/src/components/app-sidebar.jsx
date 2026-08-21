"use client"

import {
  Boxes,
  Globe2,
  LayoutDashboard,
  Link2,
  ListChecks,
  Radar,
  Server,
  TriangleAlert,
  UserRound,
  UsersRound,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

const data = {
  navMain: [
    { section: "Operations", title: "Overview", url: "/", icon: LayoutDashboard, status: "available" },
    { section: "Operations", title: "Hosts", url: "/hosts", icon: Server, status: "coming_soon" },
    { section: "Operations", title: "Services", url: "/services", icon: Boxes, status: "coming_soon" },
    { section: "Operations", title: "Rules", url: "/rules", icon: ListChecks, status: "coming_soon" },
    { section: "Operations", title: "Incidents", url: "/incidents", icon: TriangleAlert, status: "coming_soon" },
    { section: "Ownership", title: "Teams", url: "/teams", icon: UsersRound, status: "coming_soon" },
    { section: "Ownership", title: "Members", url: "/teams/members", icon: UserRound, status: "coming_soon" },
    { section: "Ownership", title: "Assignments", url: "/teams/assignments", icon: Link2, status: "coming_soon" },
    { section: "Status", title: "Status Page", url: "/status", icon: Globe2, status: "coming_soon" },
  ],
}

export function AppSidebar({ user, onLogout, ...props }) {
  return (
    <Sidebar collapsible="icon" variant="inset" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="/">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <Radar className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">Nodara</span>
                  <span className="truncate text-xs">Monitoring Console</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={{ name: user.username, email: user.email }} onLogout={onLogout} />
      </SidebarFooter>
    </Sidebar>
  );
}

import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"
import type { NavGroup } from "@/components/layout/nav-main"

import * as React from "react"

import { NavMain } from "@/components/layout/nav-main"
import { NavUser } from "@/components/layout/nav-user"
import { Logo } from "@/components/logo"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"

export function AppSidebar({
  user,
  navGroups,
  headerExtra,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  user?: GithubCom4H1RZooraInternalDomainUser
  navGroups: NavGroup[]
  headerExtra?: React.ReactNode
}) {
  const { state } = useSidebar()

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <div
          className={cn("hidden h-16 items-center md:flex", state === "expanded" ? "gap-2 px-2.5" : "justify-center")}
        >
          {state === "expanded" && <Logo />}
          {state === "expanded" && <SidebarTrigger className="ms-auto" />}
          {state === "collapsed" && <SidebarTrigger />}
        </div>
        <div className="flex h-16 items-center gap-2 px-2.5 md:hidden">
          <Logo />
        </div>
        {headerExtra}
      </SidebarHeader>
      <SidebarContent>
        <NavMain groups={navGroups} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
    </Sidebar>
  )
}

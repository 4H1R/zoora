import { Link, useRouterState } from "@tanstack/react-router"

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"

export interface NavItem {
  title: string
  url: string
  icon?: React.ReactNode
}

export interface NavGroup {
  label: string
  items: NavItem[]
  indent?: boolean
}

export function NavMain({ groups }: { groups: NavGroup[] }) {
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  return (
    <>
      {groups.map((group) => (
        <SidebarGroup key={group.label}>
          <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
          <SidebarMenu className="gap-0.5">
            {group.items.map((item) => (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton
                  tooltip={item.title}
                  isActive={pathname.startsWith(item.url)}
                  render={<Link to={item.url} />}
                  className={cn(
                    "data-active:[&_svg]:text-primary gap-2.5 px-2.5 py-1.5 text-sm [&_svg]:size-4 [&_svg]:[stroke-width:1.75]",
                    group.indent && "ps-6"
                  )}
                >
                  {item.icon}
                  <span>{item.title}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  )
}

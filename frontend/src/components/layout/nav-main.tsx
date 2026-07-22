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
  // Optional trailing content (e.g. an unread-count pill). Rendered at the
  // logical end of the row; existing items without it are unaffected.
  badge?: React.ReactNode
}

export interface NavGroup {
  label: string
  items: NavItem[]
  indent?: boolean
}

export function NavMain({ groups }: { groups: NavGroup[] }) {
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  // Only the most-specific match is active. A nested route like
  // /org/settings/custom-fields matches both its own url and the parent
  // /org/settings, so pick the longest matching prefix and light up only that.
  const activeUrl = groups
    .flatMap((g) => g.items.map((i) => i.url))
    .filter((url) => pathname === url || pathname.startsWith(url + "/"))
    .sort((a, b) => b.length - a.length)[0]

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
                  isActive={item.url === activeUrl}
                  render={<Link to={item.url} />}
                  className={cn(
                    "data-active:[&_svg]:text-primary gap-2.5 px-2.5 py-1.5 text-sm [&_svg]:size-4 [&_svg]:[stroke-width:1.75]",
                    group.indent && "ps-6"
                  )}
                >
                  {item.icon}
                  <span className="flex-1 truncate">{item.title}</span>
                  {item.badge}
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
      ))}
    </>
  )
}

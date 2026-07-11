import { Link, useRouterState } from "@tanstack/react-router"
import { GraduationCapIcon, SparklesIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetChangelogStatus } from "@/api/changelog/changelog"
import { Badge } from "@/components/ui/badge"
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

type ZooraItem = {
  title: string
  url: string
  icon: React.ReactNode
  // Optional trailing badge (e.g. unseen changelog count). Hidden when 0.
  badge?: number
}

// The "Zoora" brand section of the org sidebar. Groups product-level links —
// What's New today, Learn / help / etc. later. Add a route to `items` to extend.
export function NavZoora() {
  const { t } = useTranslation()
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const { data } = useGetChangelogStatus()
  const status = (data?.status === 200 && data.data.data) || undefined
  const unseen = status?.unseen_count ?? 0

  const items: ZooraItem[] = [
    {
      title: t("whatsNew.title"),
      url: "/org/whats-new",
      icon: <SparklesIcon />,
      badge: unseen,
    },
    {
      // Evergreen help library — no badge (nothing time-sensitive to nag about).
      title: t("tutorials.title"),
      url: "/org/tutorials",
      icon: <GraduationCapIcon />,
    },
  ]

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{t("org.nav.zoora")}</SidebarGroupLabel>
      <SidebarMenu className="gap-0.5">
        {items.map((item) => (
          <SidebarMenuItem key={item.url}>
            <SidebarMenuButton
              tooltip={item.title}
              isActive={pathname.startsWith(item.url)}
              render={<Link to={item.url} />}
              className="data-active:[&_svg]:text-primary gap-2.5 px-2.5 py-1.5 text-sm [&_svg]:size-4 [&_svg]:[stroke-width:1.75]"
            >
              {item.icon}
              <span>{item.title}</span>
              {!!item.badge && item.badge > 0 && (
                <Badge className="ms-auto h-4 min-w-4 justify-center rounded-full px-1 text-[10px] group-data-[collapsible=icon]:hidden">
                  {item.badge > 9 ? "9+" : item.badge}
                </Badge>
              )}
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}

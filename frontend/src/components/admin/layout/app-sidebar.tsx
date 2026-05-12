import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"
import type { NavGroup } from "@/components/layout/nav-main"

import { useRouterState } from "@tanstack/react-router"
import {
  Building2Icon,
  CalendarIcon,
  DumbbellIcon,
  FileVideoIcon,
  KeyIcon,
  LayoutDashboardIcon,
  SchoolIcon,
  ShieldIcon,
  TrophyIcon,
  UsersIcon,
  VideoIcon,
} from "lucide-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { OrgSwitcher } from "@/components/admin/layout/org-switcher"
import { AppSidebar as AppSidebarShared } from "@/components/layout/app-sidebar"
import { Sidebar } from "@/components/ui/sidebar"
import { useAdminStore } from "@/stores/admin"

const CLASS_ID_RE = /^\/admin\/classes\/([^/]+)(?:\/|$)/

export function AppSidebar({
  user,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  user?: GithubCom4H1RZooraInternalDomainUser
}) {
  const { t } = useTranslation()
  const { activeOrganization, setActiveOrganization } = useAdminStore()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  const classIdMatch = pathname.match(CLASS_ID_RE)
  const activeClassId = classIdMatch?.[1]

  const navGroups: NavGroup[] = [
    {
      label: t("admin.platform"),
      items: [
        { title: t("admin.dashboard.title"), url: "/admin/dashboard", icon: <LayoutDashboardIcon /> },
        { title: t("admin.nav.classes"), url: "/admin/classes", icon: <SchoolIcon /> },
        { title: t("admin.nav.liveRooms"), url: "/admin/live-rooms", icon: <VideoIcon /> },
        { title: t("admin.nav.offlines"), url: "/admin/offlines", icon: <FileVideoIcon /> },
        { title: t("admin.nav.practices"), url: "/admin/practices", icon: <DumbbellIcon /> },
        { title: t("admin.organizations"), url: "/admin/organizations", icon: <Building2Icon /> },
      ],
    },
    {
      label: t("admin.nav.users"),
      items: [
        { title: t("admin.nav.users"), url: "/admin/users", icon: <UsersIcon /> },
        { title: t("admin.nav.roles"), url: "/admin/roles", icon: <ShieldIcon /> },
        { title: t("admin.permissions.title"), url: "/admin/permissions", icon: <KeyIcon /> },
      ],
    },
  ]

  if (activeClassId) {
    navGroups.push({
      label: t("admin.classManagement.title"),
      indent: true,
      items: [
        {
          title: t("admin.classManagement.sessions"),
          url: `/admin/classes/${activeClassId}/sessions`,
          icon: <CalendarIcon />,
        },
        {
          title: t("admin.classManagement.gradebook"),
          url: `/admin/classes/${activeClassId}/gradebook`,
          icon: <TrophyIcon />,
        },
      ],
    })
  }

  return (
    <AppSidebarShared
      user={user}
      navGroups={navGroups}
      headerExtra={<OrgSwitcher selected={activeOrganization} onSelect={setActiveOrganization} />}
      {...props}
    />
  )
}

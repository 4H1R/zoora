import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"

import { Building2Icon, KeyIcon, LayoutDashboardIcon, SchoolIcon, ShieldIcon, UsersIcon } from "lucide-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { OrgSwitcher } from "@/components/admin/layout/org-switcher"
import { AppSidebar as AppSidebarShared } from "@/components/layout/app-sidebar"
import { Sidebar } from "@/components/ui/sidebar"
import { useAdminStore } from "@/stores/admin"

export function AppSidebar({
  user,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  user?: GithubCom4H1RZooraInternalDomainUser
}) {
  const { t } = useTranslation()
  const { activeOrganization, setActiveOrganization } = useAdminStore()

  const navGroups = [
    {
      label: t("admin.platform"),
      items: [
        { title: t("admin.dashboard.title"), url: "/admin/dashboard", icon: <LayoutDashboardIcon /> },
        { title: t("admin.nav.classes"), url: "/admin/classes", icon: <SchoolIcon /> },
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

  return (
    <AppSidebarShared
      user={user}
      navGroups={navGroups}
      headerExtra={<OrgSwitcher selected={activeOrganization} onSelect={setActiveOrganization} />}
      {...props}
    />
  )
}

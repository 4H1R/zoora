import type { AppPermission } from "@/lib/access"
import type { NavGroup } from "@/components/layout/nav-main"
import type { TFunction } from "i18next"

import { FileIcon, LayoutDashboardIcon, SchoolIcon, SettingsIcon, ShieldIcon, UsersIcon } from "lucide-react"

type NavItemSpec = {
  title: string
  url: string
  icon: React.ReactNode
  perms?: AppPermission[]
}

type NavGroupSpec = {
  label: string
  items: NavItemSpec[]
}

export function buildOrgNavGroups(
  t: TFunction,
  orgId: string,
  has: (perm: AppPermission) => boolean
): NavGroup[] {
  const groups: NavGroupSpec[] = [
    {
      label: t("org.panel"),
      items: [
        { title: t("org.nav.dashboard"), url: `/org/${orgId}/dashboard`, icon: <LayoutDashboardIcon /> },
        {
          title: t("org.nav.classes"),
          url: `/org/${orgId}/classes`,
          icon: <SchoolIcon />,
          perms: ["classes:view", "classes:view_any"],
        },
      ],
    },
    {
      label: t("org.nav.management"),
      items: [
        {
          title: t("org.nav.users"),
          url: `/org/${orgId}/users`,
          icon: <UsersIcon />,
          perms: ["users:view", "users:view_any"],
        },
        {
          title: t("org.nav.roles"),
          url: `/org/${orgId}/roles`,
          icon: <ShieldIcon />,
          perms: ["roles:view"],
        },
        {
          title: t("org.nav.settings"),
          url: `/org/${orgId}/settings`,
          icon: <SettingsIcon />,
          perms: ["organizations:update"],
        },
        {
          title: t("org.nav.files"),
          url: `/org/${orgId}/files`,
          icon: <FileIcon />,
          perms: ["media:view", "media:view_any"],
        },
      ],
    },
  ]

  return groups
    .map((g) => ({
      label: g.label,
      items: g.items.filter((it) => !it.perms || it.perms.some(has)),
    }))
    .filter((g) => g.items.length > 0)
}

import type { NavGroup } from "@/components/layout/nav-main"
import type { AppPermission } from "@/lib/access"
import type { OrgRouteKey } from "@/lib/org-routes"
import type { TFunction } from "i18next"

import { ORG_ROUTES } from "@/lib/org-routes"

// Each group lists the org routes it contains, in display order. The per-route
// metadata (icon, label, perms, path) comes from ORG_ROUTES so the nav and the
// dashboard tiles never drift.
type NavGroupSpec = {
  label: string
  keys: OrgRouteKey[]
}

export function buildOrgNavGroups(t: TFunction, has: (perm: AppPermission) => boolean): NavGroup[] {
  const groups: NavGroupSpec[] = [
    { label: t("org.panel"), keys: ["dashboard", "calendar", "classes"] },
    {
      label: t("org.nav.learning"),
      keys: ["online-classes", "exams", "practices", "grades", "attendance"],
    },
    {
      label: t("org.nav.management"),
      keys: ["users", "roles", "settings", "files", "notifications"],
    },
  ]

  return groups
    .map((g) => ({
      label: g.label,
      items: g.keys
        .map((key) => ORG_ROUTES[key])
        .filter((spec) => !spec.perms || spec.perms.some(has))
        .map((spec) => ({
          title: t(spec.i18nKey),
          url: `/org/${spec.segment}`,
          icon: spec.icon,
        })),
    }))
    .filter((g) => g.items.length > 0)
}

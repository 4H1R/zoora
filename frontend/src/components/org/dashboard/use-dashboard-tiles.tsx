import type { OrgRouteKey } from "@/lib/org-routes"

import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { ORG_ROUTES } from "@/lib/org-routes"

import type { DashboardTileSpec } from "./tile-grid"

// TILE_KEYS lists the launcher tiles in display order. Unlike the sidebar nav,
// the dashboard excludes "dashboard" itself and includes "online-classes".
// Per-route metadata (icon, label, perms, path) comes from ORG_ROUTES so the
// tiles and the nav never drift.
const TILE_KEYS: OrgRouteKey[] = [
  "calendar",
  "classes",
  "online-classes",
  "exams",
  "grades",
  "attendance",
  "users",
  "roles",
  "files",
  "settings",
]

// useDashboardTiles returns the launcher tiles the current user is allowed to
// open, gated by permission. Every role sees the same grid design; only the
// tiles they can access are shown.
export function useDashboardTiles(orgId: string): DashboardTileSpec[] {
  const { can } = useAccess()
  const { t } = useTranslation()

  return TILE_KEYS.map((key) => ({ key, spec: ORG_ROUTES[key] }))
    .filter(({ spec }) => !spec.perms || spec.perms.some((p) => can(p)))
    .map(({ key, spec }) => ({
      key,
      label: t(spec.i18nKey),
      icon: spec.icon,
      to: `/org/$orgId/${spec.segment}`,
      params: { orgId },
    }))
}

import type { OrgRouteKey } from "@/lib/org-routes"

import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { ORG_ROUTES } from "@/lib/org-routes"

import type { DashboardTileSpec } from "./tile-grid"

// Launcher tiles in display order. Excludes "dashboard", includes "online-classes".
// Metadata comes from ORG_ROUTES so tiles and nav never drift.
const TILE_KEYS: OrgRouteKey[] = [
  "calendar",
  "classes",
  "online-classes",
  "exams",
  "grades",
  "attendance",
  "tickets",
  "users",
  "roles",
  "files",
  "settings",
]

// Returns permission-filtered launcher tiles; same grid for all roles.
export function useDashboardTiles(): DashboardTileSpec[] {
  const { can } = useAccess()
  const { t } = useTranslation()

  return TILE_KEYS.map((key) => ({ key, spec: ORG_ROUTES[key] }))
    .filter(({ spec }) => !spec.perms || spec.perms.some((p) => can(p)))
    .map(({ key, spec }) => ({
      key,
      label: t(spec.i18nKey),
      icon: spec.icon,
      to: `/org/${spec.segment}`,
    }))
}

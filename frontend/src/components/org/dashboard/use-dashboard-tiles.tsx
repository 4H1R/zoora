import type { DashboardTileSpec } from "./tile-grid"
import type { OrgRouteKey } from "@/lib/org-routes"

import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useFeatureGate } from "@/lib/entitlements"
import { navVisible } from "@/lib/org-nav"
import { ORG_ROUTES } from "@/lib/org-routes"

// Launcher tiles in display order. Excludes "dashboard", includes "online-classes".
// Metadata comes from ORG_ROUTES so tiles and nav never drift.
const TILE_KEYS: OrgRouteKey[] = [
  "calendar",
  "classes",
  "conversations",
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

// Returns launcher tiles gated by the SAME rule as the sidebar nav (navVisible):
// permission + plan feature, with the manage-only upgrade exemption. Same grid
// for all roles.
export function useDashboardTiles(): DashboardTileSpec[] {
  const { can } = useAccess()
  const { t } = useTranslation()
  const hasFeature = useFeatureGate()

  return TILE_KEYS.map((key) => ({ key, spec: ORG_ROUTES[key] }))
    .filter(({ spec }) => navVisible(spec, can, hasFeature))
    .map(({ key, spec }) => ({
      key,
      label: t(spec.i18nKey),
      icon: spec.icon,
      to: `/org/${spec.segment}`,
    }))
}

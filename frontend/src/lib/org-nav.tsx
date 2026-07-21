import type { NavGroup } from "@/components/layout/nav-main"
import type { AppPermission } from "@/lib/access"
import type { FeatureKey } from "@/lib/entitlements"
import type { OrgRouteKey, OrgRouteSpec } from "@/lib/org-routes"
import type { TFunction } from "i18next"

import { ConversationsNavBadge } from "@/components/org/conversations/conversations-nav-badge"
import { ORG_ROUTES } from "@/lib/org-routes"

// Each group lists the org routes it contains, in display order. The per-route
// metadata (icon, label, perms, path) comes from ORG_ROUTES so the nav and the
// dashboard tiles never drift.
type NavGroupSpec = {
  label: string
  keys: OrgRouteKey[]
}

// navVisible decides whether a route appears in the sidebar. Two gates:
//  1. perms — the caller must hold at least one (undefined = always allowed).
//  2. feature — the org plan must include it. When it doesn't, the item is
//     hidden, EXCEPT for holders of featureExemptPerms, who see it and land on
//     an upgrade page (drives conversion for the people who can pay).
export function navVisible(
  spec: OrgRouteSpec,
  has: (perm: AppPermission) => boolean,
  hasFeature: (feature: FeatureKey) => boolean
): boolean {
  if (spec.perms && !spec.perms.some(has)) return false
  if (spec.feature && !hasFeature(spec.feature)) {
    return !!spec.featureExemptPerms?.some(has)
  }
  return true
}

export function buildOrgNavGroups(
  t: TFunction,
  has: (perm: AppPermission) => boolean,
  // hasFeature defaults to always-true so callers without entitlement data
  // (or during load) fall back to permission-only gating.
  hasFeature: (feature: FeatureKey) => boolean = () => true
): NavGroup[] {
  const groups: NavGroupSpec[] = [
    { label: t("org.panel"), keys: ["dashboard", "calendar", "classes", "conversations"] },
    {
      label: t("org.nav.learning"),
      keys: ["online-classes", "exams", "question-banks", "practices", "grades", "attendance", "tickets"],
    },
    {
      label: t("org.nav.management"),
      keys: ["users", "roles", "settings", "custom-fields", "billing", "files", "notifications"],
    },
  ]

  return groups
    .map((g) => ({
      label: g.label,
      items: g.keys
        .map((key) => ({ key, spec: ORG_ROUTES[key] }))
        .filter(({ spec }) => navVisible(spec, has, hasFeature))
        .map(({ key, spec }) => ({
          title: t(spec.i18nKey),
          url: `/org/${spec.segment}`,
          icon: spec.icon,
          // Live unread count, gated inside the component on the chat
          // entitlement — see conversations-nav-badge.tsx / use-total-unread.ts.
          badge: key === "conversations" ? <ConversationsNavBadge /> : undefined,
        })),
    }))
    .filter((g) => g.items.length > 0)
}

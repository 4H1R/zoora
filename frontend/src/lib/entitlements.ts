import { useGetUsersMeEntitlements } from "@/api/users/users"

// FeatureKey mirrors backend domain.Feature (internal/domain/plan.go). Add keys
// here as the backend gains plan-gated features; the values must match exactly.
export const FEATURE = {
  recording: "recording",
  offlineRooms: "offline_rooms",
  advancedAntiCheat: "advanced_anticheat",
  customRoles: "custom_roles",
  sso: "sso",
  whiteboard: "whiteboard",
  chat: "chat",
  connectors: "connectors",
  autoGrading: "auto_grading",
  ai: "ai",
} as const

export type FeatureKey = (typeof FEATURE)[keyof typeof FEATURE]

// useEntitlements reads the caller org's resolved plan entitlements
// (GET /users/me/entitlements) — the authoritative feature/limit snapshot the
// backend gates on. The SPA uses it to gate plan-locked UI (nav, paywalls)
// without duplicating the tier→feature rule.
export function useEntitlements() {
  const { data, isLoading } = useGetUsersMeEntitlements()
  const entitlements = (data?.status === 200 && data.data.data) || undefined
  return { entitlements, isLoading }
}

// useHasFeature reports whether the org's plan includes a feature. While the
// entitlements request is in flight, `enabled` is false — gate on isLoading if
// you need to distinguish "loading" from "not entitled".
export function useHasFeature(feature: FeatureKey): { enabled: boolean; isLoading: boolean } {
  const { entitlements, isLoading } = useEntitlements()
  return { enabled: !!entitlements?.features?.[feature], isLoading }
}

// useFeatureGate returns a predicate for gating many features at once — the shape
// navVisible (org-nav.tsx) expects. Use it wherever you'd otherwise re-write the
// `(f) => !!entitlements?.features?.[f]` closure (sidebar nav, dashboard tiles).
// Loading resolves to false, matching useHasFeature.
export function useFeatureGate(): (feature: FeatureKey) => boolean {
  const { entitlements } = useEntitlements()
  return (feature) => !!entitlements?.features?.[feature]
}

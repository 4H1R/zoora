import type { LucideIcon } from "lucide-react"

import { CrownIcon, RocketIcon, SparklesIcon, ZapIcon } from "lucide-react"

// Plan keys are "<tier>_<size>" (e.g. "pro_200"): tier = feature level,
// size = member capacity. Mirrors the backend catalog in internal/domain/plan.go.
export const PLAN_TIERS = ["free", "plus", "pro", "max"] as const
export type PlanTier = (typeof PLAN_TIERS)[number]

export const PAID_TIERS = ["plus", "pro", "max"] as const

export const PLAN_SIZES = [50, 100, 200, 500, 1000] as const

export function planTier(plan?: string | null): PlanTier {
  const tier = plan?.split("_")[0] ?? ""
  return (PLAN_TIERS as readonly string[]).includes(tier) ? (tier as PlanTier) : "free"
}

export function planSize(plan?: string | null): number {
  const size = Number(plan?.split("_")[1])
  return Number.isFinite(size) && size > 0 ? size : PLAN_SIZES[0]
}

export function planKey(tier: PlanTier, size: number): string {
  return `${tier}_${size}`
}

export const TIER_RANK: Record<PlanTier, number> = { free: 0, plus: 1, pro: 2, max: 3 }

// Tier-first, size-second ordering — matches backend planRank so client-side
// downgrade guards agree with the API's DOWNGRADE_NOT_ALLOWED.
export function planRank(plan?: string | null): number {
  const sizeIdx = (PLAN_SIZES as readonly number[]).indexOf(planSize(plan))
  return TIER_RANK[planTier(plan)] * 100 + (sizeIdx < 0 ? 0 : sizeIdx + 1)
}

export const TIER_ICON: Record<PlanTier, LucideIcon> = {
  free: SparklesIcon,
  plus: ZapIcon,
  pro: RocketIcon,
  max: CrownIcon,
}

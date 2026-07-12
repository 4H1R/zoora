import { toast } from "sonner"

import i18n from "@/i18n"

// PlanErrorCode mirrors the two 402 codes MapError emits for plan/entitlement
// gates (internal/domain/errors.go). Any other error code is not a plan error.
export const PLAN_ERROR_CODES = ["FEATURE_NOT_IN_PLAN", "PLAN_LIMIT_REACHED"] as const
export type PlanErrorCode = (typeof PLAN_ERROR_CODES)[number]

// planDetail is the structured 402 payload the backend attaches under
// error.plan_detail (domain.PlanError) so the client can name the exact feature.
interface PlanDetail {
  plan?: string
  feature?: string
  limit?: string
}

interface ApiErrorBody {
  code?: string
  message?: string
  plan_detail?: PlanDetail
}

// unwrap pulls the standardized { error } body out of a redaxios error, whatever
// its declared type. Returns undefined when the shape doesn't match.
function unwrap(error: unknown): ApiErrorBody | undefined {
  const res = (error as { response?: { data?: { error?: ApiErrorBody } } })?.response
  return res?.data?.error
}

export function isPlanError(error: unknown): boolean {
  const code = unwrap(error)?.code
  return (PLAN_ERROR_CODES as readonly string[]).includes(code ?? "")
}

// planErrorMessage resolves the localized upgrade message for a plan error,
// preferring a feature-specific string keyed by plan_detail.feature and falling
// back to the generic feature/limit copy. Returns null for non-plan errors.
export function planErrorMessage(error: unknown): string | null {
  const body = unwrap(error)
  if (!body || !(PLAN_ERROR_CODES as readonly string[]).includes(body.code ?? "")) {
    return null
  }
  const feature = body.plan_detail?.feature
  if (feature && i18n.exists(`planError.features.${feature}`)) {
    return i18n.t(`planError.features.${feature}`)
  }
  return i18n.t(
    body.code === "PLAN_LIMIT_REACHED" ? "planError.limitGeneric" : "planError.featureGeneric"
  )
}

// showPlanErrorToast surfaces a plan-gate error as an upgrade toast. Returns true
// when it handled the error, so callers can early-return before their own copy.
export function showPlanErrorToast(error: unknown): boolean {
  const message = planErrorMessage(error)
  if (!message) return false
  toast.error(message)
  return true
}

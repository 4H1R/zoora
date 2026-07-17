import type { FieldValues, Path, UseFormSetError } from "react-hook-form"

import i18n from "@/i18n"

// ApiErrorBody mirrors domain.ErrorBody (internal/domain/response.go): the
// standardized { success, error } envelope the backend returns on failure.
// `request_id` is present only on 5xx responses (where `message` is scrubbed) so
// a user can quote it and support can grep straight to the server log line.
export interface ApiErrorBody {
  code?: string
  message?: string
  fields?: Record<string, string>
  request_id?: string
  plan_detail?: {
    plan?: string
    feature?: string
    limit?: string
  }
}

interface ErrorWithResponse {
  response?: {
    status?: number
    data?: { error?: ApiErrorBody }
  }
}

// apiError pulls the standardized error body out of a thrown redaxios error,
// whatever its declared type. Returns undefined when the shape doesn't match
// (e.g. a network failure with no response, or a non-envelope body).
export function apiError(error: unknown): ApiErrorBody | undefined {
  return (error as ErrorWithResponse)?.response?.data?.error
}

// apiErrorCode returns the machine-readable error code (domain error mapping),
// or undefined. Use it to branch on specific expected failures.
export function apiErrorCode(error: unknown): string | undefined {
  return apiError(error)?.code
}

// apiErrorStatus returns the HTTP status of a thrown redaxios error, or 0 when
// there is no response at all (network / CORS / offline).
export function apiErrorStatus(error: unknown): number {
  return (error as ErrorWithResponse)?.response?.status ?? 0
}

// isServerError reports the "unexpected" error class the global handler owns: a
// 5xx from the backend, or a transport failure with no response. Call sites own
// their specific 4xx copy, so the global fallback toast only fires on this set.
export function isServerError(error: unknown): boolean {
  const status = apiErrorStatus(error)
  return status === 0 || status >= 500
}

// apiErrorMessage resolves a human message for a thrown error. It prefers the
// backend's own error.message (localized/meaningful for 4xx). On a 5xx the
// server scrubs the message, so we use a localized generic and append the
// request_id for support. Falls back to `fallback` (or a generic string) when
// no envelope is present.
export function apiErrorMessage(error: unknown, fallback?: string): string {
  const body = apiError(error)
  if (isServerError(error)) {
    const base = fallback ?? i18n.t("common.apiError.serverGeneric")
    const rid = body?.request_id
    return rid ? `${base} (${i18n.t("common.apiError.requestId")}: ${rid})` : base
  }
  if (body?.message) return body.message
  return fallback ?? i18n.t("common.apiError.generic")
}

// applyFieldErrors maps the backend's per-field validation errors (error.fields,
// from domain.ValidationError) onto a React Hook Form via setError, so a 400
// shows inline on the offending inputs instead of a generic toast. Returns true
// when at least one field error was applied, letting callers skip their toast.
export function applyFieldErrors<T extends FieldValues>(
  error: unknown,
  setError: UseFormSetError<T>,
): boolean {
  const fields = apiError(error)?.fields
  if (!fields) return false
  let applied = false
  for (const [name, message] of Object.entries(fields)) {
    setError(name as Path<T>, { message })
    applied = true
  }
  return applied
}

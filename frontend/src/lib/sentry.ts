import * as Sentry from "@sentry/react"

import { clientEnv } from "@/config/env"

// initSentry starts the browser Sentry client when VITE_SENTRY_DSN is set. It is
// fully optional: with no DSN it does nothing, so the app runs unchanged until a
// DSN is added later. Browser-only — guarded so it never runs during the SSR
// prerender of the marketing landing.
export function initSentry() {
  if (typeof window === "undefined") return

  const dsn = clientEnv.VITE_SENTRY_DSN
  if (!dsn) return

  const tracesSampleRate = clientEnv.VITE_SENTRY_TRACES_SAMPLE_RATE ?? 0
  Sentry.init({
    dsn,
    environment: clientEnv.VITE_SENTRY_ENVIRONMENT ?? import.meta.env.MODE,
    tracesSampleRate,
    integrations: tracesSampleRate > 0 ? [Sentry.browserTracingIntegration()] : [],
  })
}

// reportError forwards an error to Sentry. Safe to call when Sentry is disabled:
// with no active client captureException is a no-op. Central seam so callers
// don't import the SDK directly.
export function reportError(error: unknown) {
  Sentry.captureException(error)
}

import { z } from "zod"

const clientEnvSchema = z.object({
  VITE_API_URL: z.string().min(1),
  // WebSocket endpoint for the realtime conversations hub.
  VITE_WS_URL: z.string().min(1),
  // Firebase web-push config. Optional — when any of these is unset the browser
  // push connector hides its Connect button (deployment has push disabled).
  VITE_FIREBASE_API_KEY: z.string().optional(),
  VITE_FIREBASE_PROJECT_ID: z.string().optional(),
  VITE_FIREBASE_SENDER_ID: z.string().optional(),
  VITE_FIREBASE_APP_ID: z.string().optional(),
  VITE_FIREBASE_VAPID_KEY: z.string().optional(),
  // Sentry error reporting. Optional — unset disables Sentry entirely, so the
  // app runs unchanged until a DSN is added. VITE_SENTRY_TRACES_SAMPLE_RATE
  // (0.0–1.0, default 0) additionally enables performance tracing.
  VITE_SENTRY_DSN: z.string().optional(),
  VITE_SENTRY_TRACES_SAMPLE_RATE: z.coerce.number().min(0).max(1).optional(),
  VITE_SENTRY_ENVIRONMENT: z.string().optional(),
})

export const clientEnv = clientEnvSchema.parse(import.meta.env)

// True only when the full Firebase web-push config is present; the push
// connector card gates its Connect action on this.
export const isPushConfigured =
  !!clientEnv.VITE_FIREBASE_API_KEY &&
  !!clientEnv.VITE_FIREBASE_PROJECT_ID &&
  !!clientEnv.VITE_FIREBASE_SENDER_ID &&
  !!clientEnv.VITE_FIREBASE_APP_ID &&
  !!clientEnv.VITE_FIREBASE_VAPID_KEY

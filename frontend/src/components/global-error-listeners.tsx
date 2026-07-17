import { useEffect } from "react"

import { reportError } from "@/lib/sentry"

// GlobalErrorListeners captures errors that escape React's render tree — uncaught
// exceptions and unhandled promise rejections — which the router error boundary
// never sees (it only catches errors thrown during render/loaders). Without
// this, a rejected promise outside React Query dies silently with no signal.
//
// It logs to the console and forwards to Sentry via reportError (a no-op when
// Sentry is disabled). Mounted once at the app root. The listeners are
// browser-only (attached in an effect), so this is safe under SSR/prerender.
export function GlobalErrorListeners() {
  useEffect(() => {
    const onError = (event: ErrorEvent) => {
      // event.error is the thrown value when available; fall back to message.
      console.error("[uncaught error]", event.error ?? event.message)
      reportError(event.error ?? event.message)
    }
    const onRejection = (event: PromiseRejectionEvent) => {
      console.error("[unhandled rejection]", event.reason)
      reportError(event.reason)
    }

    window.addEventListener("error", onError)
    window.addEventListener("unhandledrejection", onRejection)
    return () => {
      window.removeEventListener("error", onError)
      window.removeEventListener("unhandledrejection", onRejection)
    }
  }, [])

  return null
}

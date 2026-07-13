import { useEffect } from "react"

/**
 * Offline/precache PWA was dropped during the TanStack Start migration:
 * vite-plugin-pwa's generateSW does not emit a service worker under Start's
 * Nitro build pipeline. This component now only cleans up: it unregisters the
 * stale Workbox service worker left on returning users' devices so they aren't
 * pinned to a cached old app shell. The Firebase Cloud Messaging worker
 * (firebase-messaging-sw.js) is intentionally left untouched.
 *
 * Re-introduce offline support later via a post-build Workbox step.
 */
export function PWAUpdater() {
  useEffect(() => {
    if (typeof navigator === "undefined" || !("serviceWorker" in navigator)) return

    navigator.serviceWorker.getRegistrations().then((registrations) => {
      for (const registration of registrations) {
        const scriptURL = registration.active?.scriptURL ?? ""
        // Keep the FCM push worker; drop the old Workbox app-shell worker.
        if (scriptURL.includes("firebase-messaging-sw")) continue
        registration.unregister()
      }
    })
  }, [])

  return null
}

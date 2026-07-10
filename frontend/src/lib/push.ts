import { initializeApp } from "firebase/app"
import { getMessaging, getToken, isSupported } from "firebase/messaging"

import { clientEnv } from "@/config/env"

// Returns an FCM registration token after the user grants permission, or a
// typed failure the connector card renders as guidance.
export async function enablePush(): Promise<
  { ok: true; token: string } | { ok: false; reason: "unsupported" | "denied" }
> {
  if (!(await isSupported())) return { ok: false, reason: "unsupported" }

  const permission = await Notification.requestPermission()
  if (permission !== "granted") return { ok: false, reason: "denied" }

  const app = initializeApp({
    apiKey: clientEnv.VITE_FIREBASE_API_KEY,
    projectId: clientEnv.VITE_FIREBASE_PROJECT_ID,
    messagingSenderId: clientEnv.VITE_FIREBASE_SENDER_ID,
    appId: clientEnv.VITE_FIREBASE_APP_ID,
  })
  // Register under a dedicated subpath scope so this SW does NOT share scope "/"
  // with the workbox PWA service worker. Same scope = one registration slot, so
  // the two would overwrite each other on every load and each swap fires
  // workbox's onNeedRefresh → "new version available" toast spam.
  const registration = await navigator.serviceWorker.register("/firebase-messaging-sw.js", {
    scope: "/firebase-cloud-messaging-push-scope",
  })
  const token = await getToken(getMessaging(app), {
    vapidKey: clientEnv.VITE_FIREBASE_VAPID_KEY,
    serviceWorkerRegistration: registration,
  })
  return { ok: true, token }
}

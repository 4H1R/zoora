import { useEffect, useState } from "react"

import { clearDeferredPrompt, getDeferredPrompt, subscribeInstallPrompt } from "@/components/pwa/install-prompt-store"

/**
 * Chrome/Edge fire `beforeinstallprompt` when the PWA install criteria are met
 * (valid manifest + service worker + HTTPS). The event is captured globally at
 * app startup (see install-prompt-store) because it fires once, early — before
 * this hook's component mounts. We read the stashed event here so our own UI can
 * trigger the native install dialog on a user gesture.
 */
const DISMISS_KEY = "zoora:pwa-install-dismissed"
// Re-offer the banner after three days instead of hiding it forever — a one-time
// "not now" shouldn't permanently kill install discovery on that device.
const DISMISS_TTL_MS = 1000 * 60 * 60 * 24 * 3

function isStandalone() {
  if (typeof window === "undefined") return false
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    // iOS Safari exposes standalone via a non-standard navigator flag.
    (window.navigator as unknown as { standalone?: boolean }).standalone === true
  )
}

/**
 * iOS Safari never fires `beforeinstallprompt` — installation is manual via the
 * Share sheet. We only offer the instructions there (Chrome/Firefox on iOS use
 * the `crios`/`fxios` UA tokens and can't install PWAs at all).
 */
function isIosSafari() {
  if (typeof navigator === "undefined") return false
  const ua = navigator.userAgent
  const iOS = /iphone|ipad|ipod/i.test(ua) || (/macintosh/i.test(ua) && "ontouchend" in document)
  const otherBrowser = /crios|fxios|edgios|opios/i.test(ua)
  return iOS && !otherBrowser
}

function isDismissed() {
  try {
    const raw = localStorage.getItem(DISMISS_KEY)
    if (!raw) return false
    const ts = Number(raw)
    // Legacy value was the string "1" (permanent). Number("1") is ancient vs
    // now, so it falls outside the TTL and we treat it as expired — the banner
    // gets one more chance instead of staying hidden forever.
    if (!Number.isFinite(ts)) return false
    return Date.now() - ts < DISMISS_TTL_MS
  } catch {
    return false
  }
}

export function usePwaInstall() {
  const [deferred, setDeferred] = useState(getDeferredPrompt)
  const [installed, setInstalled] = useState(isStandalone)
  const [dismissed, setDismissed] = useState(isDismissed)

  useEffect(() => {
    // Seed from the stash (event may have fired before mount) and subscribe for
    // any later arrival / appinstalled clear.
    setDeferred(getDeferredPrompt())
    const unsub = subscribeInstallPrompt(() => setDeferred(getDeferredPrompt()))
    const onInstalled = () => setInstalled(true)
    window.addEventListener("appinstalled", onInstalled)
    return () => {
      unsub()
      window.removeEventListener("appinstalled", onInstalled)
    }
  }, [])

  const ios = isIosSafari()
  // Only surface the banner when install is actually possible and the user
  // hasn't already installed or recently dismissed it.
  const canShow = !installed && !dismissed && (!!deferred || ios)

  async function promptInstall() {
    const evt = getDeferredPrompt()
    if (!evt) return "unavailable" as const
    await evt.prompt()
    const choice = await evt.userChoice
    // The event can only be used once; drop it whatever the outcome.
    clearDeferredPrompt()
    setDeferred(null)
    if (choice.outcome === "accepted") setInstalled(true)
    return choice.outcome
  }

  function dismiss() {
    setDismissed(true)
    try {
      localStorage.setItem(DISMISS_KEY, String(Date.now()))
    } catch {
      // ignore — private mode / storage disabled
    }
  }

  return { canShow, ios, hasNativePrompt: !!deferred, promptInstall, dismiss }
}

import { useEffect, useState } from "react"

/**
 * Chrome/Edge fire `beforeinstallprompt` when the PWA install criteria are met
 * (valid manifest + service worker + HTTPS). We capture it so we can trigger the
 * native install dialog from our own UI on a user gesture — the browser no
 * longer surfaces an automatic banner on its own.
 */
type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>
}

const DISMISS_KEY = "zoora:pwa-install-dismissed"

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

export function usePwaInstall() {
  const [deferred, setDeferred] = useState<BeforeInstallPromptEvent | null>(null)
  const [installed, setInstalled] = useState(isStandalone)
  const [dismissed, setDismissed] = useState(() => {
    try {
      return localStorage.getItem(DISMISS_KEY) === "1"
    } catch {
      return false
    }
  })

  useEffect(() => {
    const onPrompt = (e: Event) => {
      // Stop the browser's own mini-infobar so our banner is the single entry point.
      e.preventDefault()
      setDeferred(e as BeforeInstallPromptEvent)
    }
    const onInstalled = () => {
      setInstalled(true)
      setDeferred(null)
    }
    window.addEventListener("beforeinstallprompt", onPrompt)
    window.addEventListener("appinstalled", onInstalled)
    return () => {
      window.removeEventListener("beforeinstallprompt", onPrompt)
      window.removeEventListener("appinstalled", onInstalled)
    }
  }, [])

  const ios = isIosSafari()
  // Only surface the banner when install is actually possible and the user
  // hasn't already installed or dismissed it.
  const canShow = !installed && !dismissed && (!!deferred || ios)

  async function promptInstall() {
    if (!deferred) return "unavailable" as const
    await deferred.prompt()
    const choice = await deferred.userChoice
    setDeferred(null)
    if (choice.outcome === "accepted") setInstalled(true)
    return choice.outcome
  }

  function dismiss() {
    setDismissed(true)
    try {
      localStorage.setItem(DISMISS_KEY, "1")
    } catch {
      // ignore — private mode / storage disabled
    }
  }

  return { canShow, ios, hasNativePrompt: !!deferred, promptInstall, dismiss }
}

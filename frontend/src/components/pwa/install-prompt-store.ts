/**
 * Global, side-effecting capture of Chrome/Edge's `beforeinstallprompt` event.
 *
 * The event fires ONCE, very early in page load — often before React has
 * mounted the route that renders the install banner. If we only listened from
 * inside a component effect, we'd miss it and the native install prompt would
 * be lost for the whole session. So we attach the listener at module-import
 * time (imported first thing in main.tsx) and stash the event here; the
 * `usePwaInstall` hook seeds itself from the stash and subscribes for updates.
 */
type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>
}

let deferredPrompt: BeforeInstallPromptEvent | null = null
const listeners = new Set<() => void>()

function emit() {
  for (const l of listeners) l()
}

if (typeof window !== "undefined") {
  window.addEventListener("beforeinstallprompt", (e) => {
    // Suppress Chrome's own mini-infobar; our banner is the single entry point.
    e.preventDefault()
    deferredPrompt = e as BeforeInstallPromptEvent
    emit()
  })
  window.addEventListener("appinstalled", () => {
    deferredPrompt = null
    emit()
  })
}

export function getDeferredPrompt() {
  return deferredPrompt
}

export function clearDeferredPrompt() {
  deferredPrompt = null
  emit()
}

/** Subscribe to changes in the stashed prompt. Returns an unsubscribe fn. */
export function subscribeInstallPrompt(cb: () => void) {
  listeners.add(cb)
  return () => {
    listeners.delete(cb)
  }
}

export type { BeforeInstallPromptEvent }

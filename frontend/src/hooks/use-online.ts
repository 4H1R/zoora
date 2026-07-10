import { useSyncExternalStore } from "react"

/**
 * Tracks browser reachability via `navigator.onLine` and the window
 * `online`/`offline` events. This is a coarse signal — the browser reports
 * "online" as soon as it has *a* network interface, which doesn't guarantee the
 * API is reachable — but it reliably catches the common case (Wi-Fi dropped,
 * airplane mode, cable pulled) that stops every fetch and mutation cold.
 *
 * Implemented with `useSyncExternalStore` so it's SSR-safe and never tears:
 * the server snapshot is optimistically `true` (assume online) to avoid a
 * false offline flash on first paint.
 */
function subscribe(onChange: () => void): () => void {
  window.addEventListener("online", onChange)
  window.addEventListener("offline", onChange)
  return () => {
    window.removeEventListener("online", onChange)
    window.removeEventListener("offline", onChange)
  }
}

const getSnapshot = () => navigator.onLine
const getServerSnapshot = () => true

export function useOnline(): boolean {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)
}

import { useEffect, useRef } from "react"

export interface TabVisibilityStats {
  count: number
  seconds: number
}

// useTabVisibility counts how many times the exam tab became hidden and the
// total seconds spent hidden, while `enabled` is true. It listens to both
// `visibilitychange` and window blur/focus, de-duplicating overlapping events
// via an internal isHidden flag so a single tab switch counts once. `onReturn`
// fires with the running count each time the student comes back. Read the
// accumulated stats at submit time via the returned `read()`.
export function useTabVisibility(enabled: boolean, onReturn?: (count: number) => void) {
  const stateRef = useRef({ count: 0, seconds: 0, hiddenAt: 0, isHidden: false })
  const onReturnRef = useRef(onReturn)
  onReturnRef.current = onReturn

  useEffect(() => {
    if (!enabled) return
    const s = stateRef.current

    function markHidden() {
      if (s.isHidden) return
      s.isHidden = true
      s.hiddenAt = Date.now()
      s.count += 1
    }
    function markVisible() {
      if (!s.isHidden) return
      s.isHidden = false
      s.seconds += Math.max(0, Math.floor((Date.now() - s.hiddenAt) / 1000))
      onReturnRef.current?.(s.count)
    }
    function onVisibility() {
      if (document.hidden) markHidden()
      else markVisible()
    }

    document.addEventListener("visibilitychange", onVisibility)
    window.addEventListener("blur", markHidden)
    window.addEventListener("focus", markVisible)
    return () => {
      document.removeEventListener("visibilitychange", onVisibility)
      window.removeEventListener("blur", markHidden)
      window.removeEventListener("focus", markVisible)
    }
  }, [enabled])

  function read(): TabVisibilityStats {
    const s = stateRef.current
    const pending = s.isHidden ? Math.max(0, Math.floor((Date.now() - s.hiddenAt) / 1000)) : 0
    return { count: s.count, seconds: s.seconds + pending }
  }

  return { read }
}

import * as React from "react"

/**
 * True on coarse-pointer (touch) devices. Use to branch touch-first interactions
 * (tap-to-open menus) from fine-pointer ones (hover, right-click) — this tracks
 * input capability, not viewport width, so a narrow desktop window stays "fine".
 */
export function useCoarsePointer() {
  const [coarse, setCoarse] = React.useState(
    () => typeof window !== "undefined" && window.matchMedia("(pointer: coarse)").matches
  )

  React.useEffect(() => {
    const mql = window.matchMedia("(pointer: coarse)")
    const onChange = () => setCoarse(mql.matches)
    mql.addEventListener("change", onChange)
    setCoarse(mql.matches)
    return () => mql.removeEventListener("change", onChange)
  }, [])

  return coarse
}

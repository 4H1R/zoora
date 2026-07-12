import { useEffect, useRef } from "react"

interface ExamLockdownOptions {
  disableCopyPaste: boolean
  disableShortcuts: boolean
  // Called (on a caller-side throttle) whenever a blocked action is prevented.
  onBlocked?: () => void
}

// useExamLockdown installs document-level, frontend-only deterrents while the
// exam is in progress: it blocks copy/cut/paste when `disableCopyPaste` is set
// and blocks the context menu plus common dev/copy shortcuts (Ctrl/Cmd+C/V/X/A/P,
// F12, Ctrl/Cmd+Shift+I/J/C) when `disableShortcuts` is set. All handlers call
// preventDefault; `onBlocked` is invoked so the caller can surface a toast.
export function useExamLockdown({
  disableCopyPaste,
  disableShortcuts,
  onBlocked,
}: ExamLockdownOptions) {
  const onBlockedRef = useRef(onBlocked)
  onBlockedRef.current = onBlocked

  useEffect(() => {
    if (!disableCopyPaste && !disableShortcuts) return

    function onClipboard(e: ClipboardEvent) {
      e.preventDefault()
      onBlockedRef.current?.()
    }
    function onContextMenu(e: MouseEvent) {
      e.preventDefault()
      onBlockedRef.current?.()
    }
    function onKeyDown(e: KeyboardEvent) {
      const key = e.key.toLowerCase()
      const mod = e.ctrlKey || e.metaKey
      const blocked =
        e.key === "F12" ||
        (mod && e.shiftKey && (key === "i" || key === "j" || key === "c")) ||
        (mod && (key === "c" || key === "v" || key === "x" || key === "a" || key === "p"))
      if (blocked) {
        e.preventDefault()
        onBlockedRef.current?.()
      }
    }

    if (disableCopyPaste) {
      document.addEventListener("copy", onClipboard)
      document.addEventListener("cut", onClipboard)
      document.addEventListener("paste", onClipboard)
    }
    if (disableShortcuts) {
      document.addEventListener("contextmenu", onContextMenu)
      document.addEventListener("keydown", onKeyDown)
    }
    return () => {
      document.removeEventListener("copy", onClipboard)
      document.removeEventListener("cut", onClipboard)
      document.removeEventListener("paste", onClipboard)
      document.removeEventListener("contextmenu", onContextMenu)
      document.removeEventListener("keydown", onKeyDown)
    }
  }, [disableCopyPaste, disableShortcuts])
}

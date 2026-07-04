import "tldraw/tldraw.css"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Tldraw, type Editor } from "tldraw"

import { useThemeStore } from "@/stores/theme"

import { useWhiteboard } from "./use-whiteboard"

// Optional tldraw business license — removes the "Get a license" watermark when
// set. Without it the watermark stays (removing it any other way breaks tldraw's
// ToS). Set VITE_TLDRAW_LICENSE_KEY in the frontend env to activate.
const TLDRAW_LICENSE_KEY = import.meta.env.VITE_TLDRAW_LICENSE_KEY as string | undefined

interface WhiteboardStageProps {
  liveId: string
  canDraw: boolean
}

export function WhiteboardStage({ liveId, canDraw }: WhiteboardStageProps) {
  const { store } = useWhiteboard(liveId, canDraw)
  const [editor, setEditor] = useState<Editor | null>(null)

  // Keep readonly in sync with canDraw — role can change mid-session
  // (viewer promoted to presenter, or demoted back).
  const { i18n } = useTranslation()
  // tldraw ships "en" and "fa" (فارسی) locales; map the site language to them.
  const locale = i18n.language?.startsWith("fa") ? "fa" : "en"
  // Follow the app's light/dark theme so the whiteboard chrome matches the room
  // instead of always rendering tldraw's default light UI.
  const theme = useThemeStore((s) => s.theme)

  useEffect(() => {
    if (!editor) return
    editor.updateInstanceState({ isReadonly: !canDraw })
  }, [editor, canDraw])

  // Match tldraw's UI language to the site language, and follow live switches.
  useEffect(() => {
    if (!editor) return
    editor.user.updateUserPreferences({ locale })
  }, [editor, locale])

  // Match tldraw's color scheme to the app theme, and follow live switches.
  useEffect(() => {
    if (!editor) return
    editor.user.updateUserPreferences({ colorScheme: theme })
  }, [editor, theme])

  return (
    <div className="zoora-whiteboard h-full w-full overflow-hidden rounded-2xl">
      <Tldraw store={store} licenseKey={TLDRAW_LICENSE_KEY} onMount={setEditor} />
    </div>
  )
}

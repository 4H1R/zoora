import "tldraw/tldraw.css"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Tldraw, type Editor } from "tldraw"

import { useWhiteboard } from "./use-whiteboard"

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

  useEffect(() => {
    if (!editor) return
    editor.updateInstanceState({ isReadonly: !canDraw })
  }, [editor, canDraw])

  // Match tldraw's UI language to the site language, and follow live switches.
  useEffect(() => {
    if (!editor) return
    editor.user.updateUserPreferences({ locale })
  }, [editor, locale])

  return (
    <div className="zoora-whiteboard h-full w-full overflow-hidden rounded-2xl">
      <Tldraw store={store} onMount={setEditor} />
    </div>
  )
}

import "tldraw/tldraw.css"
import { useEffect, useState } from "react"
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
  useEffect(() => {
    if (!editor) return
    editor.updateInstanceState({ isReadonly: !canDraw })
  }, [editor, canDraw])

  return (
    <div className="zoora-whiteboard h-full w-full overflow-hidden rounded-2xl">
      <Tldraw store={store} onMount={setEditor} />
    </div>
  )
}

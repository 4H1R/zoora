import "tldraw/tldraw.css"
import { Tldraw } from "tldraw"

import { useWhiteboard } from "./use-whiteboard"

interface WhiteboardStageProps {
  liveId: string
  canDraw: boolean
}

export function WhiteboardStage({ liveId, canDraw }: WhiteboardStageProps) {
  const { store } = useWhiteboard(liveId, canDraw)

  return (
    <div className="h-full w-full overflow-hidden rounded-2xl">
      <Tldraw
        store={store}
        onMount={(editor) => {
          if (!canDraw) {
            editor.updateInstanceState({ isReadonly: true })
          }
        }}
      />
    </div>
  )
}

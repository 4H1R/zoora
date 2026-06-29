import { ParticipantTile, useTracks } from "@livekit/components-react"
import { isTrackReference } from "@livekit/components-react"
import { Track } from "livekit-client"

import type { StageContent } from "./use-stage"
import { SlidesStage } from "./slides-stage"
import { WhiteboardStage } from "./whiteboard-stage"

interface StageProps {
  stage: StageContent
  isHost: boolean
  liveId: string
  canDraw: boolean
  onPageChange: (page: number) => void
  onLoadNumPages: (n: number) => void
}

// The single content surface.
// Priority: whiteboard > slides (host-shared PDF) > screenshare > presenter camera > empty.
export function Stage({ stage, isHost, liveId, canDraw, onPageChange, onLoadNumPages }: StageProps) {
  const tracks = useTracks(
    [
      { source: Track.Source.ScreenShare, withPlaceholder: false },
      { source: Track.Source.Camera, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  )

  // Whiteboard takes top priority
  if (stage.kind === "whiteboard") {
    return <WhiteboardStage liveId={liveId} canDraw={canDraw} />
  }

  // Slides take priority over screenshare/camera
  if (stage.kind === "slides" && stage.url) {
    return (
      <SlidesStage
        url={stage.url}
        page={stage.page ?? 1}
        numPages={stage.numPages ?? 0}
        isHost={isHost}
        onPageChange={onPageChange}
        onLoadNumPages={onLoadNumPages}
      />
    )
  }

  const screenShare = tracks.find(
    (tr) => isTrackReference(tr) && tr.publication.source === Track.Source.ScreenShare
  )

  // Presenter-cam fallback: first camera track (only hosts publish in Phase 1).
  const presenterCam = tracks.find(
    (tr) => isTrackReference(tr) && tr.publication.source === Track.Source.Camera
  )

  const active = screenShare ?? presenterCam

  if (!active || !isTrackReference(active)) {
    // Quiet empty surface — no loud placeholder. Keeps the stage container so layout holds.
    return <div className="h-full w-full rounded-2xl border border-border bg-muted/30" />
  }

  // transform-gpu forces this clipping container onto its own GPU layer so the
  // rounded-corner + overflow:hidden clip composites correctly. Without it, some
  // GPUs (Android Chrome/Firefox, certain Windows drivers) paint the clipped
  // <video> black even though frames decode fine. See livekit-overrides.css.
  return (
    <div className="h-full w-full transform-gpu overflow-hidden rounded-2xl bg-black">
      <ParticipantTile trackRef={active} className="h-full w-full" />
    </div>
  )
}

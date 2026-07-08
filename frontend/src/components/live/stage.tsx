import { ParticipantTile, useTracks, VideoTrack } from "@livekit/components-react"
import { isTrackReference } from "@livekit/components-react"
import { Track } from "livekit-client"
import { MonitorUp } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { StageContent } from "./use-stage"
import { SlidesStage } from "./slides-stage"
import { WhiteboardStage } from "./whiteboard-stage"
import { ZoomableStage } from "./zoomable-stage"

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
  const { t } = useTranslation()
  const tracks = useTracks(
    [
      { source: Track.Source.ScreenShare, withPlaceholder: false },
      { source: Track.Source.Camera, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  )

  if (stage.kind === "whiteboard") {
    return <WhiteboardStage liveId={liveId} canDraw={canDraw} />
  }

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

  // LiveKit's default ParticipantTile hardcodes an English "'s screen" suffix on
  // screen-share tiles, so we supply our own translated label there. Camera tiles
  // only show the participant name (no English), so keep the default.
  const isScreenShare = active.publication.source === Track.Source.ScreenShare
  const participantName = active.participant.name || active.participant.identity

  return (
    <ZoomableStage>
      {isScreenShare ? (
        <ParticipantTile trackRef={active} className="size-full">
          <VideoTrack trackRef={active} />
          <div className="lk-participant-metadata">
            <div className="lk-participant-metadata-item">
              <MonitorUp className="me-1 size-3.5" />
              <span>{t("liveRoom.screenShareLabel", { name: participantName })}</span>
            </div>
          </div>
        </ParticipantTile>
      ) : (
        <ParticipantTile trackRef={active} className="size-full" />
      )}
    </ZoomableStage>
  )
}

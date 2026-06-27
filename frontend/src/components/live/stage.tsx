import { ParticipantTile, useTracks } from "@livekit/components-react"
import { isTrackReference } from "@livekit/components-react"
import { Track } from "livekit-client"
import { MonitorPlay } from "lucide-react"
import { useTranslation } from "react-i18next"

// The single content surface. Priority: screenshare > presenter camera > empty.
export function Stage() {
  const { t } = useTranslation()
  const tracks = useTracks(
    [
      { source: Track.Source.ScreenShare, withPlaceholder: false },
      { source: Track.Source.Camera, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  )

  const screenShare = tracks.find(
    (tr) => isTrackReference(tr) && tr.publication.source === Track.Source.ScreenShare
  )

  // Presenter-cam fallback: first camera track (only hosts publish in Phase 1).
  const presenterCam = tracks.find(
    (tr) => isTrackReference(tr) && tr.publication.source === Track.Source.Camera
  )

  const active = screenShare ?? presenterCam

  if (!active || !isTrackReference(active)) {
    return (
      <div className="flex h-full w-full flex-col items-center justify-center gap-3 rounded-2xl border border-white/5 bg-black/40 text-center">
        <span className="flex size-14 items-center justify-center rounded-2xl bg-primary/15 text-primary">
          <MonitorPlay className="size-7" />
        </span>
        <p className="text-sm text-zinc-400">{t("liveRoom.stage.empty")}</p>
      </div>
    )
  }

  return (
    <div className="h-full w-full overflow-hidden rounded-2xl bg-black">
      <ParticipantTile trackRef={active} className="h-full w-full" />
    </div>
  )
}

import { ParticipantTile, isTrackReference, useTracks } from "@livekit/components-react"
import { Track } from "livekit-client"

// Camera tiles for whoever is publishing video (host/presenters). Viewers never
// publish in Phase 1, so this stays small (1-3 tiles) and scales to big classes.
export function WebcamRail({ orientation }: { orientation: "vertical" | "horizontal" }) {
  const tracks = useTracks([{ source: Track.Source.Camera, withPlaceholder: false }], {
    onlySubscribed: false,
  }).filter(isTrackReference)

  if (tracks.length === 0) return null

  const vertical = orientation === "vertical"

  return (
    <div
      className={
        vertical
          ? "flex h-full w-40 shrink-0 flex-col gap-2 overflow-y-auto lg:w-48"
          : "flex w-full gap-2 overflow-x-auto"
      }
    >
      {tracks.map((tr) => (
        <div
          key={tr.publication.trackSid}
          className={
            vertical
              ? "aspect-video w-full shrink-0 overflow-hidden rounded-xl bg-black"
              : "aspect-video h-20 shrink-0 overflow-hidden rounded-xl bg-black"
          }
        >
          <ParticipantTile trackRef={tr} className="h-full w-full" />
        </div>
      ))}
    </div>
  )
}

import {
  CarouselLayout,
  FocusLayout,
  FocusLayoutContainer,
  GridLayout,
  ParticipantTile,
  isTrackReference,
  useTracks,
} from "@livekit/components-react"
import { Track } from "livekit-client"

export function VideoGrid() {
  const tracks = useTracks(
    [
      { source: Track.Source.Camera, withPlaceholder: true },
      { source: Track.Source.ScreenShare, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  )

  // Auto-focus an active screen share; otherwise show an even grid.
  const focusTrack = tracks.find((tr) => isTrackReference(tr) && tr.publication.source === Track.Source.ScreenShare)

  if (focusTrack) {
    const carousel = tracks.filter((tr) => tr !== focusTrack)
    return (
      <FocusLayoutContainer className="h-full">
        <CarouselLayout tracks={carousel}>
          <ParticipantTile />
        </CarouselLayout>
        <FocusLayout trackRef={focusTrack} />
      </FocusLayoutContainer>
    )
  }

  return (
    <GridLayout tracks={tracks} className="h-full">
      <ParticipantTile />
    </GridLayout>
  )
}

import {
  CarouselLayout,
  FocusLayout,
  FocusLayoutContainer,
  GridLayout,
  type LayoutContextType,
  ParticipantTile,
  isTrackReference,
  usePinnedTracks,
  useTracks,
} from "@livekit/components-react"
import { Track } from "livekit-client"
import { useEffect, useRef } from "react"

export function VideoGrid({ layoutContext }: { layoutContext: LayoutContextType }) {
  const tracks = useTracks(
    [
      { source: Track.Source.Camera, withPlaceholder: true },
      { source: Track.Source.ScreenShare, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  )

  const screenShareTrack = tracks.find(
    (tr) => isTrackReference(tr) && tr.publication.source === Track.Source.ScreenShare
  )
  const screenShareSid =
    screenShareTrack && isTrackReference(screenShareTrack) ? (screenShareTrack.publication.trackSid ?? null) : null

  // Tap-to-focus comes from ParticipantTile's FocusToggle (drives the pin
  // through layout context). Screen share auto-pins on top of that.
  const focusTrack = usePinnedTracks(layoutContext)[0]
  const lastAutoPin = useRef<string | null>(null)

  useEffect(() => {
    if (screenShareSid && screenShareTrack) {
      if (lastAutoPin.current !== screenShareSid) {
        layoutContext.pin.dispatch?.({ msg: "set_pin", trackReference: screenShareTrack })
        lastAutoPin.current = screenShareSid
      }
    } else if (lastAutoPin.current) {
      layoutContext.pin.dispatch?.({ msg: "clear_pin" })
      lastAutoPin.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [screenShareSid, layoutContext])

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

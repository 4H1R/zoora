import { isTrackReference, type TrackReferenceOrPlaceholder } from "@livekit/components-react"
import { useEffect, useRef, useState } from "react"

interface TrackDiagProps {
  trackRef: TrackReferenceOrPlaceholder
}

interface Stats {
  hasTrack: boolean
  w: number
  h: number
  readyState: number
  currentTime: number
  advancing: boolean
  paused: boolean
  brightness: number
}

// DEV-ONLY diagnostic. Gated behind ?diag=1 (see stage.tsx).
//
// Shows three renderings of the SAME active track side by side so we can tell,
// on the real device, which paint path works:
//   1. <canvas>  — drawImage readback (known to work)
//   2. RAW <video> — a plain element with NO app/LiveKit CSS. If this paints
//      while the main stage stays black, an app/LiveKit style is the trigger
//      (fix the CSS, skip canvas). If this is ALSO black, the bug is
//      fundamental to <video> compositing here → canvas render is the fix.
// Plus live stats (decode proof).
export function TrackDiag({ trackRef }: TrackDiagProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const rawVideoRef = useRef<HTMLVideoElement>(null)
  const track = isTrackReference(trackRef) ? trackRef.publication.track : undefined
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    const rawVideo = rawVideoRef.current
    if (!canvas || !rawVideo) return
    const ctx = canvas.getContext("2d", { willReadFrequently: true })
    if (!ctx) return

    // Sampling video drives the canvas (kept off the visible tree).
    const video = document.createElement("video")
    video.muted = true
    video.autoplay = true
    video.playsInline = true
    if (track) {
      track.attach(video)
      track.attach(rawVideo) // raw on-DOM element, plain styling
    }
    void video.play().catch(() => {})
    void rawVideo.play().catch(() => {})

    let lastTime = -1
    const id = window.setInterval(() => {
      const w = video.videoWidth
      const h = video.videoHeight
      let brightness = 0
      if (w && h) {
        canvas.width = 160
        canvas.height = Math.max(1, Math.round((160 * h) / w))
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height)
        try {
          const data = ctx.getImageData(0, 0, canvas.width, canvas.height).data
          let sum = 0
          for (let i = 0; i < data.length; i += 4) sum += data[i] + data[i + 1] + data[i + 2]
          brightness = Math.round(sum / (data.length / 4) / 3)
        } catch {
          brightness = -1 // tainted / readback blocked
        }
      }
      const advancing = video.currentTime !== lastTime
      lastTime = video.currentTime
      setStats({
        hasTrack: !!track,
        w,
        h,
        readyState: video.readyState,
        currentTime: Math.round(video.currentTime * 100) / 100,
        advancing,
        paused: video.paused,
        brightness,
      })
    }, 500)

    return () => {
      window.clearInterval(id)
      if (track) {
        track.detach(video)
        track.detach(rawVideo)
      }
      video.srcObject = null
    }
  }, [track])

  return (
    <div className="absolute end-2 top-2 z-50 flex flex-col gap-1 rounded-lg bg-black/80 p-2 font-mono text-[10px] leading-tight text-lime-300">
      <div className="flex gap-1">
        <div className="flex flex-col items-center gap-0.5">
          <span>canvas</span>
          <canvas ref={canvasRef} className="w-24 rounded border border-lime-500/40 bg-black" />
        </div>
        <div className="flex flex-col items-center gap-0.5">
          <span>raw video</span>
          {/* deliberately NO zoora/livekit classes — plain element */}
          <video
            ref={rawVideoRef}
            muted
            autoPlay
            playsInline
            className="w-24 rounded border border-fuchsia-500/40 bg-black"
            style={{ width: "6rem", height: "auto", objectFit: "contain" }}
          />
        </div>
      </div>
      {stats ? (
        <pre className="whitespace-pre-wrap">
{`track:  ${stats.hasTrack}
size:   ${stats.w}x${stats.h}
rstate: ${stats.readyState}
ctime:  ${stats.currentTime}
moving: ${stats.advancing}
paused: ${stats.paused}
bright: ${stats.brightness}`}
        </pre>
      ) : (
        <span>sampling…</span>
      )}
    </div>
  )
}

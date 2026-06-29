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
// Attaches the active track to an offscreen <video>, samples it every 500ms,
// paints it to a small visible <canvas>, and overlays live stats. Tells us on
// the real device whether frames DECODE (w/h>0, readyState 4, currentTime
// advancing, canvas brightness>0 = paint/compositing bug) or DON'T decode
// (everything 0/frozen = codec/delivery problem). The corner canvas is the
// visual tell: if it shows real content while the main <video> is black, the
// fix is canvas rendering; if it is also black, decode/delivery is broken.
export function TrackDiag({ trackRef }: TrackDiagProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const track = isTrackReference(trackRef) ? trackRef.publication.track : undefined
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext("2d", { willReadFrequently: true })
    if (!ctx) return

    const video = document.createElement("video")
    video.muted = true
    video.autoplay = true
    video.playsInline = true
    if (track) track.attach(video)
    void video.play().catch(() => {})

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
      if (track) track.detach(video)
      video.srcObject = null
    }
  }, [track])

  return (
    <div className="absolute end-2 top-2 z-50 flex flex-col gap-1 rounded-lg bg-black/80 p-2 font-mono text-[10px] leading-tight text-lime-300">
      <canvas ref={canvasRef} className="w-40 rounded border border-lime-500/40 bg-black" />
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

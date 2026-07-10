import { AlertCircleIcon, Maximize2Icon, Minimize2Icon, PauseIcon, PlayIcon, XIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { ProgressRing } from "../attachment-progress-ring"
import { DownloadButton, RatePill, SeekLine } from "./controls"
import { formatMediaTime } from "./playback"
import { useMediaPlayback } from "./use-media-playback"

export interface VideoAttachmentProps {
  /** Playable source — presigned download URL or a local blob URL. */
  src?: string
  name: string
  /** Upload-in-progress state for optimistic bubbles. */
  uploading?: boolean
  progress?: number
  errored?: boolean
  onCancel?: () => void
}

/**
 * Inline Telegram-style video: a rounded dark frame with a big center play
 * button, and custom overlay controls (scrub line, time, global speed pill,
 * download, fullscreen). Controls fade out while playing and return on hover
 * or pause; clicking the frame toggles playback.
 */
export function VideoAttachment({
  src,
  name,
  uploading = false,
  progress = 0,
  errored = false,
  onCancel,
}: VideoAttachmentProps) {
  const { t } = useTranslation()
  const containerRef = useRef<HTMLDivElement>(null)
  const videoRef = useRef<HTMLVideoElement>(null)
  const playback = useMediaPlayback(videoRef)
  const [fullscreen, setFullscreen] = useState(false)

  const started = playback.playing || playback.currentTime > 0
  const fraction = playback.duration > 0 ? Math.min(1, playback.currentTime / playback.duration) : 0

  useEffect(() => {
    const onChange = () => setFullscreen(document.fullscreenElement === containerRef.current)
    document.addEventListener("fullscreenchange", onChange)
    return () => document.removeEventListener("fullscreenchange", onChange)
  }, [])

  function toggleFullscreen() {
    if (document.fullscreenElement) void document.exitFullscreen()
    else void containerRef.current?.requestFullscreen()
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        "group/video relative w-full overflow-hidden rounded-xl bg-black",
        fullscreen && "flex items-center justify-center rounded-none"
      )}
    >
      {src ? (
        <video
          ref={videoRef}
          src={src}
          preload="metadata"
          playsInline
          // The bubble-wide context menu opens on `mousedown`; tapping the frame
          // toggles playback and must not also open the menu.
          onMouseDown={(e) => e.stopPropagation()}
          onClick={(e) => {
            e.stopPropagation()
            if (!uploading) playback.toggle()
          }}
          className={cn("w-full cursor-pointer object-contain", fullscreen ? "h-full max-h-full" : "max-h-80")}
        />
      ) : (
        <div className="bg-muted aspect-video w-full animate-pulse" />
      )}

      {/* Center play button — pauses collapse back to it. */}
      {src && !uploading && !errored && !playback.playing && (
        <button
          type="button"
          onMouseDown={(e) => e.stopPropagation()}
          onClick={(e) => {
            e.stopPropagation()
            playback.toggle()
          }}
          aria-label={t("conversations.player.play")}
          className="absolute inset-0 m-auto flex size-14 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur-sm transition hover:bg-black/65 active:scale-95"
        >
          <PlayIcon className="size-6 translate-x-px fill-current" />
        </button>
      )}

      {/* Upload overlay: dim + progress ring + cancel — mirrors image cells. */}
      {uploading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/40">
          <ProgressRing value={progress} className="size-11" />
          <button
            type="button"
            onClick={onCancel}
            aria-label={t("conversations.attachments.cancel")}
            className="absolute inset-0 m-auto flex size-11 items-center justify-center text-white"
          >
            <XIcon className="size-4" />
          </button>
        </div>
      )}

      {errored && (
        <div className="bg-destructive/70 absolute inset-0 flex items-center justify-center">
          <AlertCircleIcon className="size-6 text-white" />
        </div>
      )}

      {/* Bottom control bar over a soft gradient. */}
      {src && !uploading && !errored && (
        <div
          className={cn(
            "absolute inset-x-0 bottom-0 flex flex-col bg-gradient-to-t from-black/70 to-transparent px-2.5 pt-8 pb-1 text-white transition-opacity duration-200",
            playback.playing && "opacity-0 group-hover/video:opacity-100 focus-within:opacity-100",
            !playback.playing && !started && "opacity-0 group-hover/video:opacity-100 focus-within:opacity-100"
          )}
        >
          <SeekLine progress={fraction} tone="overlay" onSeek={playback.seekTo} />
          <div className="flex items-center gap-2">
            <button
              type="button"
              onMouseDown={(e) => e.stopPropagation()}
              onClick={(e) => {
                e.stopPropagation()
                playback.toggle()
              }}
              aria-label={playback.playing ? t("conversations.player.pause") : t("conversations.player.play")}
              className="flex size-6 items-center justify-center rounded-md transition hover:bg-white/15"
            >
              {playback.playing ? (
                <PauseIcon className="size-3.5 fill-current" />
              ) : (
                <PlayIcon className="size-3.5 translate-x-px fill-current" />
              )}
            </button>
            <span className="text-xs leading-none tabular-nums">
              {formatMediaTime(playback.currentTime)} / {formatMediaTime(playback.duration)}
            </span>
            <span className="ms-auto" />
            <RatePill tone="overlay" />
            <DownloadButton url={src} name={name} tone="overlay" />
            <button
              type="button"
              onMouseDown={(e) => e.stopPropagation()}
              onClick={(e) => {
                e.stopPropagation()
                toggleFullscreen()
              }}
              aria-label={fullscreen ? t("conversations.player.exitFullscreen") : t("conversations.player.fullscreen")}
              className="flex size-6 items-center justify-center rounded-md transition hover:bg-white/15"
            >
              {fullscreen ? <Minimize2Icon className="size-3.5" /> : <Maximize2Icon className="size-3.5" />}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

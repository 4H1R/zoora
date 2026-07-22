import { useQuery } from "@tanstack/react-query"
import { AlertCircleIcon, XIcon } from "lucide-react"
import { useRef } from "react"
import { useTranslation } from "react-i18next"

import { formatBytes } from "@/components/org/files/utils"
import { cn } from "@/lib/utils"

import { ProgressRing } from "../attachment-progress-ring"
import { DownloadButton, PlayPauseButton, RatePill, SeekLine, WaveformBars } from "./controls"
import { formatMediaTime } from "./playback"
import { useMediaPlayback } from "./use-media-playback"
import { extractPeaks, isVoiceName, syntheticPeaks } from "./waveform"

export interface AudioAttachmentProps {
  /** Playable source — presigned download URL or a local blob URL. */
  src?: string
  name: string
  size?: number
  isOwn: boolean
  /** Stable id (media id / local id) seeding peak caching + the fallback shape. */
  seed: string
  /** Upload-in-progress state for optimistic bubbles. */
  uploading?: boolean
  progress?: number
  errored?: boolean
  onCancel?: () => void
}

/**
 * Real peaks + exact duration for a voice note, decoded from the fetched bytes.
 * Presigned URLs may lack CORS headers and exotic codecs may not decode — both
 * fall back to a deterministic synthetic waveform (and no duration, leaving the
 * element's own probe to supply it) so the bubble never loses its shape.
 */
function useVoicePeaks(src: string | undefined, seed: string, enabled: boolean) {
  return useQuery({
    queryKey: ["media", "peaks", seed],
    queryFn: async (): Promise<{ peaks: number[]; duration: number }> => {
      try {
        const blob = await (await fetch(src!)).blob()
        return await extractPeaks(blob)
      } catch {
        return { peaks: syntheticPeaks(seed), duration: 0 }
      }
    },
    enabled: enabled && !!src,
    staleTime: Infinity,
  })
}

/**
 * Telegram-style audio bubble. Two layouts by kind:
 * - voice note (`voice-*` recordings) — play circle + scrubbable waveform
 * - music / any other audio — play circle + title + slim seek line
 * Both share the global speed pill and a download affordance. While the
 * optimistic upload is in flight the play circle hosts the progress ring +
 * cancel, and the source is the local blob so playback works immediately.
 */
export function AudioAttachment({
  src,
  name,
  size,
  isOwn,
  seed,
  uploading = false,
  progress = 0,
  errored = false,
  onCancel,
}: AudioAttachmentProps) {
  const { t } = useTranslation()
  const voice = isVoiceName(name)
  const tone = isOwn ? "own" : "accent"

  const { data: decoded } = useVoicePeaks(src, seed, voice)
  const peaks = decoded?.peaks

  const audioRef = useRef<HTMLAudioElement>(null)
  const playback = useMediaPlayback(audioRef, src, decoded?.duration)
  const started = playback.playing || playback.currentTime > 0
  const fraction = playback.duration > 0 ? Math.min(1, playback.currentTime / playback.duration) : 0

  const timeLabel = started
    ? `${formatMediaTime(playback.currentTime)} / ${formatMediaTime(playback.duration)}`
    : playback.duration > 0
      ? formatMediaTime(playback.duration)
      : size !== undefined
        ? formatBytes(size)
        : ""

  const mutedText = isOwn ? "text-primary-foreground/70" : "text-muted-foreground"

  return (
    <div className={cn("flex max-w-full items-center gap-2.5 px-1 py-1", voice ? "w-64" : "w-72")}>
      <audio ref={audioRef} src={src} preload="metadata" className="hidden" />

      {uploading || errored ? (
        <button
          type="button"
          onClick={onCancel}
          disabled={!onCancel}
          aria-label={t("conversations.attachments.cancel")}
          className={cn(
            "relative flex size-10 shrink-0 items-center justify-center rounded-full transition",
            isOwn ? "bg-primary-foreground/20 text-primary-foreground" : "bg-primary text-primary-foreground"
          )}
        >
          {errored ? (
            <AlertCircleIcon className="size-4" />
          ) : (
            <>
              <ProgressRing value={progress} className="absolute inset-0 size-10" />
              <XIcon className="size-3.5" />
            </>
          )}
        </button>
      ) : (
        <PlayPauseButton playing={playback.playing} onToggle={playback.toggle} tone={tone} disabled={!src} />
      )}

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        {voice ? (
          <WaveformBars
            peaks={peaks ?? syntheticPeaks(seed)}
            pending={!peaks}
            progress={fraction}
            tone={tone}
            onSeek={src && !uploading ? playback.seekTo : undefined}
          />
        ) : (
          <>
            <div className="flex items-center gap-2">
              <span className="truncate text-sm font-medium">{name}</span>
              <DownloadButton url={uploading ? undefined : src} name={name} tone={tone} className="ms-auto" />
            </div>
            <SeekLine progress={fraction} tone={tone} onSeek={src && !uploading ? playback.seekTo : undefined} />
          </>
        )}

        <div className={cn("flex items-center gap-1.5 text-xs leading-none tabular-nums", mutedText)}>
          <span className="truncate">{timeLabel}</span>
          {started && <RatePill tone={tone} />}
          {voice && <DownloadButton url={uploading ? undefined : src} name={name} tone={tone} className="ms-auto" />}
        </div>
      </div>
    </div>
  )
}

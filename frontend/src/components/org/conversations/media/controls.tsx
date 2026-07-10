import { DownloadIcon, PauseIcon, PlayIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { useMediaSettings } from "./playback"

/**
 * Color tone for the shared player controls:
 * - `own`     — inside the signed-in user's primary-colored bubble
 * - `accent`  — inside a received (muted) bubble or the composer
 * - `overlay` — on top of video, over a dark gradient
 */
export type ControlTone = "own" | "accent" | "overlay"

const FILL: Record<ControlTone, string> = {
  own: "bg-primary-foreground",
  accent: "bg-primary",
  overlay: "bg-white",
}
const TRACK: Record<ControlTone, string> = {
  own: "bg-primary-foreground/30",
  accent: "bg-primary/25",
  overlay: "bg-white/30",
}

/** Pointer-scrub handlers shared by the waveform and the seek line (RTL-aware). */
function seekHandlers(onSeek?: (fraction: number) => void) {
  if (!onSeek) return {}

  const fractionAt = (el: HTMLElement, clientX: number) => {
    const rect = el.getBoundingClientRect()
    const raw = Math.min(1, Math.max(0, (clientX - rect.left) / Math.max(1, rect.width)))
    return getComputedStyle(el).direction === "rtl" ? 1 - raw : raw
  }

  return {
    onPointerDown: (e: React.PointerEvent<HTMLDivElement>) => {
      e.currentTarget.setPointerCapture(e.pointerId)
      onSeek(fractionAt(e.currentTarget, e.clientX))
    },
    onPointerMove: (e: React.PointerEvent<HTMLDivElement>) => {
      if (!e.currentTarget.hasPointerCapture(e.pointerId)) return
      onSeek(fractionAt(e.currentTarget, e.clientX))
    },
  }
}

function sliderA11y(label: string, progress: number, onSeek?: (fraction: number) => void) {
  if (!onSeek) return {}
  return {
    role: "slider" as const,
    tabIndex: 0,
    "aria-label": label,
    "aria-valuemin": 0,
    "aria-valuemax": 100,
    "aria-valuenow": Math.round(progress * 100),
    onKeyDown: (e: React.KeyboardEvent) => {
      if (e.key === "ArrowRight" || e.key === "ArrowUp") {
        e.preventDefault()
        onSeek(Math.min(1, progress + 0.05))
      }
      if (e.key === "ArrowLeft" || e.key === "ArrowDown") {
        e.preventDefault()
        onSeek(Math.max(0, progress - 0.05))
      }
    },
  }
}

/**
 * Telegram-style voice waveform: equal-width rounded bars whose heights follow
 * the peaks; the played portion is fully saturated. The whole strip is a
 * scrub surface.
 */
export function WaveformBars({
  peaks,
  progress,
  tone,
  onSeek,
  pending = false,
  className,
}: {
  peaks: number[]
  /** 0..1 played fraction. */
  progress: number
  tone: ControlTone
  onSeek?: (fraction: number) => void
  /** True while real peaks are still decoding — pulses the placeholder shape. */
  pending?: boolean
  className?: string
}) {
  const { t } = useTranslation()
  const activeCount = Math.round(progress * peaks.length)

  return (
    <div
      {...seekHandlers(onSeek)}
      {...sliderA11y(t("conversations.player.seek"), progress, onSeek)}
      className={cn(
        "flex h-7 w-full touch-none items-center gap-0.5",
        onSeek && "cursor-pointer",
        pending && "animate-pulse",
        className
      )}
    >
      {peaks.map((peak, i) => (
        <span
          key={i}
          style={{ height: `${Math.round(Math.max(0.12, peak) * 100)}%` }}
          className={cn(
            "min-h-1 flex-1 rounded-full transition-colors duration-150",
            i < activeCount ? FILL[tone] : TRACK[tone]
          )}
        />
      ))}
    </div>
  )
}

/** A slim scrubbable progress line (music + video). */
export function SeekLine({
  progress,
  tone,
  onSeek,
  className,
}: {
  progress: number
  tone: ControlTone
  onSeek?: (fraction: number) => void
  className?: string
}) {
  const { t } = useTranslation()

  return (
    <div
      {...seekHandlers(onSeek)}
      {...sliderA11y(t("conversations.player.seek"), progress, onSeek)}
      className={cn("group/seek flex h-4 w-full touch-none items-center", onSeek && "cursor-pointer", className)}
    >
      <div
        className={cn(
          "relative h-1 w-full overflow-hidden rounded-full transition-all group-hover/seek:h-1.5",
          TRACK[tone]
        )}
      >
        <div
          className={cn("absolute inset-y-0 start-0 rounded-full", FILL[tone])}
          style={{ width: `${Math.min(100, Math.max(0, progress * 100))}%` }}
        />
      </div>
    </div>
  )
}

/** The tappable `1× / 1.5× / 2×` speed pill — one global setting for all players. */
export function RatePill({ tone, className }: { tone: ControlTone; className?: string }) {
  const { t } = useTranslation()
  const rate = useMediaSettings((s) => s.rate)
  const cycleRate = useMediaSettings((s) => s.cycleRate)

  const tones: Record<ControlTone, string> = {
    own: "bg-primary-foreground/20 text-primary-foreground hover:bg-primary-foreground/30",
    accent: "bg-primary/10 text-primary hover:bg-primary/20",
    overlay: "bg-white/20 text-white hover:bg-white/30",
  }

  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        cycleRate()
      }}
      aria-label={t("conversations.player.speed")}
      className={cn(
        "rounded-full px-1.5 py-0.5 text-[10px] leading-none font-semibold tabular-nums transition active:scale-95",
        tones[tone],
        className
      )}
    >
      {rate}×
    </button>
  )
}

/** Round Telegram-style play/pause button (solid triangle / bars). */
export function PlayPauseButton({
  playing,
  onToggle,
  tone,
  disabled = false,
  size = "lg",
  className,
}: {
  playing: boolean
  onToggle: () => void
  tone: ControlTone
  disabled?: boolean
  size?: "sm" | "lg"
  className?: string
}) {
  const { t } = useTranslation()

  const tones: Record<ControlTone, string> = {
    own: "bg-primary-foreground/20 text-primary-foreground hover:bg-primary-foreground/30",
    accent: "bg-primary text-primary-foreground hover:bg-primary/90",
    overlay: "text-white hover:bg-white/15",
  }

  return (
    <button
      type="button"
      onClick={onToggle}
      disabled={disabled}
      aria-label={playing ? t("conversations.player.pause") : t("conversations.player.play")}
      className={cn(
        "flex shrink-0 items-center justify-center rounded-full transition active:scale-95 disabled:opacity-50",
        size === "lg" ? "size-10" : "size-7",
        tones[tone],
        className
      )}
    >
      {playing ? (
        <PauseIcon className={cn("fill-current", size === "lg" ? "size-4" : "size-3.5")} />
      ) : (
        <PlayIcon className={cn("translate-x-px fill-current", size === "lg" ? "size-4" : "size-3.5")} />
      )}
    </button>
  )
}

/** Small download affordance used by audio/video players. */
export function DownloadButton({
  url,
  name,
  tone,
  className,
}: {
  url?: string
  name: string
  tone: ControlTone
  className?: string
}) {
  const { t } = useTranslation()
  if (!url) return null

  const tones: Record<ControlTone, string> = {
    own: "text-primary-foreground/70 hover:text-primary-foreground",
    accent: "text-muted-foreground hover:text-foreground",
    overlay: "text-white/80 hover:text-white",
  }

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      download={name}
      onClick={(e) => e.stopPropagation()}
      aria-label={t("conversations.player.download")}
      className={cn("shrink-0 transition", tones[tone], className)}
    >
      <DownloadIcon className="size-3.5" />
    </a>
  )
}

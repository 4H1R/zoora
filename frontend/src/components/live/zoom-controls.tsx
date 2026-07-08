import { RotateCcw, ZoomIn, ZoomOut } from "lucide-react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

export const MIN_ZOOM = 1
export const MAX_ZOOM = 3
export const ZOOM_STEP = 0.5

// Clamp to [MIN, MAX] and round to whole percents so float drift from repeated
// +/- steps can't leave the level at e.g. 2.4999999×.
export const clampZoom = (z: number) => Math.round(Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, z)) * 100) / 100

// Floating zoom pill shared by the video stage and the slides stage. Solid bg,
// no backdrop-blur: it floats over a <video>/<canvas>, and a backdrop-filter
// over those paints black on some GPUs.
export function ZoomControls({
  zoom,
  onZoom,
  onReset,
  className,
}: {
  zoom: number
  onZoom: (next: number) => void
  onReset: () => void
  className?: string
}) {
  const { t } = useTranslation()
  const zoomed = zoom > MIN_ZOOM

  return (
    <div
      className={cn(
        "absolute bottom-4 start-4 z-20 flex flex-col items-center overflow-hidden rounded-full bg-black/70 text-zinc-100 shadow-lg",
        className
      )}
    >
      <ZoomButton
        icon={ZoomIn}
        label={t("liveRoom.zoom.in")}
        disabled={zoom >= MAX_ZOOM}
        onClick={() => onZoom(zoom + ZOOM_STEP)}
      />
      {zoomed && (
        <span className="px-1 text-[11px] font-semibold tabular-nums" aria-hidden>
          {Math.round(zoom * 100)}%
        </span>
      )}
      <ZoomButton
        icon={ZoomOut}
        label={t("liveRoom.zoom.out")}
        disabled={zoom <= MIN_ZOOM}
        onClick={() => onZoom(zoom - ZOOM_STEP)}
      />
      {zoomed && <ZoomButton icon={RotateCcw} label={t("liveRoom.zoom.reset")} onClick={onReset} />}
    </div>
  )
}

function ZoomButton({
  icon: Icon,
  label,
  disabled,
  onClick,
}: {
  icon: typeof ZoomIn
  label: string
  disabled?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      title={label}
      className="flex size-9 items-center justify-center transition-colors hover:bg-white/15 disabled:pointer-events-none disabled:opacity-35"
    >
      <Icon className="size-4" />
    </button>
  )
}

import { cn } from "@/lib/utils"

interface ProgressRingProps {
  /** 0..1 completion fraction. */
  value: number
  className?: string
}

// SVG user-unit geometry (not CSS px — the element is sized by `className`).
const RADIUS = 16
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

/**
 * A circular determinate progress ring, drawn in SVG user units and sized by
 * the caller's `className`. A faint track sits under a bright accent arc that
 * sweeps clockwise from the top as `value` climbs, with a soft transition so
 * incremental upload ticks glide rather than jump.
 */
export function ProgressRing({ value, className }: ProgressRingProps) {
  const clamped = Math.min(1, Math.max(0, value))
  const offset = CIRCUMFERENCE * (1 - clamped)

  return (
    <svg viewBox="0 0 40 40" className={cn("size-10 -rotate-90", className)} role="progressbar" aria-valuenow={Math.round(clamped * 100)}>
      <circle cx="20" cy="20" r={RADIUS} fill="none" strokeWidth="3" className="stroke-white/25" />
      <circle
        cx="20"
        cy="20"
        r={RADIUS}
        fill="none"
        strokeWidth="3"
        strokeLinecap="round"
        strokeDasharray={CIRCUMFERENCE}
        strokeDashoffset={offset}
        className="stroke-white transition-[stroke-dashoffset] duration-300 ease-out"
      />
    </svg>
  )
}

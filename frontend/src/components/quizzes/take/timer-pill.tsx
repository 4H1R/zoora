import { cn } from "@/lib/utils"

import { formatClock } from "./utils"

interface TimerPillProps {
  remainingSeconds: number
  /** Full exam window in seconds — drives the depletion gauge. */
  totalSeconds: number
}

// Geometry for the circular depletion gauge. r chosen so the 24px viewBox has a
// comfortable 2.5px stroke without clipping.
const GAUGE_R = 9
const GAUGE_C = 2 * Math.PI * GAUGE_R

export function TimerPill({ remainingSeconds, totalSeconds }: TimerPillProps) {
  const danger = remainingSeconds <= 30
  const warn = remainingSeconds <= 60 && !danger
  const fraction = Math.max(0, Math.min(1, totalSeconds > 0 ? remainingSeconds / totalSeconds : 0))

  return (
    <div
      role="timer"
      aria-live="off"
      className={cn(
        "text-foreground ring-foreground/12 bg-foreground/[0.03] inline-flex items-center gap-2.5 rounded-2xl py-1.5 ps-2 pe-3.5 ring-1 transition-colors",
        warn && "text-amber-600 ring-amber-500/40 bg-amber-500/[0.06] dark:text-amber-400",
        danger && "text-destructive ring-destructive/50 bg-destructive/[0.07] animate-pulse",
      )}
    >
      <span className="relative inline-flex size-6 items-center justify-center">
        <svg viewBox="0 0 24 24" className="size-6 -rotate-90" aria-hidden>
          <circle cx="12" cy="12" r={GAUGE_R} fill="none" stroke="currentColor" strokeOpacity={0.15} strokeWidth={2.5} />
          <circle
            cx="12"
            cy="12"
            r={GAUGE_R}
            fill="none"
            stroke="currentColor"
            strokeWidth={2.5}
            strokeLinecap="round"
            strokeDasharray={GAUGE_C}
            strokeDashoffset={GAUGE_C * (1 - fraction)}
            className="transition-[stroke-dashoffset] duration-1000 ease-linear"
          />
        </svg>
        <span className="bg-current absolute size-1 rounded-full" />
      </span>
      <span className="font-mono text-base font-semibold tabular-nums md:text-lg">{formatClock(remainingSeconds)}</span>
    </div>
  )
}

import { ClockIcon } from "lucide-react"

import { cn } from "@/lib/utils"

import { formatClock } from "./utils"

interface TimerPillProps {
  remainingSeconds: number
}

export function TimerPill({ remainingSeconds }: TimerPillProps) {
  const danger = remainingSeconds <= 30
  const warn = remainingSeconds <= 60 && !danger
  return (
    <div
      role="timer"
      aria-live="off"
      className={cn(
        "ring-foreground/15 inline-flex items-center gap-2 rounded-xl px-3 py-1.5 font-mono text-base tabular-nums ring-1 transition-colors md:text-lg",
        warn && "ring-amber-500/50 text-amber-600 dark:text-amber-400",
        danger && "ring-destructive/60 text-destructive animate-pulse",
      )}
    >
      <ClockIcon className="size-4" />
      {formatClock(remainingSeconds)}
    </div>
  )
}

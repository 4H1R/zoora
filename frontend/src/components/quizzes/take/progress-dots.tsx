import type { AnswerState } from "./types"

import { cn } from "@/lib/utils"

import { hasAnswer } from "./utils"

interface ProgressRailProps {
  order: string[]
  index: number
  answers: Record<string, AnswerState>
}

// A segmented rail — one tick per question. The active tick widens and lifts with
// a soft glow; answered ticks read solid; untouched ones stay faint. Reads like a
// filmstrip so a student can gauge position and coverage at a glance.
export function ProgressRail({ order, index, answers }: ProgressRailProps) {
  return (
    <div className="flex items-center gap-1" aria-hidden>
      {order.map((qid, i) => {
        const isCurrent = i === index
        const isAnswered = hasAnswer(answers[qid])
        return (
          <span
            key={qid}
            className={cn(
              "h-1.5 rounded-full transition-all duration-300 ease-out",
              isCurrent
                ? "bg-primary shadow-primary/40 w-6 shadow-[0_0_0_1px] md:w-7"
                : isAnswered
                  ? "bg-foreground/55 w-1.5"
                  : "bg-foreground/15 w-1.5",
            )}
          />
        )
      })}
    </div>
  )
}

// Back-compat alias — the play area still imports ProgressDots.
export { ProgressRail as ProgressDots }

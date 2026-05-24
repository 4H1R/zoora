import { cn } from "@/lib/utils"

import type { AnswerState } from "./types"
import { hasAnswer } from "./utils"

interface ProgressDotsProps {
  order: string[]
  index: number
  answers: Record<string, AnswerState>
}

export function ProgressDots({ order, index, answers }: ProgressDotsProps) {
  return (
    <div className="flex items-center gap-1" aria-hidden>
      {order.map((qid, i) => (
        <span
          key={qid}
          className={cn(
            "size-1.5 rounded-full transition-all",
            i === index ? "bg-foreground w-4" : hasAnswer(answers[qid]) ? "bg-foreground/70" : "bg-foreground/20",
          )}
        />
      ))}
    </div>
  )
}

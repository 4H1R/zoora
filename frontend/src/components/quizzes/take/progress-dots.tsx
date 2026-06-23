import type { AnswerState } from "./types"

import { cn } from "@/lib/utils"

import { hasAnswer } from "./utils"

interface ProgressDotsProps {
  order: string[]
  index: number
  answers: Record<string, AnswerState>
}

export function ProgressDots({ order, index, answers }: ProgressDotsProps) {
  const getDotClass = (qid: string, i: number) => {
    if (i === index) return "bg-foreground w-4"
    if (hasAnswer(answers[qid])) return "bg-foreground/70"
    return "bg-foreground/20"
  }

  return (
    <div className="flex items-center gap-1" aria-hidden>
      {order.map((qid, i) => (
        <span key={qid} className={cn("size-1.5 rounded-full transition-all", getDotClass(qid, i))} />
      ))}
    </div>
  )
}

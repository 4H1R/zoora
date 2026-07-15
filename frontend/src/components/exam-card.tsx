import type {
  GithubCom4H1RZooraInternalDomainMyExam as MyExam,
  GithubCom4H1RZooraInternalDomainMyExamState as MyExamState,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Eyebrow } from "@/components/eyebrow"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useCanAny } from "@/lib/access"
import { formatSessionDate } from "@/lib/session-status"

export function examStateBadgeVariant(state: MyExamState | undefined) {
  switch (state) {
    case "open":
      return "default" as const
    case "graded":
      return "secondary" as const
    case "submitted":
      return "outline" as const
    default:
      return "ghost" as const
  }
}

/** Trailing action for an exam row/card: start button, opens-at hint, or score. */
export function ExamAction({ exam }: { exam: MyExam }) {
  const { t, i18n } = useTranslation()
  // Only enrolled takers (Student preset) may start; a viewer would hit a 403.
  const canTake = useCanAny(["quizzes:take"])
  return (
    <div className="flex shrink-0 items-center gap-3">
      {exam.state === "open" && canTake && (
        <Link to="/quiz/$quizId" params={{ quizId: exam.quiz_id! }}>
          <Button size="sm">{t("org.exams.start")}</Button>
        </Link>
      )}
      {exam.state === "upcoming" && exam.room?.started_at && (
        <span className="text-muted-foreground text-xs">
          {t("org.exams.opensAt", { date: formatSessionDate(exam.room.started_at, i18n.language, "short") })}
        </span>
      )}
      {exam.state === "graded" && (
        <Eyebrow className="tracking-normal normal-case">
          {t("org.exams.score", { score: exam.score ?? 0, total: exam.total_score ?? 0 })}
        </Eyebrow>
      )}
    </div>
  )
}

export function ExamCard({ exam }: { exam: MyExam }) {
  const { t } = useTranslation()
  return (
    <Card size="sm" className="flex-row items-center gap-3 p-4">
      <div className="bg-muted text-muted-foreground flex size-10 shrink-0 items-center justify-center rounded-lg [&>svg]:size-5">
        <ClipboardListIcon />
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="truncate text-sm font-medium">{exam.title || "—"}</p>
          <Badge variant={examStateBadgeVariant(exam.state)}>{t(`org.exams.state.${exam.state}`)}</Badge>
        </div>
        <p className="text-muted-foreground mt-0.5 truncate text-xs">
          {exam.class_name || "—"}
          {typeof exam.duration_minutes === "number"
            ? ` · ${t("org.exams.duration", { count: exam.duration_minutes })}`
            : ""}
        </p>
      </div>

      <ExamAction exam={exam} />
    </Card>
  )
}

export function ExamCardSkeleton() {
  return (
    <Card size="sm" className="flex-row items-center gap-3 p-4">
      <Skeleton className="size-10 rounded-lg" />
      <div className="flex flex-1 flex-col gap-2">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-3 w-32" />
      </div>
      <Skeleton className="h-8 w-24" />
    </Card>
  )
}

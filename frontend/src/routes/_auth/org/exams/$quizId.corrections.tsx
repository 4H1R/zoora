import type { RoomWindowStatus } from "@/components/org/exams/room-window"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, ClockIcon, ExternalLinkIcon, TargetIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetQuizzesId } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { roomWindowStatus, surfacedRoom } from "@/components/org/exams/room-window"
import { QuizCorrectionsPanel } from "@/components/org/quizzes/QuizCorrectionsPanel"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useCanSelfOr, useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/exams/$quizId/corrections")({
  head: () => orgHead("org.exams.corrections.title"),
  component: RouteComponent,
})

const STATUS_BADGE_VARIANT: Record<RoomWindowStatus, "default" | "secondary" | "outline" | "ghost"> = {
  in_progress: "default",
  not_started: "outline",
  ended: "secondary",
  not_scheduled: "ghost",
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { quizId } = Route.useParams()
  // Students never hold quizzes:update — they bounce to the dashboard here.
  const allowed = useOrgGuard(["quizzes:update", "quizzes:update_any"])

  const { data, isPending, isError } = useGetQuizzesId(quizId)
  const quiz = (data?.status === 200 && data.data.data) || undefined

  // Managers grade any quiz; teachers only their own.
  const canGrade = useCanSelfOr("quizzes:update", "quizzes:update_any", quiz?.user_id)

  useBreadcrumb([
    { label: t("org.exams.manage.title"), to: "/org/exams" },
    { label: quiz?.title ?? null, loading: !quiz },
  ])

  if (!allowed) return null

  if (isPending) {
    return (
      <div className="flex flex-col gap-8 py-6">
        <Skeleton className="h-4 w-40" />
        <div className="flex flex-col gap-4">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-10 w-2/3" />
          <Skeleton className="h-5 w-96" />
        </div>
        <Skeleton className="h-32 w-full rounded-2xl" />
        <div className="flex flex-col gap-3">
          <Skeleton className="h-24 w-full rounded-2xl" />
          <Skeleton className="h-24 w-full rounded-2xl" />
          <Skeleton className="h-24 w-full rounded-2xl" />
        </div>
      </div>
    )
  }

  if (isError || !quiz || !canGrade) {
    return (
      <div className="flex flex-col items-start gap-4 py-16">
        <h1 className="text-2xl font-semibold tracking-tight">{t("org.exams.corrections.notFound")}</h1>
        <Button variant="outline" render={<Link to="/org/exams" />}>
          <ArrowLeftIcon className="size-4 rtl:rotate-180" />
          {t("org.exams.corrections.back")}
        </Button>
      </div>
    )
  }

  const room = surfacedRoom(quiz)
  const status = roomWindowStatus(quiz)
  const sessionId = room?.class_session_id

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_bottom_left,var(--color-primary)/8%,transparent_55%)]"
      />

      <div className="pt-6">
        {/* When the quiz is tied to a session, back means the session the
            grader came from — not the org-wide exams list. */}
        {sessionId ? (
          <Link
            to="/org/classes/class-sessions/$classSessionId"
            params={{ classSessionId: sessionId }}
            search={{ tab: "quizzes" }}
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
          >
            <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
            {t("org.exams.corrections.backToSession")}
          </Link>
        ) : (
          <Link
            to="/org/exams"
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
          >
            <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
            {t("org.exams.corrections.back")}
          </Link>
        )}
      </div>

      <header className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-3">
          <Eyebrow>{t("org.exams.corrections.title")}</Eyebrow>
          <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{quiz.title || "—"}</h1>
          <div className="text-muted-foreground flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
            {quiz.class?.name && <span>{quiz.class.name}</span>}
            {typeof quiz.duration_minutes === "number" && (
              <span className="inline-flex items-center gap-1.5 tabular-nums">
                <ClockIcon className="size-4" />
                {t("org.exams.duration", { count: quiz.duration_minutes })}
              </span>
            )}
            {typeof quiz.total_score === "number" && (
              <span className="inline-flex items-center gap-1.5 tabular-nums">
                <TargetIcon className="size-4" />
                {t("org.exams.table.totalScore")}: {quiz.total_score}
              </span>
            )}
            <Badge variant={STATUS_BADGE_VARIANT[status]}>{t(`org.exams.roomStatus.${status}`)}</Badge>
            {sessionId && (
              <Link
                to="/org/classes/class-sessions/$classSessionId"
                params={{ classSessionId: sessionId }}
                search={{ tab: "quizzes" }}
                className="hover:text-foreground inline-flex items-center gap-1.5 underline-offset-4 transition-colors hover:underline"
              >
                <ExternalLinkIcon className="size-3.5" />
                {t("org.exams.corrections.openSession")}
              </Link>
            )}
          </div>
        </div>
        <span className="text-muted-foreground hidden font-mono text-[11px] tracking-[0.25em] uppercase md:inline">
          {`// ${quiz.title ?? ""}`}
        </span>
      </header>

      <QuizCorrectionsPanel quiz={quiz} />
    </div>
  )
}

import type {
  GithubCom4H1RZooraInternalDomainMyExam as MyExam,
  GithubCom4H1RZooraInternalDomainMyExamState as MyExamState,
} from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetQuizzesMe } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/$orgId/exams/")({
  head: () => orgHead("org.nav.exams"),
  component: RouteComponent,
})

function examStateBadgeVariant(state: MyExamState | undefined) {
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

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { orgId } = Route.useParams()
  const allowed = useOrgGuard(["quizzes:view", "quizzes:take"])

  const examsQ = useGetQuizzesMe(undefined, { query: { enabled: allowed } })
  const exams: MyExam[] = (examsQ.data?.status === 200 && examsQ.data.data.data?.items) || []
  const loading = examsQ.isPending

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.exams.title")} />

      {loading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} size="sm" className="flex-row items-center gap-3 p-4">
              <Skeleton className="size-10 rounded-lg" />
              <div className="flex flex-1 flex-col gap-2">
                <Skeleton className="h-4 w-48" />
                <Skeleton className="h-3 w-32" />
              </div>
              <Skeleton className="h-8 w-24" />
            </Card>
          ))}
        </div>
      ) : exams.length === 0 ? (
        <Card className="flex flex-col items-center gap-2 px-6 py-12 text-center">
          <div className="bg-muted text-muted-foreground mb-1 flex size-12 items-center justify-center rounded-xl [&>svg]:size-6">
            <ClipboardListIcon />
          </div>
          <p className="text-muted-foreground max-w-sm text-sm">{t("org.exams.empty")}</p>
        </Card>
      ) : (
        <div className="flex flex-col gap-3">
          {exams.map((e) => (
            <Card key={e.quiz_id} size="sm" className="flex-row items-center gap-3 p-4">
              <div className="bg-muted text-muted-foreground flex size-10 shrink-0 items-center justify-center rounded-lg [&>svg]:size-5">
                <ClipboardListIcon />
              </div>

              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <p className="truncate text-sm font-medium">{e.title || "—"}</p>
                  <Badge variant={examStateBadgeVariant(e.state)}>{t(`org.exams.state.${e.state}`)}</Badge>
                </div>
                <p className="text-muted-foreground mt-0.5 truncate text-xs">
                  {e.class_name || "—"}
                  {typeof e.duration_minutes === "number"
                    ? ` · ${t("org.exams.duration", { count: e.duration_minutes })}`
                    : ""}
                </p>
              </div>

              <div className="flex shrink-0 items-center gap-3">
                {e.state === "open" ? (
                  <Link to="/org/$orgId/exams/$quizId/take" params={{ orgId, quizId: e.quiz_id! }}>
                    <Button size="sm">{t("org.exams.start")}</Button>
                  </Link>
                ) : null}
                {e.state === "upcoming" && e.room?.started_at ? (
                  <span className="text-muted-foreground text-xs">
                    {t("org.exams.opensAt", { date: formatSessionDate(e.room.started_at, i18n.language, "short") })}
                  </span>
                ) : null}
                {e.state === "graded" ? (
                  <Eyebrow className="normal-case tracking-normal">
                    {t("org.exams.score", { score: e.score ?? 0, total: e.total_score ?? 0 })}
                  </Eyebrow>
                ) : null}
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}

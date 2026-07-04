import type {
  GithubCom4H1RZooraInternalDomainMyExam as MyExam,
  GithubCom4H1RZooraInternalDomainMyExamState as MyExamState,
} from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ArrowRightIcon, ClipboardListIcon, GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetGradebookMe } from "@/api/gradebook/gradebook"
import { useGetQuizzesMe } from "@/api/quizzes/quizzes"
import { useGetUsersMe } from "@/api/users/users"
import { Eyebrow } from "@/components/eyebrow"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"

import { TileGrid } from "./tile-grid"
import { useDashboardTiles } from "./use-dashboard-tiles"
import { useGreeting } from "./use-greeting"

// Maps exam state to Badge variant using theme tokens only.
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

type LatestGrade = {
  classId: string
  className: string
  columnTitle: string
  value: string
}

export function StudentDashboard() {
  const { t } = useTranslation()
  const tiles = useDashboardTiles()

  const { data: meData } = useGetUsersMe()
  const me = (meData?.status === 200 && meData.data.data) || undefined
  const firstName = (me?.name ?? "").trim().split(/\s+/)[0] || me?.username || ""
  const greeting = useGreeting(firstName)

  const examsQ = useGetQuizzesMe()
  const allExams: MyExam[] = (examsQ.data?.status === 200 && examsQ.data.data.data?.items) || []
  const examsLoading = examsQ.isPending
  const upcomingExams = allExams
    .filter((e) => e.state === "open" || e.state === "upcoming")
    .slice(0, 5)

  const gradesQ = useGetGradebookMe()
  const gradebook = (gradesQ.data?.status === 200 && gradesQ.data.data.data) || undefined
  const gradesLoading = gradesQ.isPending
  const latestGrades: LatestGrade[] = []
  for (const cls of gradebook?.classes ?? []) {
    for (const col of cls.columns ?? []) {
      const value = col.id ? cls.cells?.[col.id] : undefined
      if (value && value.trim()) {
        latestGrades.push({
          classId: cls.class_id ?? "",
          className: cls.class_name ?? "—",
          columnTitle: col.title ?? "—",
          value,
        })
      }
      if (latestGrades.length >= 5) break
    }
    if (latestGrades.length >= 5) break
  }

  return (
    <div className="relative isolate flex flex-col gap-6">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 -top-6 -z-10 h-48 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/8%,transparent_60%)]"
      />

      <div className="flex flex-col gap-1.5">
        <Eyebrow className="text-primary">{t("org.dashboard.overview")}</Eyebrow>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">{greeting}</h1>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card className="gap-0 overflow-hidden p-0">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <Eyebrow>{t("org.portal.widgets.upcomingExams")}</Eyebrow>
            <Link
              to="/org/exams"
              className="text-primary inline-flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-80"
            >
              {t("org.portal.widgets.viewAll")}
              <ArrowRightIcon className="size-3" />
            </Link>
          </div>
          {examsLoading ? (
            <div className="divide-y">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3 px-4 py-3">
                  <Skeleton className="size-9 rounded-lg" />
                  <div className="flex flex-1 flex-col gap-1.5">
                    <Skeleton className="h-4 w-40" />
                    <Skeleton className="h-3 w-24" />
                  </div>
                </div>
              ))}
            </div>
          ) : upcomingExams.length === 0 ? (
            <p className="text-muted-foreground px-4 py-8 text-center text-sm">
              {t("org.portal.widgets.noExams")}
            </p>
          ) : (
            <div className="divide-y">
              {upcomingExams.map((e) => (
                <div key={e.quiz_id} className="flex items-center gap-3 px-4 py-3">
                  <div className="bg-muted text-muted-foreground flex size-9 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
                    <ClipboardListIcon />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{e.title || "—"}</p>
                    <p className="text-muted-foreground truncate text-xs">{e.class_name || "—"}</p>
                  </div>
                  <Badge variant={examStateBadgeVariant(e.state)}>
                    {t(`org.exams.state.${e.state}`)}
                  </Badge>
                  {e.state === "open" && (
                    <Link to="/quiz/$quizId" params={{ quizId: e.quiz_id! }}>
                      <Button size="sm">{t("org.exams.start")}</Button>
                    </Link>
                  )}
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card className="gap-0 overflow-hidden p-0">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <Eyebrow>{t("org.portal.widgets.latestGrades")}</Eyebrow>
            <Link
              to="/org/grades"
              className="text-primary inline-flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-80"
            >
              {t("org.portal.widgets.viewAll")}
              <ArrowRightIcon className="size-3" />
            </Link>
          </div>
          {gradesLoading ? (
            <div className="divide-y">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3 px-4 py-3">
                  <Skeleton className="size-9 rounded-lg" />
                  <div className="flex flex-1 flex-col gap-1.5">
                    <Skeleton className="h-4 w-40" />
                    <Skeleton className="h-3 w-24" />
                  </div>
                </div>
              ))}
            </div>
          ) : latestGrades.length === 0 ? (
            <p className="text-muted-foreground px-4 py-8 text-center text-sm">
              {t("org.portal.widgets.noGrades")}
            </p>
          ) : (
            <div className="divide-y">
              {latestGrades.map((g, i) => (
                <div key={`${g.classId}-${i}`} className="flex items-center gap-3 px-4 py-3">
                  <div className="bg-muted text-muted-foreground flex size-9 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
                    <GraduationCapIcon />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{g.columnTitle}</p>
                    <p className="text-muted-foreground truncate text-xs">{g.className}</p>
                  </div>
                  <span className="text-sm font-semibold tabular-nums">{g.value}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>

      {tiles.length > 0 && <TileGrid tiles={tiles} />}

      {tiles.length === 0 && !examsLoading && !gradesLoading && upcomingExams.length === 0 && latestGrades.length === 0 && (
        <EmptyState
          icon={GraduationCapIcon}
          title={t("org.dashboard.memberEmpty.title")}
          description={t("org.dashboard.memberEmpty.hint")}
        />
      )}
    </div>
  )
}

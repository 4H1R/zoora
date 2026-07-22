import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"
import type { SectionSort } from "@/lib/use-section-list"

import { CheckCheckIcon, CheckSquareIcon, ClipboardListIcon, ClockIcon, HourglassIcon, PencilIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetQuizzesIdSubmissions } from "@/api/quizzes/quizzes"
import { GradeSubmissionDialog } from "@/components/admin/quizzes/corrections/GradeSubmissionDialog"
import { AIGradingControl } from "@/components/org/quizzes/ai-grading-control"
import { SortPicker } from "@/components/data-table/sort-picker"
import { Eyebrow } from "@/components/eyebrow"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { DEFAULT_PAGE_SIZE } from "@/lib/list"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

import { useQuizPermissions } from "./use-quiz-permissions"

const STATUS_FILTERS = ["all", "submitted", "graded", "in_progress"] as const
type StatusFilter = (typeof STATUS_FILTERS)[number]

interface QuizCorrectionsPanelProps {
  quiz?: Quiz
}

// Per-quiz corrections workbench: stat strip, status filter, sort, submission
// rows and the grade dialog. Shared by the session corrections section (which
// adds a quiz picker) and the /org/exams/$quizId/corrections page.
export function QuizCorrectionsPanel({ quiz }: QuizCorrectionsPanelProps) {
  const { t, i18n } = useTranslation()
  const { canEdit } = useQuizPermissions()

  const [status, setStatus] = useState<StatusFilter>("all")
  const [sort, setSort] = useState<SectionSort | undefined>(undefined)
  const [page, setPage] = useState(1)
  const [gradeOpen, setGradeOpen] = useState(false)
  const [active, setActive] = useState<QuizSubmission | null>(null)

  const quizId = quiz?.id
  const quizMaxScore = quiz?.total_score

  const sortOptions: SortOption[] = [
    { id: "submitted_at", label: t("org.session.controls.sortFields.submitted_at") },
    { id: "started_at", label: t("org.session.controls.sortFields.started_at") },
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "total_score", label: t("org.session.controls.sortFields.total_score") },
  ]

  useEffect(() => {
    setPage(1)
  }, [quizId, status, sort?.id, sort?.desc])

  const subsQ = useGetQuizzesIdSubmissions(
    quizId ?? "",
    {
      status: status === "all" ? undefined : status,
      order_by: sort?.id,
      order_dir: sort ? (sort.desc ? "desc" : "asc") : undefined,
      page,
    },
    { query: { enabled: !!quizId && canEdit } }
  )
  const subsData = (subsQ.data?.status === 200 && subsQ.data.data.data) || undefined
  const submissions = subsData?.items ?? []
  const total = subsData?.total ?? 0
  const pageSize = subsData?.page_size ?? DEFAULT_PAGE_SIZE

  const pendingCount = submissions.filter((s) => s.status === "submitted").length
  const gradedCount = submissions.filter((s) => s.status === "graded").length

  const isLoadingSubs = subsQ.isPending && !!quizId

  return (
    <>
      <StatStrip
        total={total}
        pending={pendingCount}
        graded={gradedCount}
        maxScore={quizMaxScore}
        loading={isLoadingSubs}
      />

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0 flex-1">
          <StatusFilterBar
            value={status}
            onChange={setStatus}
            counts={{ all: total, submitted: pendingCount, graded: gradedCount }}
          />
        </div>
        <div className="flex items-end gap-2">
          <AIGradingControl quizId={quizId} disabled={isLoadingSubs} />
          <SortPicker options={sortOptions} value={sort} onChange={setSort} label={t("org.session.controls.sort")} />
        </div>
      </div>

      {isLoadingSubs ? (
        <div className="flex flex-col gap-3">
          <SubmissionRowSkeleton />
          <SubmissionRowSkeleton />
          <SubmissionRowSkeleton />
        </div>
      ) : submissions.length === 0 ? (
        <EmptyState
          icon={CheckSquareIcon}
          title={t("org.session.corrections.noResults")}
          description={t("org.session.corrections.noResultsHint")}
        />
      ) : (
        <>
          <ul className="flex flex-col gap-3">
            {submissions.map((s, i) => (
              <SubmissionRow
                key={s.id ?? i}
                submission={s}
                index={(page - 1) * pageSize + i}
                lang={i18n.language}
                maxScore={quizMaxScore}
                onGrade={() => {
                  setActive(s)
                  setGradeOpen(true)
                }}
              />
            ))}
          </ul>
          <SectionPagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} />
        </>
      )}

      <GradeSubmissionDialog
        open={gradeOpen}
        onOpenChange={(open) => {
          setGradeOpen(open)
          if (!open) setActive(null)
        }}
        submission={active}
        quizId={quizId}
        quizMaxScore={quizMaxScore}
      />
    </>
  )
}

function StatStrip({
  total,
  pending,
  graded,
  maxScore,
  loading,
}: {
  total: number
  pending: number
  graded: number
  maxScore?: number
  loading: boolean
}) {
  const { t } = useTranslation()
  const completion = total > 0 ? Math.round((graded / total) * 100) : 0

  return (
    <div className="bg-card border-border dark:ring-foreground/10 grid grid-cols-2 overflow-hidden rounded-2xl border shadow-sm md:grid-cols-4 dark:border-0 dark:shadow-none dark:ring-1">
      <StatCell
        icon={<ClipboardListIcon className="size-4" />}
        label={t("org.session.corrections.stats.total")}
        value={loading ? null : total}
      />
      <StatCell
        icon={<HourglassIcon className="size-4" />}
        label={t("org.session.corrections.stats.pending")}
        value={loading ? null : pending}
        accent={pending > 0 ? "warn" : "neutral"}
      />
      <StatCell
        icon={<CheckCheckIcon className="size-4" />}
        label={t("org.session.corrections.stats.graded")}
        value={loading ? null : graded}
        accent="success"
        suffix={
          !loading &&
          total > 0 && <span className="text-muted-foreground font-mono text-xs tabular-nums">{completion}%</span>
        }
      />
      <StatCell
        icon={<CheckSquareIcon className="size-4" />}
        label={t("org.session.corrections.stats.maxScore")}
        value={formatScore(maxScore)}
        mono
      />
    </div>
  )
}

function StatCell({
  icon,
  label,
  value,
  accent = "neutral",
  mono = false,
  suffix = null,
}: {
  icon: React.ReactNode
  label: string
  value: number | string | null
  accent?: "neutral" | "success" | "warn"
  mono?: boolean
  suffix?: React.ReactNode
}) {
  const accentClass =
    accent === "success"
      ? "text-emerald-600 dark:text-emerald-400"
      : accent === "warn"
        ? "text-amber-600 dark:text-amber-400"
        : "text-foreground"

  return (
    <div className="border-border flex flex-col gap-2 border-b border-dashed p-5 md:border-s md:border-b-0 md:first:border-s-0">
      <div className="text-muted-foreground flex items-center gap-2">
        {icon}
        <Eyebrow className="text-[10px]">{label}</Eyebrow>
      </div>
      <div className="flex items-baseline gap-2">
        {value === null ? (
          <Skeleton className="h-8 w-14" />
        ) : (
          <span
            className={cn(
              "text-3xl font-semibold tracking-tight tabular-nums",
              mono && "font-mono text-2xl",
              accentClass
            )}
          >
            {value}
          </span>
        )}
        {suffix}
      </div>
    </div>
  )
}

function getCount(s: StatusFilter, counts: { all: number; submitted: number; graded: number }): number | null {
  if (s === "all") return counts.all
  if (s === "submitted") return counts.submitted
  if (s === "graded") return counts.graded
  return null
}

function StatusFilterBar({
  value,
  onChange,
  counts,
}: {
  value: StatusFilter
  onChange: (s: StatusFilter) => void
  counts: { all: number; submitted: number; graded: number }
}) {
  const { t } = useTranslation()
  return (
    <div className="border-border flex flex-wrap items-center gap-1 border-b border-dashed pb-2">
      {STATUS_FILTERS.map((s) => {
        const active = value === s
        const count = getCount(s, counts)
        return (
          <button
            key={s}
            type="button"
            onClick={() => onChange(s)}
            className={cn(
              "relative inline-flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
              active ? "text-foreground" : "text-muted-foreground hover:text-foreground"
            )}
          >
            <span>{t(`org.session.corrections.filter.${s}`)}</span>
            {count != null && count > 0 && (
              <span
                className={cn(
                  "rounded-full px-1.5 py-0.5 font-mono text-[10px] tabular-nums",
                  active ? "bg-foreground text-background" : "bg-muted text-muted-foreground"
                )}
              >
                {count}
              </span>
            )}
            {active && <span aria-hidden className="bg-foreground absolute inset-x-2 -bottom-2 h-px" />}
          </button>
        )
      })}
    </div>
  )
}

function SubmissionRow({
  submission,
  index,
  lang,
  maxScore,
  onGrade,
}: {
  submission: QuizSubmission
  index: number
  lang: string
  maxScore?: number
  onGrade: () => void
}) {
  const { t } = useTranslation()
  const name = submission.user?.name ?? "—"
  const username = submission.user?.username
  const status = submission.status ?? "in_progress"
  const tile = String(index + 1).padStart(2, "0")
  const score = submission.total_score ?? 0
  const submittedStr = submission.submitted_at ? formatSessionDate(submission.submitted_at, lang, "short") : null
  const isInProgress = status === "in_progress"

  const statusTone =
    status === "graded"
      ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
      : status === "submitted"
        ? "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300"
        : "border-foreground/15 text-muted-foreground"

  return (
    <li className="group/row bg-card text-card-foreground border-border hover:border-foreground/25 dark:ring-foreground/10 dark:hover:ring-foreground/30 relative isolate flex flex-col gap-4 overflow-hidden rounded-2xl border p-4 transition-all sm:flex-row sm:items-center sm:gap-6 sm:p-5 dark:border-0 dark:ring-1">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/6%,transparent_60%)] opacity-0 transition-opacity group-hover/row:opacity-100"
      />

      <span className="text-muted-foreground absolute end-4 top-3 font-mono text-[10px] tracking-[0.25em] sm:static sm:end-auto sm:top-auto">
        /{tile}
      </span>

      <div className="flex min-w-0 flex-1 items-center gap-3">
        <div
          className={cn(
            "flex size-11 shrink-0 items-center justify-center rounded-xl text-xs font-semibold text-white",
            getEntityColor(name)
          )}
        >
          {getInitials(name)}
        </div>
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold tracking-tight">{name}</div>
          {username && <div className="text-muted-foreground truncate font-mono text-xs">@{username}</div>}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-2 sm:flex-nowrap">
        <span
          className={cn(
            "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 font-mono text-[10px] tracking-wider uppercase",
            statusTone
          )}
        >
          <StatusDot status={status} />
          {t(`admin.corrections.statuses.${status}`)}
        </span>

        <div className="flex flex-col gap-0.5 sm:min-w-[6rem]">
          <Eyebrow className="text-[9px]">{t("org.session.corrections.row.score")}</Eyebrow>
          <span className="inline-flex items-baseline gap-1 font-mono text-base tabular-nums">
            <span className={cn(status === "graded" ? "text-foreground" : "text-muted-foreground")}>
              {formatScore(score)}
            </span>
            {maxScore != null && maxScore > 0 && (
              <span className="text-muted-foreground text-[10px]">/ {formatScore(maxScore)}</span>
            )}
          </span>
        </div>

        <div className="flex flex-col gap-0.5 sm:min-w-[8rem]">
          <Eyebrow className="text-[9px]">{t("org.session.corrections.row.submitted")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 text-xs">
            {submittedStr ? (
              <>
                <ClockIcon className="size-3" />
                {submittedStr}
              </>
            ) : (
              <span className="text-muted-foreground italic">{t("org.session.corrections.row.notSubmitted")}</span>
            )}
          </span>
        </div>
      </div>

      <Button
        size="sm"
        variant={status === "submitted" ? "default" : "outline"}
        disabled={isInProgress}
        onClick={onGrade}
        title={
          isInProgress ? t("org.session.corrections.row.gradeDisabledHint") : t("org.session.corrections.row.gradeHint")
        }
      >
        <PencilIcon data-icon="inline-start" />
        {status === "graded" ? t("org.session.corrections.row.regrade") : t("org.session.corrections.row.grade")}
      </Button>
    </li>
  )
}

function StatusDot({ status }: { status: string }) {
  const cls =
    status === "graded" ? "bg-emerald-500" : status === "submitted" ? "bg-amber-500" : "bg-muted-foreground/50"
  return <span aria-hidden className={cn("size-1.5 rounded-full", cls)} />
}

function SubmissionRowSkeleton() {
  return (
    <div className="bg-card border-border dark:ring-foreground/10 flex items-center gap-4 rounded-2xl border p-5 shadow-sm dark:border-0 dark:shadow-none dark:ring-1">
      <Skeleton className="size-11 rounded-xl" />
      <div className="flex flex-1 flex-col gap-1.5">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-3 w-24" />
      </div>
      <Skeleton className="h-6 w-20 rounded-full" />
      <Skeleton className="h-6 w-16" />
      <Skeleton className="h-8 w-20" />
    </div>
  )
}

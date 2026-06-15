import type {
  GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom,
  GithubCom4H1RZooraInternalDomainPracticeSubmission as PracticeSubmission,
} from "@/api/model"

import {
  CheckCheckIcon,
  CheckSquareIcon,
  ClipboardListIcon,
  ClockIcon,
  DumbbellIcon,
  HourglassIcon,
  PencilIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetPractices, useGetPracticesIdSubmissions } from "@/api/practices/practices"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

import { PracticeGradeDialog } from "./PracticeGradeDialog"
import { usePracticePermissions } from "./use-practice-permissions"

const STATUS_FILTERS = ["all", "pending", "graded"] as const
type StatusFilter = (typeof STATUS_FILTERS)[number]

const isGraded = (s: PracticeSubmission) => s.score != null

interface PracticeScoresSectionProps {
  classSessionId: string
}

export function PracticeScoresSection({ classSessionId }: PracticeScoresSectionProps) {
  const { t, i18n } = useTranslation()
  const { canView, canGrade } = usePracticePermissions()

  const [practiceId, setPracticeId] = useState<string | undefined>(undefined)
  const [status, setStatus] = useState<StatusFilter>("all")
  const [gradeOpen, setGradeOpen] = useState(false)
  const [active, setActive] = useState<PracticeSubmission | null>(null)

  const practicesQ = useGetPractices(
    { class_session_id: classSessionId },
    { query: { enabled: canView } }
  )
  const practices: PracticeRoom[] =
    (practicesQ.data?.status === 200 && practicesQ.data.data.data?.items) || []

  const effectiveId = practiceId ?? practices[0]?.id
  const selected = practices.find((p) => p.id === effectiveId)
  const maxScore = selected?.max_score

  const subsQ = useGetPracticesIdSubmissions(
    effectiveId ?? "",
    undefined,
    { query: { enabled: !!effectiveId && canGrade } }
  )
  const subsData = (subsQ.data?.status === 200 && subsQ.data.data.data) || undefined
  const allSubmissions = subsData?.items ?? []
  const total = subsData?.total ?? allSubmissions.length

  const gradedCount = allSubmissions.filter(isGraded).length
  const pendingCount = allSubmissions.length - gradedCount

  const submissions = allSubmissions.filter((s) =>
    status === "all" ? true : status === "graded" ? isGraded(s) : !isGraded(s)
  )

  if (!canView || !canGrade) return null

  const isLoadingPractices = practicesQ.isPending
  const isLoadingSubs = subsQ.isPending && !!effectiveId
  const noPractices = !isLoadingPractices && practices.length === 0

  return (
    <section
      id="practice-scores"
      className="relative isolate flex scroll-mt-20 flex-col gap-6 overflow-hidden rounded-3xl"
    >
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_bottom_left,var(--color-primary)/8%,transparent_55%)]"
      />

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{t("org.session.practiceScores.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">
            {t("org.session.practiceScores.title")}
          </h2>
          <p className="text-muted-foreground max-w-xl text-sm leading-relaxed">
            {t("org.session.practiceScores.subtitle")}
          </p>
        </div>
        <span className="text-muted-foreground hidden font-mono text-[11px] tracking-[0.25em] uppercase md:inline">
          {selected ? `// ${selected.title}` : `// ${t("org.session.practiceScores.noneSelected")}`}
        </span>
      </div>

      {noPractices ? (
        <EmptyState />
      ) : (
        <>
          <PracticeSelector
            practices={practices}
            selectedId={effectiveId}
            onSelect={setPracticeId}
            isLoading={isLoadingPractices}
          />

          <StatStrip
            total={total}
            pending={pendingCount}
            graded={gradedCount}
            maxScore={maxScore}
            loading={isLoadingSubs}
          />

          <StatusFilterBar
            value={status}
            onChange={setStatus}
            counts={{ all: allSubmissions.length, pending: pendingCount, graded: gradedCount }}
          />

          {isLoadingSubs ? (
            <div className="flex flex-col gap-3">
              <SubmissionRowSkeleton />
              <SubmissionRowSkeleton />
              <SubmissionRowSkeleton />
            </div>
          ) : submissions.length === 0 ? (
            <NoSubmissions />
          ) : (
            <ul className="flex flex-col gap-3">
              {submissions.map((s, i) => (
                <SubmissionRow
                  key={s.id ?? i}
                  submission={s}
                  index={i}
                  lang={i18n.language}
                  maxScore={maxScore}
                  onGrade={() => {
                    setActive(s)
                    setGradeOpen(true)
                  }}
                />
              ))}
            </ul>
          )}
        </>
      )}

      <PracticeGradeDialog
        open={gradeOpen}
        onOpenChange={(open) => {
          setGradeOpen(open)
          if (!open) setActive(null)
        }}
        submission={active}
        practiceId={effectiveId}
        maxScore={maxScore}
      />
    </section>
  )
}

function EmptyState() {
  const { t } = useTranslation()
  return (
    <div className="bg-card border-border flex flex-col items-center gap-3 rounded-2xl border px-6 py-16 text-center shadow-sm dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10">
      <DumbbellIcon className="text-muted-foreground size-8" />
      <h3 className="text-foreground text-lg font-semibold tracking-tight">
        {t("org.session.practiceScores.emptyTitle")}
      </h3>
      <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
        {t("org.session.practiceScores.emptyHint")}
      </p>
    </div>
  )
}

function NoSubmissions() {
  const { t } = useTranslation()
  return (
    <div className="bg-card border-border flex flex-col items-center gap-2 rounded-2xl border px-6 py-12 text-center shadow-sm dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10">
      <CheckSquareIcon className="text-muted-foreground size-7 opacity-60" />
      <h3 className="text-foreground text-base font-semibold tracking-tight">
        {t("org.session.practiceScores.noResults")}
      </h3>
      <p className="text-muted-foreground max-w-sm text-xs leading-relaxed">
        {t("org.session.practiceScores.noResultsHint")}
      </p>
    </div>
  )
}

function PracticeSelector({
  practices,
  selectedId,
  onSelect,
  isLoading,
}: {
  practices: PracticeRoom[]
  selectedId?: string
  onSelect: (id: string | undefined) => void
  isLoading: boolean
}) {
  const { t } = useTranslation()
  if (isLoading) {
    return (
      <div className="flex flex-wrap gap-2">
        <Skeleton className="h-9 w-32 rounded-full" />
        <Skeleton className="h-9 w-40 rounded-full" />
        <Skeleton className="h-9 w-28 rounded-full" />
      </div>
    )
  }
  return (
    <div className="flex flex-col gap-2">
      <Eyebrow className="text-[10px]">{t("org.session.practiceScores.pick")}</Eyebrow>
      <div className="flex flex-wrap gap-2">
        {practices.map((p, i) => {
          const tile = String(i + 1).padStart(2, "0")
          const active = p.id === selectedId
          return (
            <button
              key={p.id}
              type="button"
              onClick={() => onSelect(p.id)}
              className={cn(
                "group/pill border-border inline-flex items-center gap-2.5 rounded-full border px-4 py-2 text-sm font-medium shadow-sm transition-all dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10",
                active
                  ? "bg-foreground text-background border-foreground dark:ring-foreground"
                  : "bg-card text-foreground hover:-translate-y-0.5 hover:border-foreground/30 dark:hover:ring-foreground/30"
              )}
            >
              <span
                className={cn(
                  "font-mono text-[10px] tracking-[0.25em]",
                  active ? "text-background/70" : "text-muted-foreground"
                )}
              >
                /{tile}
              </span>
              <span className="line-clamp-1 max-w-[16rem]">{p.title ?? "—"}</span>
            </button>
          )
        })}
      </div>
    </div>
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
    <div className="bg-card border-border grid grid-cols-2 overflow-hidden rounded-2xl border shadow-sm dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10 md:grid-cols-4">
      <StatCell
        icon={<ClipboardListIcon className="size-4" />}
        label={t("org.session.practiceScores.stats.total")}
        value={loading ? null : total}
      />
      <StatCell
        icon={<HourglassIcon className="size-4" />}
        label={t("org.session.practiceScores.stats.pending")}
        value={loading ? null : pending}
        accent={pending > 0 ? "warn" : "neutral"}
      />
      <StatCell
        icon={<CheckCheckIcon className="size-4" />}
        label={t("org.session.practiceScores.stats.graded")}
        value={loading ? null : graded}
        accent="success"
        suffix={
          !loading && total > 0 ? (
            <span className="text-muted-foreground font-mono text-xs tabular-nums">{completion}%</span>
          ) : null
        }
      />
      <StatCell
        icon={<CheckSquareIcon className="size-4" />}
        label={t("org.session.practiceScores.stats.maxScore")}
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
    <div className="border-border flex flex-col gap-2 border-b border-dashed p-5 md:border-b-0 md:border-s md:first:border-s-0">
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

function StatusFilterBar({
  value,
  onChange,
  counts,
}: {
  value: StatusFilter
  onChange: (s: StatusFilter) => void
  counts: { all: number; pending: number; graded: number }
}) {
  const { t } = useTranslation()
  return (
    <div className="border-border flex flex-wrap items-center gap-1 border-b border-dashed pb-2">
      {STATUS_FILTERS.map((s) => {
        const active = value === s
        const count = s === "all" ? counts.all : s === "pending" ? counts.pending : counts.graded
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
            <span>{t(`org.session.practiceScores.filter.${s}`)}</span>
            {count > 0 && (
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
  submission: PracticeSubmission
  index: number
  lang: string
  maxScore?: number
  onGrade: () => void
}) {
  const { t } = useTranslation()
  const name = submission.user?.name ?? "—"
  const username = submission.user?.username
  const graded = isGraded(submission)
  const tile = String(index + 1).padStart(2, "0")
  const score = submission.score ?? 0
  const submittedStr = submission.submitted_at
    ? formatSessionDate(submission.submitted_at, lang, "short")
    : null

  const statusTone = graded
    ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
    : "border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300"

  return (
    <li className="group/row bg-card text-card-foreground border-border relative isolate flex flex-col gap-4 overflow-hidden rounded-2xl border p-4 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md hover:border-foreground/25 dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10 dark:hover:ring-foreground/30 sm:flex-row sm:items-center sm:gap-6 sm:p-5">
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
          {username && (
            <div className="text-muted-foreground truncate font-mono text-xs">@{username}</div>
          )}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-2 sm:flex-nowrap">
        <span
          className={cn(
            "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider",
            statusTone
          )}
        >
          <span
            aria-hidden
            className={cn("size-1.5 rounded-full", graded ? "bg-emerald-500" : "bg-amber-500")}
          />
          {t(`org.session.practiceScores.statuses.${graded ? "graded" : "pending"}`)}
        </span>

        <div className="flex flex-col gap-0.5 sm:min-w-[6rem]">
          <Eyebrow className="text-[9px]">{t("org.session.practiceScores.row.score")}</Eyebrow>
          <span className="inline-flex items-baseline gap-1 font-mono text-base tabular-nums">
            <span className={cn(graded ? "text-foreground" : "text-muted-foreground")}>
              {graded ? formatScore(score) : "—"}
            </span>
            {maxScore != null && maxScore > 0 && (
              <span className="text-muted-foreground text-[10px]">/ {formatScore(maxScore)}</span>
            )}
          </span>
        </div>

        <div className="flex flex-col gap-0.5 sm:min-w-[8rem]">
          <Eyebrow className="text-[9px]">{t("org.session.practiceScores.row.submitted")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 text-xs">
            {submittedStr ? (
              <>
                <ClockIcon className="size-3" />
                {submittedStr}
              </>
            ) : (
              <span className="text-muted-foreground italic">
                {t("org.session.practiceScores.row.notSubmitted")}
              </span>
            )}
          </span>
        </div>
      </div>

      <Button size="sm" variant={graded ? "outline" : "default"} onClick={onGrade}>
        <PencilIcon data-icon="inline-start" />
        {graded ? t("org.session.practiceScores.row.regrade") : t("org.session.practiceScores.row.grade")}
      </Button>
    </li>
  )
}

function SubmissionRowSkeleton() {
  return (
    <div className="bg-card border-border flex items-center gap-4 rounded-2xl border p-5 shadow-sm dark:border-0 dark:shadow-none dark:ring-1 dark:ring-foreground/10">
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

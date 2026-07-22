import type { SessionStatus } from "@/lib/session-status"
import type { ReactNode } from "react"

import { createFileRoute, Link } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CalendarClockIcon,
  CheckCircle2Icon,
  CheckSquareIcon,
  ClipboardListIcon,
  ClockIcon,
  DumbbellIcon,
  FilmIcon,
  RadioIcon,
  ShieldCheckIcon,
  SparklesIcon,
  UserCheckIcon,
  XCircleIcon,
} from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetClassesIdSessionsSessionIdAttendance } from "@/api/attendance/attendance"
import { useGetClassesId, useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { useGetLiveRooms } from "@/api/live-sessions/live-sessions"
import { useGetOfflines } from "@/api/offlines/offlines"
import { useGetPractices } from "@/api/practices/practices"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { LiveRoomsSection } from "@/components/org/livesessions/LiveRoomsSection"
import { useAttendancePermissions } from "@/components/org/livesessions/use-attendance-permissions"
import { useLivesessionPermissions } from "@/components/org/livesessions/use-livesession-permissions"
import { OfflinesSection } from "@/components/org/offlines/OfflinesSection"
import { useOfflinePermissions } from "@/components/org/offlines/use-offline-permissions"
import { PracticesSection } from "@/components/org/practices/PracticesSection"
import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { QuizzesSection } from "@/components/org/quizzes/QuizzesSection"
import { useQuizPermissions } from "@/components/org/quizzes/use-quiz-permissions"
import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatRelativeTime, formatSessionDate, useNow, useSessionStatus } from "@/lib/session-status"
import { cn } from "@/lib/utils"

// The tab selection lives in the URL (mirrors the class detail page's ?tab=
// idiom) so a session view is shareable and survives reload. Keys are
// permission-gated and computed at render, so the param stays a loose string;
// invalid/stale keys fall back to the first available surface.
const sessionDetailSearchSchema = z.object({
  tab: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/classes/class-sessions/$classSessionId")({
  head: () => orgHead("org.session.title"),
  validateSearch: sessionDetailSearchSchema,
  component: RouteComponent,
})

// A single live-status accent threads the tab dots so the eye lands on what's
// actionable — destructive when live, primary when scheduled, muted once ended.
const STATUS_DOT: Record<SessionStatus, string> = {
  live: "bg-destructive",
  scheduled: "bg-primary",
  ended: "bg-muted-foreground/40",
}

function itemsCount(payload: unknown): number {
  const p = payload as { status?: number; data?: { data?: { total?: number; items?: unknown[] } } } | undefined
  if (!p || p.status !== 200) return 0
  return p.data?.data?.total ?? p.data?.data?.items?.length ?? 0
}

// Every surface is a single leaf now — attendance, corrections, banks, and
// practice grading all moved to dedicated pages reachable from the header and
// the cards, so the page never stacks two tab tiers or long inline panels.
type Surface = {
  key: string
  navLabel: string
  icon: ReactNode
  count: number
  loading: boolean
  // Grader tabs surface "needs review" counts instead of inventory counts —
  // amber, because the number is a to-do, not a tally.
  pending?: boolean
  content: ReactNode
}

// Trailing count for a tab: a status-keyed dot + number when the surface holds
// something, nothing at all when it's empty. A "0" with a dot is noise, not
// signal, so empty surfaces fall back to their bare label.
function CountBadge({
  count,
  loading,
  status,
  pending = false,
}: {
  count: number
  loading: boolean
  status: SessionStatus
  pending?: boolean
}) {
  if (loading) return <Skeleton className="h-3 w-4" />
  if (count <= 0) return null
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 font-mono text-xs tabular-nums",
        pending ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground"
      )}
    >
      <span className={cn("size-1.5 rounded-full", pending ? "bg-amber-500" : STATUS_DOT[status])} />
      {count}
    </span>
  )
}

function DecorativeBackground() {
  return (
    <>
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/6%,transparent_55%)]"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)] [background-size:48px_48px] opacity-[0.04]"
      />
    </>
  )
}

// Relative time ("in 2 minutes") is the only header bit that ticks. Isolating
// it here — with its own minute-granularity clock — keeps the second/minute
// churn off the whole route tree; the string never changes faster than a minute
// anyway (formatRelativeTime rounds, sub-minute collapses to a constant).
function SessionRelativeTime({ startIso, status }: { startIso: string | undefined; status: SessionStatus }) {
  const { i18n } = useTranslation()
  const now = useNow(60_000)
  if (status === "ended") return null
  const relativeStr = formatRelativeTime(startIso, now, i18n.language)
  if (!relativeStr) return null
  return (
    <>
      <span className="text-muted-foreground/40">·</span>
      <span className={cn("font-medium", status === "live" ? "text-destructive" : "text-primary")}>{relativeStr}</span>
    </>
  )
}

// Amber to-do chip in the header: "N submissions await grading". Clicking it
// lands on the tab where the work lives. Rendered only when the count is
// positive — a zero would congratulate nobody.
function PendingChip({ icon, label, onClick }: { icon: ReactNode; label: string; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-3 py-1.5 text-xs font-medium text-amber-700 ring-1 ring-amber-500/30 transition-colors hover:bg-amber-500/20 dark:text-amber-300"
    >
      {icon}
      {label}
    </button>
  )
}

const ATTENDANCE_PILL_META: Record<string, { style: string; icon: typeof CheckCircle2Icon }> = {
  present: { style: "bg-emerald-500/10 text-emerald-600 ring-emerald-500/30 dark:text-emerald-400", icon: CheckCircle2Icon },
  absent: { style: "bg-destructive/10 text-destructive ring-destructive/30", icon: XCircleIcon },
  late: { style: "bg-amber-500/10 text-amber-600 ring-amber-500/30 dark:text-amber-400", icon: ClockIcon },
  excused: { style: "bg-primary/10 text-primary ring-primary/30", icon: ShieldCheckIcon },
}

// The student's one attendance fact, surfaced where they land instead of
// buried behind a tab: a status pill in the session header.
function MyAttendancePill({ status, isAuto }: { status: string; isAuto: boolean }) {
  const { t } = useTranslation()
  const meta = ATTENDANCE_PILL_META[status] ?? ATTENDANCE_PILL_META.absent
  const StatusIcon = meta.icon
  return (
    <span className="inline-flex items-center gap-2">
      <Eyebrow className="text-[10px]">{t("org.session.header.myAttendance")}</Eyebrow>
      <span
        className={cn(
          "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[10px] tracking-[0.2em] uppercase ring-1",
          meta.style
        )}
      >
        <StatusIcon className="size-3" />
        {t(`common.statuses.attendance.${status}`)}
      </span>
      <span className="text-muted-foreground font-mono text-[10px] tracking-[0.2em] uppercase">
        {t(isAuto ? "org.session.attendance.auto" : "org.session.attendance.manual")}
      </span>
    </span>
  )
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { user: accessUser } = useAccess()
  const { canView: canViewQuizzes, canEdit: canGradeQuizzes } = useQuizPermissions()
  const { canView: canViewLive, canJoin: canJoinLive } = useLivesessionPermissions()
  const { canView: canViewPractices, canGrade: canGradePractices } = usePracticePermissions()
  const { canView: canViewOfflines } = useOfflinePermissions()
  const {
    canView: canViewAttendance,
    canCreate: canCreateAttendance,
    canEdit: canEditAttendance,
  } = useAttendancePermissions()
  const canViewLiveAny = canViewLive || canJoinLive
  const canMarkAttendance = canCreateAttendance || canEditAttendance

  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  const {
    data: sessionData,
    isPending: sessionPending,
    isError: sessionError,
  } = useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id
  // Boundary-scheduled — re-renders only when the session actually flips
  // scheduled→live→ended, not on a wall-clock tick.
  const status = useSessionStatus(session?.start_time)

  const { data: classData } = useGetClassesId(classId ?? "", { query: { enabled: !!classId } })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  useBreadcrumb([
    { label: t("org.nav.classes"), to: "/org/classes" },
    {
      label: cls?.name ?? null,
      to: "/org/classes/$classId",
      params: { classId: classId ?? "" },
      loading: !cls,
    },
    { label: session?.name ?? null, loading: !session },
  ])

  const enabled = !!session
  const liveQ = useGetLiveRooms({ class_session_id: classSessionId }, { query: { enabled: enabled && canViewLiveAny } })
  const quizQ = useGetQuizzes({ class_session_id: classSessionId }, { query: { enabled: enabled && canViewQuizzes } })
  const practiceQ = useGetPractices(
    { class_session_id: classSessionId },
    { query: { enabled: enabled && canViewPractices } }
  )
  const offlineQ = useGetOfflines(
    { class_session_id: classSessionId },
    { query: { enabled: enabled && canViewOfflines } }
  )
  // Markers manage attendance on its own page now; this fetch only feeds the
  // student's header pill.
  const showMyAttendance = canViewAttendance && !canMarkAttendance && !!classId
  const attendanceQ = useGetClassesIdSessionsSessionIdAttendance(
    classId ?? "",
    classSessionId,
    undefined,
    { query: { enabled: enabled && showMyAttendance } }
  )

  if (!allowed) return null

  if (sessionPending) {
    return (
      <div className="flex flex-col gap-6 py-6">
        <Skeleton className="h-28 w-full rounded-2xl" />
        <Skeleton className="h-10 w-full max-w-md rounded-lg" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <Skeleton className="h-40 w-full rounded-2xl" />
          <Skeleton className="h-40 w-full rounded-2xl" />
          <Skeleton className="h-40 w-full rounded-2xl" />
        </div>
      </div>
    )
  }

  if (sessionError || !session) {
    return (
      <div className="flex flex-col items-start gap-4 py-16">
        <Eyebrow>{t("org.session.notFound.eyebrow")}</Eyebrow>
        <h1 className="text-3xl font-semibold tracking-tight">{t("org.session.notFound.title")}</h1>
        <p className="text-muted-foreground max-w-md text-base leading-relaxed">
          {t("org.session.notFound.description")}
        </p>
        <Button variant="outline" render={<Link to="/org/classes" />}>
          <ArrowLeftIcon className="size-4" />
          {t("org.session.notFound.backToClasses")}
        </Button>
      </div>
    )
  }

  const startStr = formatSessionDate(session.start_time, i18n.language, "long")

  const quizzes = (quizQ.data?.status === 200 && quizQ.data.data.data?.items) || []
  const practices = (practiceQ.data?.status === 200 && practiceQ.data.data.data?.items) || []

  // Grader-only to-do totals. The quiz field only arrives for callers who can
  // manage the quiz; practice stats only arrive for graders — so a plain sum
  // is already permission-scoped.
  const pendingCorrections = quizzes.reduce((sum, q) => sum + (q.pending_submissions_count ?? 0), 0)
  const pendingPracticeGrades = practices.reduce((sum, p) => {
    if (!p.can_grade || !p.stats) return sum
    return sum + Math.max((p.stats.submitted_count ?? 0) - (p.stats.graded_count ?? 0), 0)
  }, 0)

  const myAttendance = showMyAttendance
    ? ((attendanceQ.data?.status === 200 && attendanceQ.data.data.data?.items) || []).find(
        (a) => !a.user || a.user.id === accessUser.id
      )
    : undefined

  // Build each surface once; the tabs and their content both derive from it.
  const surfaces: Surface[] = []

  if (canViewLiveAny) {
    surfaces.push({
      key: "live",
      navLabel: t("org.session.nav.live"),
      icon: <RadioIcon className="size-4" />,
      count: itemsCount(liveQ.data),
      loading: liveQ.isPending,
      content: <LiveRoomsSection classSessionId={classSessionId} />,
    })
  }

  if (canViewQuizzes && classId) {
    surfaces.push({
      key: "quizzes",
      navLabel: t("org.session.nav.quizzes"),
      icon: <ClipboardListIcon className="size-4" />,
      count: canGradeQuizzes ? pendingCorrections : itemsCount(quizQ.data),
      loading: quizQ.isPending,
      pending: canGradeQuizzes,
      content: <QuizzesSection classId={classId} classSessionId={classSessionId} />,
    })
  }

  if (canViewPractices) {
    surfaces.push({
      key: "practices",
      navLabel: t("org.session.nav.practices"),
      icon: <DumbbellIcon className="size-4" />,
      count: canGradePractices ? pendingPracticeGrades : itemsCount(practiceQ.data),
      loading: practiceQ.isPending,
      pending: canGradePractices,
      content: <PracticesSection classSessionId={classSessionId} />,
    })
  }

  if (canViewOfflines) {
    surfaces.push({
      key: "offlines",
      navLabel: t("org.session.nav.recordings"),
      icon: <FilmIcon className="size-4" />,
      count: itemsCount(offlineQ.data),
      loading: offlineQ.isPending,
      content: <OfflinesSection classSessionId={classSessionId} />,
    })
  }

  const activeSurface = surfaces.find((s) => s.key === search.tab) ?? surfaces[0]

  const handleSurfaceChange = (key: string) => {
    navigate({ search: { tab: key } })
  }

  const showAttendanceButton = canMarkAttendance && canViewAttendance && !!classId
  const showChips =
    (canGradeQuizzes && pendingCorrections > 0) || (canGradePractices && pendingPracticeGrades > 0)
  const showHeaderStrip = showChips || !!myAttendance

  return (
    <div className="relative isolate flex flex-col gap-6 pb-10">
      <DecorativeBackground />

      <header className="border-foreground/10 bg-card/50 relative mt-5 flex flex-col gap-4 overflow-hidden rounded-2xl border p-4 backdrop-blur-sm md:p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex min-w-0 flex-col gap-2.5">
            <div className="flex flex-wrap items-center gap-2.5">
              <Eyebrow>{t("org.session.eyebrow")}</Eyebrow>
              <SessionStatusPill status={status} size="sm" />
            </div>

            <h1 className="max-w-2xl text-2xl leading-tight font-semibold tracking-tight text-balance md:text-3xl">
              {session.name}
            </h1>

            {session.description && (
              <p className="text-muted-foreground line-clamp-2 max-w-xl text-sm leading-relaxed">
                {session.description}
              </p>
            )}

            <div className="text-muted-foreground mt-0.5 flex flex-wrap items-center gap-x-2 gap-y-1.5 text-sm">
              <span className="inline-flex items-center gap-1.5">
                <CalendarClockIcon className="size-3.5 opacity-70" />
                {startStr}
              </span>
              <SessionRelativeTime startIso={session.start_time} status={status} />
              {cls?.name && (
                <>
                  <span className="text-muted-foreground/40">·</span>
                  <Link
                    to="/org/classes/$classId"
                    params={{ classId: classId ?? "" }}
                    className="hover:text-foreground inline-flex max-w-[22ch] items-center gap-1.5 truncate transition-colors"
                  >
                    <SparklesIcon className="size-3.5 opacity-70" />
                    {cls.name}
                  </Link>
                </>
              )}
            </div>
          </div>

          {showAttendanceButton && (
            <Button
              variant="outline"
              render={
                <Link
                  to="/org/classes/class-sessions/$classSessionId/attendance"
                  params={{ classSessionId }}
                />
              }
            >
              <UserCheckIcon className="size-4" />
              {t("org.session.attendance.title")}
            </Button>
          )}
        </div>

        {/* Role-keyed summary strip: graders see amber to-do chips, students
            see their own attendance pill. One place, two readings. */}
        {showHeaderStrip && (
          <div className="border-foreground/10 flex flex-wrap items-center gap-2 border-t border-dashed pt-3.5">
            {canGradeQuizzes && pendingCorrections > 0 && (
              <PendingChip
                icon={<CheckSquareIcon className="size-3.5" />}
                label={t("org.session.header.pendingCorrections", { count: pendingCorrections })}
                onClick={() => handleSurfaceChange("quizzes")}
              />
            )}
            {canGradePractices && pendingPracticeGrades > 0 && (
              <PendingChip
                icon={<DumbbellIcon className="size-3.5" />}
                label={t("org.session.header.pendingPracticeGrades", { count: pendingPracticeGrades })}
                onClick={() => handleSurfaceChange("practices")}
              />
            )}
            {myAttendance && (
              <MyAttendancePill status={myAttendance.status ?? "absent"} isAuto={!!myAttendance.is_auto_marked} />
            )}
          </div>
        )}
      </header>

      {surfaces.length > 0 && activeSurface && (
        <div className="flex flex-col gap-6">
          <nav className="border-foreground/10 border-b border-dashed pb-4">
            <Tabs value={activeSurface.key} onValueChange={handleSurfaceChange}>
              <TabsList
                variant="default"
                className="h-auto max-w-full [scrollbar-width:none] justify-start overflow-x-auto p-1 [&::-webkit-scrollbar]:hidden"
              >
                {surfaces.map((surface) => (
                  <TabsTrigger key={surface.key} value={surface.key} className="shrink-0 gap-2 px-3 py-1.5">
                    {surface.icon}
                    <span>{surface.navLabel}</span>
                    <CountBadge
                      count={surface.count}
                      loading={surface.loading}
                      status={status}
                      pending={surface.pending}
                    />
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          </nav>

          <div className="flex flex-col gap-6">{activeSurface.content}</div>
        </div>
      )}
    </div>
  )
}

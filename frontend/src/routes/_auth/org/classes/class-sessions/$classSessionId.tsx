import type { SessionStatus } from "@/lib/session-status"
import type { ReactNode } from "react"

import { createFileRoute, Link } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CalendarClockIcon,
  CheckSquareIcon,
  ClipboardListIcon,
  DumbbellIcon,
  FilmIcon,
  LibraryIcon,
  RadioIcon,
  SparklesIcon,
  UserCheckIcon,
  VideoIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetClassesIdSessionsSessionIdAttendance } from "@/api/attendance/attendance"
import { useGetClassesId, useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { useGetLiveRooms } from "@/api/live-sessions/live-sessions"
import { useGetOfflines } from "@/api/offlines/offlines"
import { useGetPractices } from "@/api/practices/practices"
import { useGetQuestionBanks } from "@/api/question-banks/question-banks"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { AttendanceSection } from "@/components/org/livesessions/AttendanceSection"
import { LiveRoomsSection } from "@/components/org/livesessions/LiveRoomsSection"
import { useAttendancePermissions } from "@/components/org/livesessions/use-attendance-permissions"
import { useLivesessionPermissions } from "@/components/org/livesessions/use-livesession-permissions"
import { OfflinesSection } from "@/components/org/offlines/OfflinesSection"
import { useOfflinePermissions } from "@/components/org/offlines/use-offline-permissions"
import { PracticeScoresSection } from "@/components/org/practices/PracticeScoresSection"
import { PracticesSection } from "@/components/org/practices/PracticesSection"
import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { QuestionBanksSection } from "@/components/org/question-banks/QuestionBanksSection"
import { useBankPermissions } from "@/components/org/question-banks/use-bank-permissions"
import { QuizCorrectionsSection } from "@/components/org/quizzes/QuizCorrectionsSection"
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

// Surface and leaf selection live in the URL (mirrors the class detail page's
// ?tab= idiom) so a session view is shareable and survives reload. Keys are
// permission-gated and computed at render, so both params stay loose strings;
// invalid/stale keys fall back to the first available surface/leaf.
const sessionDetailSearchSchema = z.object({
  tab: z.string().optional(),
  subtab: z.string().optional(),
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

// One surface drives a single top-level tab. Surfaces with more than one leaf
// section expose them as light line sub-tabs — mirrors the class detail page's
// flat tab idiom rather than the old overview-dashboard detour.
type SubTab = { key: string; label: string; count: number; loading: boolean; icon: ReactNode; content: ReactNode }

type Surface = {
  key: string
  navLabel: string
  icon: ReactNode
  count: number
  loading: boolean
  subTabs: SubTab[]
}

// Trailing count for a tab: a status-keyed dot + number when the surface holds
// something, nothing at all when it's empty. A "0" with a dot is noise, not
// signal, so empty surfaces fall back to their bare label.
function CountBadge({ count, loading, status }: { count: number; loading: boolean; status: SessionStatus }) {
  if (loading) return <Skeleton className="h-3 w-4" />
  if (count <= 0) return null
  return (
    <span className="text-muted-foreground inline-flex items-center gap-1 font-mono text-xs tabular-nums">
      <span className={cn("size-1.5 rounded-full", STATUS_DOT[status])} />
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

// Light line sub-tabs, rendered as the second tier of the unified nav group.
// Presentational only — selection state is lifted to the route so both tiers
// share one bordered surface instead of floating as two detached bars.
function SubTabsBar({
  tabs,
  value,
  onValueChange,
  status,
}: {
  tabs: SubTab[]
  value: string
  onValueChange: (key: string) => void
  status: SessionStatus
}) {
  return (
    <Tabs value={value} onValueChange={onValueChange}>
      {/* Scrolls horizontally on narrow viewports rather than wrapping into a
          jumbled stack. The -mb/pb pair reserves room for the active underline
          (bottom-[-5px]) so the overflow clip doesn't shave it off. */}
      <div className="-mb-1.5 max-w-full [scrollbar-width:none] overflow-x-auto pb-1.5 [&::-webkit-scrollbar]:hidden">
        <TabsList variant="line">
          {tabs.map((tab) => (
            <TabsTrigger key={tab.key} value={tab.key} className="shrink-0 gap-2">
              {tab.icon}
              <span>{tab.label}</span>
              <CountBadge count={tab.count} loading={tab.loading} status={status} />
            </TabsTrigger>
          ))}
        </TabsList>
      </div>
    </Tabs>
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

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView: canViewBanks } = useBankPermissions()
  const { canView: canViewQuizzes, canEdit: canGradeQuizzes } = useQuizPermissions()
  const { canView: canViewLive, canJoin: canJoinLive } = useLivesessionPermissions()
  const { canView: canViewPractices, canGrade: canGradePractices } = usePracticePermissions()
  const { canView: canViewOfflines } = useOfflinePermissions()
  const { canView: canViewAttendance } = useAttendancePermissions()
  const canViewLiveAny = canViewLive || canJoinLive

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
  const banksQ = useGetQuestionBanks(undefined, { query: { enabled: enabled && canViewBanks } })
  const attendanceQ = useGetClassesIdSessionsSessionIdAttendance(
    classId ?? "",
    classSessionId,
    { order_by: "status", order_dir: "asc" },
    { query: { enabled: enabled && canViewAttendance && !!classId } }
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

  // Build each surface once; the top tabs and their content both derive from it.
  const surfaces: Surface[] = []

  if (canViewLiveAny) {
    const subTabs: SubTab[] = [
      {
        key: "rooms",
        label: t("org.session.liveWorkspace.tabs.rooms"),
        count: itemsCount(liveQ.data),
        loading: liveQ.isPending,
        icon: <VideoIcon className="size-4" />,
        content: <LiveRoomsSection classSessionId={classSessionId} />,
      },
    ]
    if (canViewAttendance && classId) {
      subTabs.push({
        key: "presence",
        label: t("org.session.liveWorkspace.tabs.presence"),
        count: itemsCount(attendanceQ.data),
        loading: attendanceQ.isPending,
        icon: <UserCheckIcon className="size-4" />,
        content: <AttendanceSection classId={classId} classSessionId={classSessionId} />,
      })
    }
    surfaces.push({
      key: "live",
      navLabel: t("org.session.nav.live"),
      icon: <RadioIcon className="size-4" />,
      count: itemsCount(liveQ.data),
      loading: liveQ.isPending,
      subTabs,
    })
  }

  if (canViewQuizzes) {
    const subTabs: SubTab[] = []
    if (classId) {
      subTabs.push({
        key: "quizzes",
        label: t("org.session.workspace.tabs.quizzes"),
        count: itemsCount(quizQ.data),
        loading: quizQ.isPending,
        icon: <ClipboardListIcon className="size-4" />,
        content: <QuizzesSection classId={classId} classSessionId={classSessionId} />,
      })
    }
    if (canGradeQuizzes) {
      subTabs.push({
        key: "corrections",
        label: t("org.session.workspace.tabs.corrections"),
        count: itemsCount(quizQ.data),
        loading: quizQ.isPending,
        icon: <CheckSquareIcon className="size-4" />,
        content: <QuizCorrectionsSection classSessionId={classSessionId} />,
      })
    }
    if (canViewBanks) {
      subTabs.push({
        key: "banks",
        label: t("org.session.workspace.tabs.banks"),
        count: itemsCount(banksQ.data),
        loading: banksQ.isPending,
        icon: <LibraryIcon className="size-4" />,
        content: <QuestionBanksSection />,
      })
    }
    surfaces.push({
      key: "quizzes",
      navLabel: t("org.session.nav.quizzes"),
      icon: <ClipboardListIcon className="size-4" />,
      count: itemsCount(quizQ.data),
      loading: quizQ.isPending,
      subTabs,
    })
  }

  if (canViewPractices) {
    const subTabs: SubTab[] = [
      {
        key: "practices",
        label: t("org.session.practiceWorkspace.tabs.practices"),
        count: itemsCount(practiceQ.data),
        loading: practiceQ.isPending,
        icon: <DumbbellIcon className="size-4" />,
        content: <PracticesSection classSessionId={classSessionId} />,
      },
    ]
    if (canGradePractices) {
      subTabs.push({
        key: "practiceScores",
        label: t("org.session.practiceWorkspace.tabs.practiceScores"),
        count: itemsCount(practiceQ.data),
        loading: practiceQ.isPending,
        icon: <CheckSquareIcon className="size-4" />,
        content: <PracticeScoresSection classSessionId={classSessionId} />,
      })
    }
    surfaces.push({
      key: "practices",
      navLabel: t("org.session.nav.practices"),
      icon: <DumbbellIcon className="size-4" />,
      count: itemsCount(practiceQ.data),
      loading: practiceQ.isPending,
      subTabs,
    })
  }

  if (canViewOfflines) {
    surfaces.push({
      key: "offlines",
      navLabel: t("org.session.nav.recordings"),
      icon: <FilmIcon className="size-4" />,
      count: itemsCount(offlineQ.data),
      loading: offlineQ.isPending,
      subTabs: [
        {
          key: "offlines",
          label: t("org.session.offlineWorkspace.tabs.offlines"),
          count: itemsCount(offlineQ.data),
          loading: offlineQ.isPending,
          icon: <FilmIcon className="size-4" />,
          content: <OfflinesSection classSessionId={classSessionId} />,
        },
      ],
    })
  }

  const activeSurface = surfaces.find((s) => s.key === search.tab) ?? surfaces[0]
  // A stored sub-tab only applies while its parent surface is active; switching
  // surfaces falls back to the new surface's first leaf so a stale key never
  // renders a blank panel.
  const activeSub = activeSurface?.subTabs.find((s) => s.key === search.subtab) ?? activeSurface?.subTabs[0]

  const handleSurfaceChange = (key: string) => {
    navigate({ search: { tab: key } })
  }

  const handleSubChange = (key: string) => {
    navigate({ search: { ...search, subtab: key } })
  }

  return (
    <div className="relative isolate flex flex-col gap-6 pb-10">
      <DecorativeBackground />

      <header className="border-foreground/10 bg-card/50 relative mt-5 flex flex-col gap-5 overflow-hidden rounded-2xl border p-4 backdrop-blur-sm md:p-5">
        <div className="flex min-w-0 flex-col gap-2.5">
          <div className="flex flex-wrap items-center gap-2.5">
            <Eyebrow>{t("org.session.eyebrow")}</Eyebrow>
            <SessionStatusPill status={status} size="sm" />
          </div>

          <h1 className="max-w-2xl text-2xl leading-tight font-semibold tracking-tight text-balance md:text-3xl">
            {session.name}
          </h1>

          {session.description && (
            <p className="text-muted-foreground line-clamp-2 max-w-xl text-sm leading-relaxed">{session.description}</p>
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
      </header>

      {surfaces.length > 0 && activeSurface && activeSub && (
        <div className="flex flex-col gap-6">
          {/* Both tab tiers share one bordered surface so they read as a single
              nav group rather than two detached, right-ragged bars. The dashed
              underline echoes the section dividers used across the app. */}
          <nav className="border-foreground/10 flex flex-col gap-3 border-b border-dashed pb-4">
            {/* Tier 1 — primary surfaces as a filled segmented control. The
                filled treatment sets it apart from the line sub-tabs below. */}
            <Tabs value={activeSurface.key} onValueChange={handleSurfaceChange}>
              <TabsList
                variant="default"
                className="h-auto max-w-full [scrollbar-width:none] justify-start overflow-x-auto p-1 [&::-webkit-scrollbar]:hidden"
              >
                {surfaces.map((surface) => (
                  <TabsTrigger key={surface.key} value={surface.key} className="shrink-0 gap-2 px-3 py-1.5">
                    {surface.icon}
                    <span>{surface.navLabel}</span>
                    <CountBadge count={surface.count} loading={surface.loading} status={status} />
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>

            {/* Tier 2 — leaf sections of the active surface. Hidden when the
                surface has a single leaf, so a lone underlined tab never sits
                orphaned beneath the primary control. */}
            {activeSurface.subTabs.length > 1 && (
              <SubTabsBar
                tabs={activeSurface.subTabs}
                value={activeSub.key}
                onValueChange={handleSubChange}
                status={status}
              />
            )}
          </nav>

          <div className="flex flex-col gap-6">{activeSub.content}</div>
        </div>
      )}
    </div>
  )
}

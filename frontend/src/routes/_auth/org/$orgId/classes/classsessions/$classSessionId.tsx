import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"
import type { SessionStatus } from "@/lib/session-status"
import type { ReactNode } from "react"

import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CalendarClockIcon,
  CheckSquareIcon,
  ClipboardListIcon,
  DumbbellIcon,
  FilmIcon,
  LayoutDashboardIcon,
  LibraryIcon,
  PlusIcon,
  RadioIcon,
  SparklesIcon,
  UserCheckIcon,
  VideoIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetClassesIdSessionsSessionIdAttendance } from "@/api/attendance/attendance"
import { useGetClassesId, useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { getLiveRooms, useGetLiveRooms, usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { useGetOfflines } from "@/api/offlines/offlines"
import { useGetPractices } from "@/api/practices/practices"
import { useGetQuestionBanks } from "@/api/question-banks/question-banks"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { getSessionStatus, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/classes/classsessions/$classSessionId")({
  head: () => orgHead("org.session.title"),
  component: RouteComponent,
})

type Accent = { text: string; bg: string; border: string; glow: string; dot: string }

const ACCENTS: Record<SessionStatus, Accent> = {
  live: {
    text: "text-destructive",
    bg: "bg-destructive/12 dark:bg-destructive/10",
    border: "border-destructive/45 dark:border-destructive/40",
    glow: "from-destructive/30 via-destructive/8 dark:from-destructive/25 dark:via-destructive/5",
    dot: "bg-destructive",
  },
  scheduled: {
    text: "text-primary",
    bg: "bg-primary/12 dark:bg-primary/10",
    border: "border-primary/40 dark:border-primary/30",
    glow: "from-primary/25 via-primary/8 dark:from-primary/20 dark:via-primary/5",
    dot: "bg-primary",
  },
  ended: {
    text: "text-muted-foreground",
    bg: "bg-muted",
    border: "border-border",
    glow: "from-foreground/12 via-foreground/0 dark:from-foreground/8",
    dot: "bg-muted-foreground",
  },
}

// Static class lookups — Tailwind needs literal class names, so dynamic counts map here.
const STAT_COLS = ["", "md:grid-cols-1", "md:grid-cols-2", "md:grid-cols-3", "md:grid-cols-4", "md:grid-cols-5"]

const pad = (n: number) => String(n).padStart(2, "0")

function countdownLabel(iso: string | undefined, now: number): string {
  if (!iso) return "—"
  const target = new Date(iso).getTime()
  if (Number.isNaN(target)) return "—"
  const abs = Math.abs(target - now)
  const days = Math.floor(abs / 86_400_000)
  const hours = Math.floor((abs % 86_400_000) / 3_600_000)
  const minutes = Math.floor((abs % 3_600_000) / 60_000)
  const seconds = Math.floor((abs % 60_000) / 1000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${pad(minutes)}m`
  return `${pad(minutes)}m ${pad(seconds)}s`
}

function itemsCount(payload: unknown): number {
  const p = payload as { status?: number; data?: { data?: { total?: number; items?: unknown[] } } } | undefined
  if (!p || p.status !== 200) return 0
  return p.data?.data?.total ?? p.data?.data?.items?.length ?? 0
}

// ── Descriptors ──────────────────────────────────────────────────────────────
// One surface drives its overview stat, overview card, top tab, and sub-tabs.

type SubTab = { key: string; label: string; count: number; loading: boolean; icon: ReactNode; content: ReactNode }

type Surface = {
  key: string
  eyebrow: string
  title: string
  navLabel: string
  statLabel: string
  icon: ReactNode
  count: number
  loading: boolean
  canCreate: boolean
  newLabel: string
  summaryLabel: string
  emptyHint: string
  subTabs: SubTab[]
}

// ── Presentational pieces ─────────────────────────────────────────────────────

function DecorativeBackground({ accent }: { accent: Accent }) {
  return (
    <>
      <div
        aria-hidden
        className={cn(
          "pointer-events-none absolute inset-x-0 -top-32 -z-10 h-[420px] bg-gradient-to-b to-transparent opacity-70 blur-3xl dark:opacity-90",
          accent.glow
        )}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_60%_40%_at_50%_0%,black,transparent_75%)] [background-size:56px_56px] opacity-[0.06] dark:opacity-[0.035]"
      />
    </>
  )
}

function Breadcrumb({
  orgId,
  classId,
  classLabel,
}: {
  orgId: string
  classId: string
  classLabel: string | undefined
}) {
  const { t } = useTranslation()
  return (
    <div className="animate-in fade-in-0 slide-in-from-top-2 fill-mode-both flex items-center gap-2 pt-6 font-mono text-xs tracking-[0.25em] uppercase duration-500">
      <Link
        to="/org/$orgId/classes"
        params={{ orgId }}
        className="text-muted-foreground hover:text-foreground transition-colors"
      >
        {t("org.nav.classes")}
      </Link>
      <span className="text-muted-foreground/40" aria-hidden>
        /
      </span>
      <Link
        to="/org/$orgId/classes/$classId"
        params={{ orgId, classId }}
        className="text-foreground max-w-[22ch] truncate"
      >
        {classLabel ?? t("org.nav.classes")}
      </Link>
    </div>
  )
}

function JoinAction({ session, accent }: { session: Session; accent: Accent }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { canCreate: canStart, canJoin } = useLivesessionPermissions()

  const join = usePostLiveRooms({
    mutation: {
      onSuccess: (result) => {
        const room = (result.status === 201 && result.data.data) || undefined
        if (room?.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
      },
      onError: async (err, variables) => {
        if ((err as ErrorType<unknown>).response?.status !== 409) return
        try {
          const rooms = await getLiveRooms()
          const roomsData = (rooms.status === 200 && rooms.data.data) || undefined
          const room = (roomsData?.items ?? []).find((r) => r.class_session_id === variables.data.class_session_id)
          if (room?.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
        } catch {
          // ignore
        }
      },
    },
  })

  if (!canStart && !canJoin) {
    return (
      <Button variant="outline" disabled>
        {t("org.session.actions.notPermitted")}
      </Button>
    )
  }

  return (
    <Button
      className="relative overflow-hidden shadow-[0_8px_24px_-12px_var(--color-primary)]"
      disabled={join.isPending || !session.id}
      onClick={() => session.id && join.mutate({ data: { class_session_id: session.id } })}
    >
      <span className={cn("absolute inset-0 -z-10 opacity-30 blur-xl", accent.bg)} aria-hidden />
      <RadioIcon className="size-4" />
      {canStart ? t("org.session.actions.start") : t("org.session.actions.join")}
    </Button>
  )
}

function SessionHeader({
  session,
  status,
  accent,
  classId,
  orgId,
  shortId,
}: {
  session: Session
  status: SessionStatus
  accent: Accent
  classId: string
  orgId: string
  shortId: string
}) {
  const { t } = useTranslation()
  const statusLabel = t(`status.${status === "live" ? "liveNow" : status}`)
  return (
    <header
      className={cn(
        "animate-in fade-in-0 slide-in-from-bottom-3 fill-mode-both bg-card relative isolate overflow-hidden rounded-3xl border p-7 shadow-sm duration-500 md:p-9",
        "dark:bg-card/50 dark:ring-foreground/8 dark:border-0 dark:shadow-none dark:ring-1 dark:backdrop-blur-sm",
        accent.border
      )}
    >
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_70%_90%_at_85%_-10%,black,transparent_70%)] [background-size:40px_40px] opacity-[0.05]"
      />
      <div
        aria-hidden
        className={cn("pointer-events-none absolute -end-24 -top-28 -z-10 h-72 w-72 rounded-full blur-3xl", accent.bg)}
      />

      <div className="flex items-center justify-between gap-3">
        <Link
          to="/org/$orgId/classes/$classId"
          params={{ orgId, classId }}
          className="text-muted-foreground hover:text-foreground group inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5 transition-transform group-hover:-translate-x-0.5 rtl:rotate-180 rtl:group-hover:translate-x-0.5" />
          {t("org.session.backToClass")}
        </Link>
        <span className="text-muted-foreground/70 inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em]">
          <span className="bg-muted-foreground/30 h-px w-8" />№ {shortId || "—"}
        </span>
      </div>

      <div className="mt-8 flex flex-col gap-4">
        <div className="flex flex-wrap items-center gap-2.5 font-mono text-xs tracking-[0.25em] uppercase">
          {status === "live" ? (
            <SessionStatusPill status={status} size="sm" />
          ) : (
            <span className={cn("font-semibold", accent.text)}>{statusLabel}</span>
          )}
          <span className="text-muted-foreground/40" aria-hidden>
            —
          </span>
          <span className="text-muted-foreground">{t("org.session.eyebrow")}</span>
        </div>

        <h1 className="max-w-3xl text-4xl leading-[1.05] font-semibold tracking-tight text-balance md:text-5xl">
          {session.name}
        </h1>

        {session.description ? (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed">{session.description}</p>
        ) : null}
      </div>

      <div className="mt-7 flex flex-wrap items-center gap-3">
        <JoinAction session={session} accent={accent} />
        <Button variant="outline" render={<Link to="/org/$orgId/classes/$classId" params={{ orgId, classId }} />}>
          <SparklesIcon className="size-4" />
          {t("org.session.actions.viewClass")}
        </Button>
      </div>
    </header>
  )
}

function StatCell({
  label,
  value,
  loading,
  index,
  className,
}: {
  label: string
  value: ReactNode
  loading?: boolean
  index: number
  className?: string
}) {
  return (
    <div
      className="border-border animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both flex flex-col gap-3 border-b border-dashed px-5 py-5 md:border-s md:border-b-0 md:first:border-s-0"
      style={{ animationDelay: `${index * 60}ms`, animationDuration: "450ms" }}
    >
      <Eyebrow className="text-[10px]">{label}</Eyebrow>
      {loading ? (
        <Skeleton className="h-8 w-14" />
      ) : (
        <span
          className={cn(
            "text-foreground font-mono text-3xl leading-none font-semibold tracking-tight tabular-nums md:text-4xl",
            className
          )}
        >
          {value}
        </span>
      )}
    </div>
  )
}

function SummaryCard({
  surface,
  index,
  accent,
  onOpen,
}: {
  surface: Surface
  index: number
  accent: Accent
  onOpen: () => void
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onOpen}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault()
          onOpen()
        }
      }}
      className={cn(
        "group/card bg-card text-card-foreground border-border animate-in fade-in-0 slide-in-from-bottom-3 fill-mode-both relative isolate flex cursor-pointer flex-col gap-4 overflow-hidden rounded-2xl border p-5 shadow-sm transition-all duration-300",
        "hover:border-foreground/25 focus-visible:ring-primary/40 hover:-translate-y-0.5 hover:shadow-lg focus-visible:ring-2 focus-visible:outline-none",
        "dark:ring-foreground/8 dark:hover:ring-foreground/30 dark:border-0 dark:shadow-none dark:ring-1"
      )}
      style={{ animationDelay: `${index * 70}ms`, animationDuration: "450ms" }}
    >
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/card:opacity-100"
      />

      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <div className="bg-muted text-foreground/80 group-hover/card:bg-primary/10 group-hover/card:text-foreground flex size-10 shrink-0 items-center justify-center rounded-xl transition-colors">
            {surface.icon}
          </div>
          <div className="flex flex-col gap-1">
            <Eyebrow className="text-[10px]">{surface.eyebrow}</Eyebrow>
            <h3 className="text-lg font-semibold tracking-tight">{surface.title}</h3>
          </div>
        </div>
        {surface.canCreate ? (
          <Button
            size="sm"
            onClick={(e) => {
              e.stopPropagation()
              onOpen()
            }}
          >
            <PlusIcon className="size-4" />
            {surface.newLabel}
          </Button>
        ) : null}
      </div>

      <div className="border-border mt-auto flex items-center justify-between gap-3 border-t border-dashed pt-4">
        {surface.loading ? (
          <Skeleton className="h-4 w-40" />
        ) : surface.count > 0 ? (
          <span className="text-foreground inline-flex items-center gap-2 text-sm font-medium">
            <span className={cn("font-mono text-lg font-semibold tabular-nums", accent.text)}>{surface.count}</span>
            {surface.summaryLabel}
          </span>
        ) : (
          <span className="text-muted-foreground text-sm leading-relaxed">{surface.emptyHint}</span>
        )}
        <span
          className={cn(
            "text-muted-foreground group-hover/card:bg-foreground group-hover/card:text-background inline-flex size-7 shrink-0 items-center justify-center rounded-full transition-all group-hover/card:translate-x-0.5 rtl:group-hover/card:-translate-x-0.5",
            accent.text
          )}
        >
          <span className="rtl:rotate-180" aria-hidden>
            →
          </span>
        </span>
      </div>
    </div>
  )
}

function SubTabs({ tabs, accent }: { tabs: SubTab[]; accent: Accent }) {
  if (tabs.length === 0) return null
  if (tabs.length === 1) return <div className="flex flex-col gap-6">{tabs[0]!.content}</div>

  return (
    <Tabs defaultValue={tabs[0]?.key} className="gap-6">
      <div className="bg-card border-border dark:bg-card/40 dark:ring-foreground/10 rounded-2xl border p-1.5 shadow-sm dark:border-0 dark:shadow-none dark:ring-1">
        <TabsList variant="line" className="h-auto w-full flex-wrap gap-1 bg-transparent p-0">
          {tabs.map((tab) => (
            <TabsTrigger
              key={tab.key}
              value={tab.key}
              className={cn(
                "group/wstab flex-1 justify-start gap-2.5 rounded-xl px-4 py-3",
                "data-active:bg-muted data-active:text-foreground dark:data-active:bg-foreground/5"
              )}
            >
              <span className="bg-muted text-muted-foreground group-data-[active]/wstab:bg-foreground group-data-[active]/wstab:text-background flex size-8 items-center justify-center rounded-lg transition-colors">
                {tab.icon}
              </span>
              <span className="flex flex-col items-start gap-0.5">
                <span className="text-sm leading-none font-semibold tracking-tight">{tab.label}</span>
                <span className="text-muted-foreground inline-flex items-center gap-1 font-mono text-[10px] tabular-nums">
                  {tab.loading ? (
                    <Skeleton className="h-3 w-6" />
                  ) : (
                    <>
                      <span
                        className={cn("size-1 rounded-full", tab.count > 0 ? accent.dot : "bg-muted-foreground/40")}
                        aria-hidden
                      />
                      {tab.count}
                    </>
                  )}
                </span>
              </span>
            </TabsTrigger>
          ))}
        </TabsList>
      </div>
      {tabs.map((tab) => (
        <TabsContent key={tab.key} value={tab.key} className="mt-0 flex flex-col gap-6">
          {tab.content}
        </TabsContent>
      ))}
    </Tabs>
  )
}

// ── Route ─────────────────────────────────────────────────────────────────────

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId, classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView: canViewBanks } = useBankPermissions()
  const { canView: canViewQuizzes, canEdit: canGradeQuizzes, canCreate: canCreateQuizzes } = useQuizPermissions()
  const { canView: canViewLive, canJoin: canJoinLive, canCreate: canCreateLive } = useLivesessionPermissions()
  const {
    canView: canViewPractices,
    canGrade: canGradePractices,
    canCreate: canCreatePractices,
  } = usePracticePermissions()
  const { canView: canViewOfflines, canCreate: canCreateOfflines } = useOfflinePermissions()
  const { canView: canViewAttendance } = useAttendancePermissions()
  const canViewLiveAny = canViewLive || canJoinLive
  const now = useNow(1000)

  const [tab, setTab] = useState("overview")

  const {
    data: sessionData,
    isPending: sessionPending,
    isError: sessionError,
  } = useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id

  const { data: classData } = useGetClassesId(classId ?? "", { query: { enabled: !!classId } })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

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
      <div className="flex flex-col gap-8 py-10">
        <Skeleton className="h-5 w-48" />
        <Skeleton className="h-60 w-full rounded-3xl" />
        <Skeleton className="h-12 w-full rounded-2xl" />
        <Skeleton className="h-28 w-full rounded-2xl" />
        <div className="grid gap-4 md:grid-cols-2">
          <Skeleton className="h-40 w-full rounded-2xl" />
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
        <Button variant="outline" render={<Link to="/org/$orgId/classes" params={{ orgId }} />}>
          <ArrowLeftIcon className="size-4" />
          {t("org.session.notFound.backToClasses")}
        </Button>
      </div>
    )
  }

  const status = getSessionStatus(session.start_time, now)
  const accent = ACCENTS[status]
  const shortId = (session.id ?? "").slice(0, 8).toUpperCase()

  // Build each surface once; stats, overview cards, and top tabs all derive from it.
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
    const count = itemsCount(liveQ.data)
    surfaces.push({
      key: "live",
      eyebrow: t("org.session.liveRooms.eyebrow"),
      title: t("org.session.liveRooms.title"),
      navLabel: t("org.session.nav.live"),
      statLabel: t("org.session.overview.stats.live"),
      icon: <VideoIcon className="size-5" />,
      count,
      loading: liveQ.isPending,
      canCreate: canCreateLive,
      newLabel: t("org.session.liveRooms.newRoom"),
      summaryLabel: t("org.session.overview.units.rooms", { count }),
      emptyHint: canCreateLive ? t("org.session.liveRooms.emptyHint") : t("org.session.liveRooms.emptyHintMember"),
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
    const count = itemsCount(quizQ.data)
    surfaces.push({
      key: "quizzes",
      eyebrow: t("org.session.quizzes.eyebrow"),
      title: t("org.session.quizzes.title"),
      navLabel: t("org.session.nav.quizzes"),
      statLabel: t("org.session.overview.stats.quizzes"),
      icon: <ClipboardListIcon className="size-5" />,
      count,
      loading: quizQ.isPending,
      canCreate: canCreateQuizzes,
      newLabel: t("org.session.quizzes.newQuiz"),
      summaryLabel: t("org.session.overview.units.quizzes", { count }),
      emptyHint: t("org.session.quizzes.emptyHint"),
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
    const count = itemsCount(practiceQ.data)
    surfaces.push({
      key: "practices",
      eyebrow: t("org.session.practices.eyebrow"),
      title: t("org.session.practices.title"),
      navLabel: t("org.session.nav.practices"),
      statLabel: t("org.session.overview.stats.practices"),
      icon: <DumbbellIcon className="size-5" />,
      count,
      loading: practiceQ.isPending,
      canCreate: canCreatePractices,
      newLabel: t("org.session.practices.newPractice"),
      summaryLabel: t("org.session.overview.units.practices", { count }),
      emptyHint: canCreatePractices ? t("org.session.practices.emptyHint") : t("org.session.practices.emptyHintMember"),
      subTabs,
    })
  }

  if (canViewOfflines) {
    const count = itemsCount(offlineQ.data)
    surfaces.push({
      key: "recordings",
      eyebrow: t("org.session.offlines.eyebrow"),
      title: t("org.session.offlines.title"),
      navLabel: t("org.session.nav.recordings"),
      statLabel: t("org.session.overview.stats.offline"),
      icon: <FilmIcon className="size-5" />,
      count,
      loading: offlineQ.isPending,
      canCreate: canCreateOfflines,
      newLabel: t("org.session.offlines.newOffline"),
      summaryLabel: t("org.session.overview.units.recordings", { count }),
      emptyHint: canCreateOfflines ? t("org.session.offlines.emptyHint") : t("org.session.offlines.emptyHintMember"),
      subTabs: [
        {
          key: "offlines",
          label: t("org.session.offlineWorkspace.tabs.offlines"),
          count,
          loading: offlineQ.isPending,
          icon: <FilmIcon className="size-4" />,
          content: <OfflinesSection classSessionId={classSessionId} orgId={orgId} />,
        },
      ],
    })
  }

  const getCountdownStatLabel = () => {
    if (status === "live") return t("status.liveNow")
    if (status === "ended") return t("status.ended")
    return t("org.session.meta.countdown")
  }
  const countdownStatLabel = getCountdownStatLabel()
  const statCount = surfaces.length + 1

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <DecorativeBackground accent={accent} />

      <Breadcrumb orgId={orgId} classId={classId ?? ""} classLabel={cls?.name} />

      <SessionHeader
        session={session}
        status={status}
        accent={accent}
        classId={classId ?? ""}
        orgId={orgId}
        shortId={shortId}
      />

      <Tabs value={tab} onValueChange={setTab} className="gap-8">
        <div className="bg-card border-border dark:bg-card/40 dark:ring-foreground/10 sticky top-2 z-10 rounded-2xl border p-1.5 shadow-sm dark:border-0 dark:shadow-none dark:ring-1 dark:backdrop-blur-md">
          <TabsList variant="line" className="h-auto w-full flex-wrap gap-1 bg-transparent p-0">
            <TabsTrigger
              value="overview"
              className="data-active:bg-foreground data-active:text-background gap-2 rounded-xl px-4 py-2.5 data-active:shadow-sm"
            >
              <LayoutDashboardIcon className="size-4" />
              <span className="text-sm font-semibold tracking-tight">{t("org.session.nav.overview")}</span>
            </TabsTrigger>
            {surfaces.map((surface) => (
              <TabsTrigger
                key={surface.key}
                value={surface.key}
                className="group/toptab data-active:bg-foreground data-active:text-background gap-2 rounded-xl px-4 py-2.5 data-active:shadow-sm"
              >
                <span className="text-sm font-semibold tracking-tight">{surface.navLabel}</span>
                {surface.loading ? (
                  <Skeleton className="h-3 w-4" />
                ) : (
                  <span className="inline-flex items-center gap-1 font-mono text-xs tabular-nums">
                    <span
                      className={cn(
                        "group-data-[active]/toptab:bg-background size-1.5 rounded-full",
                        surface.count > 0 ? accent.dot : "bg-muted-foreground/40"
                      )}
                      aria-hidden
                    />
                    {surface.count}
                  </span>
                )}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>

        <TabsContent value="overview" className="mt-0 flex flex-col gap-8">
          <section
            className={cn(
              "bg-card border-border grid grid-cols-2 overflow-hidden rounded-2xl border shadow-sm",
              "dark:bg-card/40 dark:ring-foreground/8 dark:border-0 dark:shadow-none dark:ring-1 dark:backdrop-blur-sm",
              STAT_COLS[statCount] ?? "md:grid-cols-5"
            )}
          >
            <StatCell
              index={0}
              label={countdownStatLabel}
              value={status === "ended" ? "—" : countdownLabel(session.start_time, now)}
              className={accent.text}
            />
            {surfaces.map((surface, i) => (
              <StatCell
                key={surface.key}
                index={i + 1}
                label={surface.statLabel}
                value={surface.count}
                loading={surface.loading}
              />
            ))}
          </section>

          {surfaces.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2">
              {surfaces.map((surface, i) => (
                <SummaryCard
                  key={surface.key}
                  surface={surface}
                  index={i}
                  accent={accent}
                  onOpen={() => setTab(surface.key)}
                />
              ))}
            </div>
          ) : null}
        </TabsContent>

        {surfaces.map((surface) => (
          <TabsContent key={surface.key} value={surface.key} className="mt-0">
            <SubTabs tabs={surface.subTabs} accent={accent} />
          </TabsContent>
        ))}
      </Tabs>

      <footer className="border-border border-t border-dashed pt-6">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <Eyebrow className="inline-flex items-center gap-2">
            <CalendarClockIcon className="size-3.5" />
            {t("org.session.footnote")}
          </Eyebrow>
          <Eyebrow>
            {t("common.brandName")} · {cls?.name ?? "—"}
          </Eyebrow>
        </div>
      </footer>
    </div>
  )
}

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
  LibraryIcon,
  RadioIcon,
  SparklesIcon,
  UserCheckIcon,
  VideoIcon,
} from "lucide-react"
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
import { formatSessionDate, getSessionStatus, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/classes/classsessions/$classSessionId")({
  head: () => orgHead("org.session.title"),
  component: RouteComponent,
})

type Accent = {
  text: string
  ring: string
  bg: string
  border: string
  glow: string
  dot: string
}

const ACCENTS: Record<SessionStatus, Accent> = {
  live: {
    text: "text-destructive",
    ring: "ring-destructive/30",
    bg: "bg-destructive/12 dark:bg-destructive/10",
    border: "border-destructive/45 dark:border-destructive/40",
    glow: "from-destructive/30 via-destructive/8 dark:from-destructive/25 dark:via-destructive/5",
    dot: "bg-destructive",
  },
  scheduled: {
    text: "text-primary",
    ring: "ring-primary/30",
    bg: "bg-primary/12 dark:bg-primary/10",
    border: "border-primary/40 dark:border-primary/30",
    glow: "from-primary/25 via-primary/8 dark:from-primary/20 dark:via-primary/5",
    dot: "bg-primary",
  },
  ended: {
    text: "text-muted-foreground",
    ring: "ring-border",
    bg: "bg-muted",
    border: "border-border",
    glow: "from-foreground/12 via-foreground/0 dark:from-foreground/8",
    dot: "bg-muted-foreground",
  },
}

function countdownParts(iso: string | undefined, now: number) {
  if (!iso) return null
  const target = new Date(iso).getTime()
  if (Number.isNaN(target)) return null
  const diff = target - now
  const abs = Math.abs(diff)
  return {
    isPast: diff <= 0,
    days: Math.floor(abs / 86_400_000),
    hours: Math.floor((abs % 86_400_000) / 3_600_000),
    minutes: Math.floor((abs % 3_600_000) / 60_000),
    seconds: Math.floor((abs % 60_000) / 1000),
  }
}

const pad = (n: number) => String(n).padStart(2, "0")

function itemsCount(payload: unknown): number {
  const p = payload as { status?: number; data?: { data?: { items?: unknown[] } } } | undefined
  if (!p || p.status !== 200) return 0
  return p.data?.data?.items?.length ?? 0
}

function DecorativeBackground({ accent }: { accent: Accent }) {
  return (
    <>
      <div
        aria-hidden
        className={cn(
          "pointer-events-none absolute inset-x-0 -top-32 -z-10 h-[480px] bg-gradient-to-b to-transparent opacity-80 blur-3xl dark:opacity-100",
          accent.glow
        )}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_60%_45%_at_50%_0%,black,transparent_75%)] [background-size:56px_56px] opacity-[0.07] dark:opacity-[0.04]"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 opacity-[0.025] mix-blend-overlay"
        style={{
          backgroundImage:
            "url(\"data:image/svg+xml;utf8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='200' height='200'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")",
        }}
      />
    </>
  )
}

function Breadcrumb({
  orgId,
  classId,
  className: classLabel,
  shortId,
  fallback,
}: {
  orgId: string
  classId: string
  className: string | undefined
  shortId: string
  fallback: string
}) {
  return (
    <div className="animate-in fade-in-0 slide-in-from-top-2 fill-mode-both flex items-center justify-between pt-6 duration-500">
      <Link
        to="/org/$orgId/classes/$classId"
        params={{ orgId, classId }}
        className="text-muted-foreground hover:text-foreground group inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
      >
        <ArrowLeftIcon className="size-3.5 transition-transform group-hover:-translate-x-0.5 rtl:group-hover:translate-x-0.5" />
        <span className="max-w-[18ch] truncate">{classLabel ?? fallback}</span>
      </Link>
      <span className="text-muted-foreground/70 inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em]">
        <span className="bg-muted-foreground/30 h-px w-8" />№ {shortId || "—"}
      </span>
    </div>
  )
}

function CountdownCard({
  parts,
  status,
  accent,
  t,
}: {
  parts: ReturnType<typeof countdownParts>
  status: SessionStatus
  accent: Accent
  t: (key: string) => string
}) {
  const showDays = (parts?.days ?? 0) > 0
  const labelKey =
    status === "live" ? "status.liveNow" : status === "ended" ? "status.ended" : "org.session.meta.countdown"

  return (
    <div
      className={cn(
        "bg-card relative isolate flex flex-col gap-5 overflow-hidden rounded-3xl border p-6 shadow-sm transition-shadow md:p-7",
        "dark:bg-card/60 dark:shadow-none dark:backdrop-blur-sm",
        accent.border
      )}
    >
      <div
        aria-hidden
        className={cn("pointer-events-none absolute -end-16 -top-24 -z-10 h-64 w-64 rounded-full blur-3xl", accent.bg)}
      />
      <div className="flex items-center justify-between">
        <Eyebrow className={cn(accent.text)}>{t(labelKey)}</Eyebrow>
        <span className="relative flex size-2">
          {status === "live" ? (
            <span
              className={cn("absolute inline-flex h-full w-full animate-ping rounded-full opacity-75", accent.dot)}
            />
          ) : null}
          <span className={cn("relative inline-flex size-2 rounded-full", accent.dot)} />
        </span>
      </div>

      <div className="flex items-baseline gap-2 font-mono tabular-nums">
        {showDays ? (
          <>
            <Segment value={String(parts?.days ?? 0)} label={t("time.daysShort")} accent={accent} />
            <span className="text-muted-foreground/50 -mt-2 text-3xl font-light">·</span>
          </>
        ) : null}
        <Segment value={pad(parts?.hours ?? 0)} label={t("time.hoursShort")} accent={accent} />
        <span className="text-muted-foreground/50 -mt-2 text-3xl font-light">:</span>
        <Segment value={pad(parts?.minutes ?? 0)} label={t("time.minutesShort")} accent={accent} />
        <span className="text-muted-foreground/50 -mt-2 text-3xl font-light">:</span>
        <Segment value={pad(parts?.seconds ?? 0)} label={t("time.secondsShort")} accent={accent} muted />
      </div>

      <div className="border-border mt-1 flex items-center justify-between border-t border-dashed pt-4">
        <Eyebrow className="text-[10px]">
          {parts?.isPast ? t("org.session.meta.elapsed") : t("org.session.meta.untilStart")}
        </Eyebrow>
        <CalendarClockIcon className={cn("size-3.5", accent.text)} />
      </div>
    </div>
  )
}

function Segment({ value, label, accent, muted }: { value: string; label: string; accent: Accent; muted?: boolean }) {
  return (
    <div className="flex flex-col items-center">
      <span
        className={cn(
          "text-4xl leading-none font-semibold tracking-tight md:text-5xl",
          muted ? "text-muted-foreground" : accent.text
        )}
      >
        {value}
      </span>
      <span className="text-muted-foreground/70 mt-2 font-mono text-[10px] tracking-[0.25em] uppercase">{label}</span>
    </div>
  )
}

function MetaCell({
  label,
  value,
  mono = false,
  index,
}: {
  label: string
  value: ReactNode
  mono?: boolean
  index: number
}) {
  return (
    <div
      className="border-border animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both flex flex-col gap-2 border-b border-dashed px-5 py-5 md:border-s md:border-b-0 md:py-6 md:first:border-s-0"
      style={{ animationDelay: `${index * 80}ms`, animationDuration: "500ms" }}
    >
      <Eyebrow>{label}</Eyebrow>
      <span
        className={cn(
          "text-foreground text-base leading-tight font-medium md:text-lg",
          mono && "font-mono tabular-nums"
        )}
      >
        {value}
      </span>
    </div>
  )
}

type TileSpec = {
  key: string
  label: string
  count: number
  loading: boolean
  icon: ReactNode
  href: string
}

function RoomTile({ spec, index, total, accent }: { spec: TileSpec; index: number; total: number; accent: Accent }) {
  const tileNumber = String(index + 1).padStart(2, "0")
  const totalStr = String(total).padStart(2, "0")
  const isAnchor = spec.href.startsWith("#")
  const className = cn(
    "group/tile bg-card text-card-foreground border-border relative isolate flex h-full min-h-[180px] flex-col justify-between overflow-hidden rounded-2xl border p-5 shadow-sm transition-all duration-300",
    "hover:-translate-y-1 hover:shadow-lg hover:border-foreground/25",
    "dark:shadow-none dark:ring-1 dark:ring-foreground/8 dark:border-0 dark:hover:ring-foreground/30 dark:hover:shadow-xl dark:hover:shadow-foreground/[0.04]",
    "animate-in fade-in-0 slide-in-from-bottom-3 fill-mode-both"
  )

  const inner = (
    <>
      <div
        aria-hidden
        className={cn(
          "pointer-events-none absolute inset-0 -z-10 opacity-0 transition-opacity duration-500 group-hover/tile:opacity-100",
          "bg-[radial-gradient(circle_at_var(--mx,80%)_var(--my,20%),var(--color-primary)/10%,transparent_55%)]"
        )}
      />
      <div
        aria-hidden
        className={cn(
          "via-foreground/30 absolute inset-x-5 top-0 -z-10 h-px scale-x-0 bg-gradient-to-r from-transparent to-transparent transition-transform duration-500 group-hover/tile:scale-x-100"
        )}
      />
      <div className="flex items-start justify-between">
        <div
          className={cn(
            "bg-muted text-foreground/80 group-hover/tile:text-foreground flex size-11 items-center justify-center rounded-xl transition-colors",
            "group-hover/tile:bg-primary/10"
          )}
        >
          {spec.icon}
        </div>
        <span className="text-muted-foreground/60 font-mono text-[10px] tracking-[0.3em]">
          {tileNumber}
          <span className="opacity-50">/{totalStr}</span>
        </span>
      </div>
      <div className="mt-6 flex items-end justify-between gap-3">
        <div className="flex flex-col gap-2">
          <Eyebrow className="text-[10px]">{spec.label}</Eyebrow>
          {spec.loading ? (
            <Skeleton className="h-10 w-16" />
          ) : (
            <span className="text-4xl font-semibold tracking-tight tabular-nums md:text-5xl">{spec.count}</span>
          )}
        </div>
        <span
          className={cn(
            "text-muted-foreground group-hover/tile:bg-foreground group-hover/tile:text-background inline-flex size-8 items-center justify-center rounded-full transition-all group-hover/tile:translate-x-1 rtl:group-hover/tile:-translate-x-1",
            accent.text
          )}
        >
          <span className="rtl:rotate-180" aria-hidden>
            →
          </span>
        </span>
      </div>
    </>
  )

  const style: React.CSSProperties = {
    animationDelay: `${index * 70}ms`,
    animationDuration: "500ms",
  }

  if (isAnchor) {
    return (
      <a href={spec.href} className={className} style={style}>
        {inner}
      </a>
    )
  }
  return (
    <Link to={spec.href} className={className} style={style}>
      {inner}
    </Link>
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
          const items = roomsData?.items ?? []
          const room = items.find((r) => r.class_session_id === variables.data.class_session_id)
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
      className={cn("relative overflow-hidden", "shadow-[0_8px_24px_-12px_var(--color-primary)]")}
      disabled={join.isPending || !session.id}
      onClick={() => session.id && join.mutate({ data: { class_session_id: session.id } })}
    >
      <span className={cn("absolute inset-0 -z-10 opacity-30 blur-xl", accent.bg)} aria-hidden />
      <RadioIcon className="size-4" />
      {canStart ? t("org.session.actions.start") : t("org.session.actions.join")}
    </Button>
  )
}

type WorkspaceTab = {
  key: string
  label: string
  count: number
  loading: boolean
  icon: ReactNode
  content: ReactNode
}

function WorkspaceSection({
  eyebrow,
  title,
  subtitle,
  tabs,
  accent,
}: {
  eyebrow: string
  title: string
  subtitle: string
  tabs: WorkspaceTab[]
  accent: Accent
}) {
  if (tabs.length === 0) return null
  const count = tabs.length
  const defaultTab = tabs[0]?.key

  return (
    <section className="flex flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="flex flex-col gap-2">
          <Eyebrow>{eyebrow}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight md:text-3xl">{title}</h2>
          <p className="text-muted-foreground max-w-xl text-sm leading-relaxed">{subtitle}</p>
        </div>
        <Eyebrow className="text-muted-foreground/70 hidden items-center gap-2 font-mono md:inline-flex">
          <span className="bg-muted-foreground/30 h-px w-6" />
          {String(count).padStart(2, "0")} {count === 1 ? "view" : "views"}
        </Eyebrow>
      </div>

      {count === 1 ? (
        <div>{tabs[0]!.content}</div>
      ) : (
        <Tabs defaultValue={defaultTab} className="gap-6">
          <div className="bg-card border-border dark:bg-card/40 dark:ring-foreground/10 rounded-2xl border p-1.5 shadow-sm dark:border-0 dark:shadow-none dark:ring-1">
            <TabsList variant="line" className="h-auto w-full gap-1 bg-transparent p-0">
              {tabs.map((tab) => (
                <TabsTrigger
                  key={tab.key}
                  value={tab.key}
                  className={cn(
                    "group/wstab flex-1 justify-start gap-2.5 rounded-xl px-4 py-3",
                    "data-active:bg-muted data-active:text-foreground dark:data-active:bg-foreground/5"
                  )}
                >
                  <span
                    className={cn(
                      "bg-muted text-muted-foreground group-data-[active]/wstab:bg-foreground group-data-[active]/wstab:text-background flex size-8 items-center justify-center rounded-lg transition-colors"
                    )}
                  >
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
      )}
    </section>
  )
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { orgId, classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView: canViewBanks } = useBankPermissions()
  const { canView: canViewQuizzes, canEdit: canGradeQuizzes } = useQuizPermissions()
  const { canView: canViewLive, canJoin: canJoinLive } = useLivesessionPermissions()
  const { canView: canViewPractices, canGrade: canGradePractices } = usePracticePermissions()
  const { canView: canViewOfflines } = useOfflinePermissions()
  const { canView: canViewAttendance } = useAttendancePermissions()
  const canViewLiveAny = canViewLive || canJoinLive
  const now = useNow(1000)

  const {
    data: sessionData,
    isPending: sessionPending,
    isError: sessionError,
  } = useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id

  const { data: classData } = useGetClassesId(classId ?? "", {
    query: { enabled: !!classId },
  })
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
      <div className="flex flex-col gap-10 py-10">
        <Skeleton className="h-5 w-40" />
        <div className="grid gap-6 lg:grid-cols-5">
          <div className="flex flex-col gap-5 lg:col-span-3">
            <Skeleton className="h-6 w-28" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-10 w-2/3" />
            <Skeleton className="h-11 w-44" />
          </div>
          <Skeleton className="h-52 w-full lg:col-span-2" />
        </div>
        <Skeleton className="h-24 w-full" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Skeleton className="h-44 w-full" />
          <Skeleton className="h-44 w-full" />
          <Skeleton className="h-44 w-full" />
          <Skeleton className="h-44 w-full" />
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
  const startStr = formatSessionDate(session.start_time, i18n.language, "long")
  const createdStr = formatSessionDate(session.created_at, i18n.language, "long")
  const parts = countdownParts(session.start_time, now)
  const shortId = (session.id ?? "").slice(0, 8).toUpperCase()
  const classPath = `/org/${orgId}/classes/${classId ?? ""}`

  const navTiles: TileSpec[] = []
  if (canViewLiveAny) {
    navTiles.push({
      key: "live",
      label: t("org.session.rooms.live"),
      count: itemsCount(liveQ.data),
      loading: liveQ.isPending,
      icon: <VideoIcon className="size-5" />,
      href: "#live-rooms",
    })
  }
  if (canViewPractices) {
    navTiles.push({
      key: "practices",
      label: t("org.session.rooms.practices"),
      count: itemsCount(practiceQ.data),
      loading: practiceQ.isPending,
      icon: <DumbbellIcon className="size-5" />,
      href: classPath,
    })
  }
  if (canViewOfflines) {
    navTiles.push({
      key: "offlines",
      label: t("org.session.rooms.offlines"),
      count: itemsCount(offlineQ.data),
      loading: offlineQ.isPending,
      icon: <FilmIcon className="size-5" />,
      href: "#offlines",
    })
  }
  const navCount = navTiles.length
  const navGridCols = navCount >= 3 ? "md:grid-cols-3" : navCount === 2 ? "md:grid-cols-2" : "md:grid-cols-1"

  const workspaceTabs: WorkspaceTab[] = []
  if (canViewQuizzes && classId) {
    workspaceTabs.push({
      key: "quizzes",
      label: t("org.session.workspace.tabs.quizzes"),
      count: itemsCount(quizQ.data),
      loading: quizQ.isPending,
      icon: <ClipboardListIcon className="size-4" />,
      content: <QuizzesSection classId={classId} classSessionId={classSessionId} />,
    })
  }
  if (canGradeQuizzes) {
    workspaceTabs.push({
      key: "corrections",
      label: t("org.session.workspace.tabs.corrections"),
      count: itemsCount(quizQ.data),
      loading: quizQ.isPending,
      icon: <CheckSquareIcon className="size-4" />,
      content: <QuizCorrectionsSection classSessionId={classSessionId} />,
    })
  }
  if (canViewBanks) {
    workspaceTabs.push({
      key: "banks",
      label: t("org.session.workspace.tabs.banks"),
      count: itemsCount(banksQ.data),
      loading: banksQ.isPending,
      icon: <LibraryIcon className="size-4" />,
      content: <QuestionBanksSection />,
    })
  }

  const practiceTabs: WorkspaceTab[] = []
  if (canViewPractices) {
    practiceTabs.push({
      key: "practices",
      label: t("org.session.practiceWorkspace.tabs.practices"),
      count: itemsCount(practiceQ.data),
      loading: practiceQ.isPending,
      icon: <DumbbellIcon className="size-4" />,
      content: <PracticesSection classSessionId={classSessionId} />,
    })
  }
  if (canGradePractices) {
    practiceTabs.push({
      key: "practiceScores",
      label: t("org.session.practiceWorkspace.tabs.practiceScores"),
      count: itemsCount(practiceQ.data),
      loading: practiceQ.isPending,
      icon: <CheckSquareIcon className="size-4" />,
      content: <PracticeScoresSection classSessionId={classSessionId} />,
    })
  }

  const offlineTabs: WorkspaceTab[] = []
  if (canViewOfflines) {
    offlineTabs.push({
      key: "offlines",
      label: t("org.session.offlineWorkspace.tabs.offlines"),
      count: itemsCount(offlineQ.data),
      loading: offlineQ.isPending,
      icon: <FilmIcon className="size-4" />,
      content: <OfflinesSection classSessionId={classSessionId} orgId={orgId} />,
    })
  }

  const liveTabs: WorkspaceTab[] = []
  if (canViewLiveAny) {
    liveTabs.push({
      key: "rooms",
      label: t("org.session.liveWorkspace.tabs.rooms"),
      count: itemsCount(liveQ.data),
      loading: liveQ.isPending,
      icon: <VideoIcon className="size-4" />,
      content: <LiveRoomsSection classSessionId={classSessionId} />,
    })
  }
  if (canViewAttendance && classId) {
    liveTabs.push({
      key: "presence",
      label: t("org.session.liveWorkspace.tabs.presence"),
      count: itemsCount(attendanceQ.data),
      loading: attendanceQ.isPending,
      icon: <UserCheckIcon className="size-4" />,
      content: <AttendanceSection classId={classId} classSessionId={classSessionId} />,
    })
  }

  return (
    <div className="relative isolate flex flex-col gap-12 pb-16">
      <DecorativeBackground accent={accent} />

      <Breadcrumb
        orgId={orgId}
        classId={classId ?? ""}
        className={cls?.name}
        shortId={shortId}
        fallback={t("org.session.backToClass")}
      />

      <section className="grid gap-8 lg:grid-cols-5 lg:gap-10">
        <header className="animate-in fade-in-0 slide-in-from-bottom-3 fill-mode-both flex flex-col gap-6 duration-700 lg:col-span-3">
          <div className="flex flex-wrap items-center gap-3">
            <SessionStatusPill status={status} />
            <span className="bg-foreground/15 h-px w-8" aria-hidden />
            <Eyebrow>{t("org.session.eyebrow")}</Eyebrow>
          </div>

          <h1 className="max-w-4xl text-4xl leading-[1.05] font-semibold tracking-tight text-balance md:text-5xl lg:text-[3.5rem] xl:text-6xl">
            {session.name}
          </h1>

          {session.description ? (
            <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">
              {session.description}
            </p>
          ) : null}

          <div className="mt-2 flex flex-wrap items-center gap-3">
            <JoinAction session={session} accent={accent} />
            <Button
              variant="outline"
              render={<Link to="/org/$orgId/classes/$classId" params={{ orgId, classId: classId ?? "" }} />}
            >
              <SparklesIcon className="size-4" />
              {t("org.session.actions.viewClass")}
            </Button>
          </div>
        </header>

        <div
          className="animate-in fade-in-0 slide-in-from-bottom-3 fill-mode-both duration-700 lg:col-span-2"
          style={{ animationDelay: "120ms" }}
        >
          <CountdownCard parts={parts} status={status} accent={accent} t={t} />
        </div>
      </section>

      <section
        className={cn(
          "bg-card border-border grid grid-cols-1 overflow-hidden rounded-2xl border shadow-sm md:grid-cols-4",
          "dark:bg-card/40 dark:ring-foreground/8 dark:border-0 dark:shadow-none dark:ring-1 dark:backdrop-blur-sm"
        )}
      >
        <MetaCell index={0} label={t("org.session.meta.starts")} value={startStr} />
        <MetaCell
          index={1}
          label={t("org.session.meta.status")}
          value={t(`status.${status === "live" ? "liveNow" : status}`)}
        />
        <MetaCell
          index={2}
          label={t("org.session.meta.countdown")}
          mono
          value={
            parts
              ? parts.days > 0
                ? `${parts.days}d ${pad(parts.hours)}:${pad(parts.minutes)}:${pad(parts.seconds)}`
                : `${pad(parts.hours)}:${pad(parts.minutes)}:${pad(parts.seconds)}`
              : "—"
          }
        />
        <MetaCell index={3} label={t("org.session.meta.created")} value={createdStr} />
      </section>

      {navCount > 0 ? (
        <section className="flex flex-col gap-4">
          <div className="flex items-end justify-between">
            <Eyebrow>{t("org.session.rooms.eyebrow")}</Eyebrow>
            <Eyebrow className="text-muted-foreground/70 hidden items-center gap-2 md:inline-flex">
              <span className="bg-muted-foreground/30 h-px w-6" />
              {t("org.session.rooms.subtitle")}
            </Eyebrow>
          </div>
          <div className={cn("grid grid-cols-1 gap-3 sm:grid-cols-2", navGridCols)}>
            {navTiles.map((spec, i) => (
              <RoomTile key={spec.key} spec={spec} index={i} total={navCount} accent={accent} />
            ))}
          </div>
        </section>
      ) : null}

      <WorkspaceSection
        eyebrow={t("org.session.liveWorkspace.eyebrow")}
        title={t("org.session.liveWorkspace.title")}
        subtitle={t("org.session.liveWorkspace.subtitle")}
        tabs={liveTabs}
        accent={accent}
      />

      <WorkspaceSection
        eyebrow={t("org.session.workspace.eyebrow")}
        title={t("org.session.workspace.title")}
        subtitle={t("org.session.workspace.subtitle")}
        tabs={workspaceTabs}
        accent={accent}
      />

      <WorkspaceSection
        eyebrow={t("org.session.practiceWorkspace.eyebrow")}
        title={t("org.session.practiceWorkspace.title")}
        subtitle={t("org.session.practiceWorkspace.subtitle")}
        tabs={practiceTabs}
        accent={accent}
      />

      <WorkspaceSection
        eyebrow={t("org.session.offlineWorkspace.eyebrow")}
        title={t("org.session.offlineWorkspace.title")}
        subtitle={t("org.session.offlineWorkspace.subtitle")}
        tabs={offlineTabs}
        accent={accent}
      />

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

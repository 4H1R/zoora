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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatRelativeTime, formatSessionDate, getSessionStatus, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/classes/class-sessions/$classSessionId")({
  head: () => orgHead("org.session.title"),
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

function JoinAction({ session }: { session: Session }) {
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
      <Button variant="outline" size="sm" disabled>
        {t("org.session.actions.notPermitted")}
      </Button>
    )
  }

  return (
    <Button
      size="sm"
      disabled={join.isPending || !session.id}
      onClick={() => session.id && join.mutate({ data: { class_session_id: session.id } })}
    >
      <RadioIcon className="size-4" />
      {canStart ? t("org.session.actions.start") : t("org.session.actions.join")}
    </Button>
  )
}

// Light line sub-tabs, consistent with the class detail page's TabsList.
function SubTabs({ tabs, status }: { tabs: SubTab[]; status: SessionStatus }) {
  if (tabs.length === 0) return null
  if (tabs.length === 1) return <div className="flex flex-col gap-6">{tabs[0]!.content}</div>

  return (
    <Tabs defaultValue={tabs[0]?.key} className="gap-6">
      <TabsList variant="line">
        {tabs.map((tab) => (
          <TabsTrigger key={tab.key} value={tab.key} className="gap-2">
            {tab.icon}
            <span>{tab.label}</span>
            <CountBadge count={tab.count} loading={tab.loading} status={status} />
          </TabsTrigger>
        ))}
      </TabsList>
      {tabs.map((tab) => (
        <TabsContent key={tab.key} value={tab.key} className="flex flex-col gap-6">
          {tab.content}
        </TabsContent>
      ))}
    </Tabs>
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
  const now = useNow(1000)

  const [tab, setTab] = useState<string | null>(null)

  const {
    data: sessionData,
    isPending: sessionPending,
    isError: sessionError,
  } = useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id

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

  const status = getSessionStatus(session.start_time, now)
  const startStr = formatSessionDate(session.start_time, i18n.language, "long")
  const relativeStr = formatRelativeTime(session.start_time, now, i18n.language)

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
      key: "recordings",
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

  const activeTab = tab ?? surfaces[0]?.key

  return (
    <div className="relative isolate flex flex-col gap-6 pb-10">
      <DecorativeBackground />


      <header className="border-foreground/10 bg-card/50 relative flex flex-col gap-5 overflow-hidden rounded-2xl border p-4 backdrop-blur-sm md:flex-row md:items-start md:justify-between md:gap-8 md:p-5">
        <div className="flex min-w-0 flex-col gap-2.5">
          <div className="flex flex-wrap items-center gap-2.5">
            <Eyebrow>{t("org.session.eyebrow")}</Eyebrow>
            <SessionStatusPill status={status} size="sm" />
          </div>

          <h1 className="max-w-2xl text-2xl leading-tight font-semibold tracking-tight text-balance md:text-3xl">
            {session.name}
          </h1>

          {session.description ? (
            <p className="text-muted-foreground line-clamp-2 max-w-xl text-sm leading-relaxed">
              {session.description}
            </p>
          ) : null}

          <div className="text-muted-foreground mt-0.5 flex flex-wrap items-center gap-x-2.5 gap-y-1 text-sm">
            <span className="inline-flex items-center gap-1.5">
              <CalendarClockIcon className="size-3.5" />
              {startStr}
            </span>
            {status !== "ended" && relativeStr ? (
              <>
                <span className="text-muted-foreground/40">·</span>
                <span className={cn("font-medium", status === "live" ? "text-destructive" : "text-foreground")}>
                  {relativeStr}
                </span>
              </>
            ) : null}
            {cls?.name ? (
              <>
                <span className="text-muted-foreground/40">·</span>
                <Link
                  to="/org/classes/$classId"
                  params={{ classId: classId ?? "" }}
                  className="hover:text-foreground inline-flex max-w-[22ch] items-center gap-1.5 truncate transition-colors"
                >
                  <SparklesIcon className="size-3.5" />
                  {cls.name}
                </Link>
              </>
            ) : null}
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-2 max-md:self-start">
          <JoinAction session={session} />
        </div>
      </header>

      {surfaces.length > 0 && activeTab ? (
        <Tabs value={activeTab} onValueChange={setTab}>
          {/* Primary nav: a filled segmented control with icons. The distinct
              filled treatment (vs. the line sub-tabs below) makes the two-tier
              hierarchy legible at a glance. */}
          <TabsList variant="default" className="h-auto flex-wrap p-1">
            {surfaces.map((surface) => (
              <TabsTrigger key={surface.key} value={surface.key} className="gap-2 px-3 py-1.5">
                {surface.icon}
                <span>{surface.navLabel}</span>
                <CountBadge count={surface.count} loading={surface.loading} status={status} />
              </TabsTrigger>
            ))}
          </TabsList>

          {surfaces.map((surface) => (
            <TabsContent key={surface.key} value={surface.key} className="flex flex-col gap-6">
              <SubTabs tabs={surface.subTabs} status={status} />
            </TabsContent>
          ))}
        </Tabs>
      ) : null}
    </div>
  )
}

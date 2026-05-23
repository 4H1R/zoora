import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"

import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CalendarClockIcon,
  ClipboardListIcon,
  DumbbellIcon,
  FilmIcon,
  LibraryIcon,
  RadioIcon,
  SparklesIcon,
  VideoIcon,
} from "lucide-react"
import type { ReactNode } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetClassesId, useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { getLiveRooms, useGetLiveRooms, usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { useGetOfflines } from "@/api/offlines/offlines"
import { useGetPractices } from "@/api/practices/practices"
import { useGetQuestionBanks } from "@/api/question-banks/question-banks"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { QuestionBanksSection } from "@/components/org/question-banks/QuestionBanksSection"
import { useBankPermissions } from "@/components/org/question-banks/use-bank-permissions"
import { QuizzesSection } from "@/components/org/quizzes/QuizzesSection"
import { useQuizPermissions } from "@/components/org/quizzes/use-quiz-permissions"
import { Eyebrow } from "@/components/eyebrow"
import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import {
  formatCountdown,
  formatSessionDate,
  getSessionStatus,
  useNow,
} from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/classes/classsessions/$classSessionId")({
  head: () => orgHead("org.session.title"),
  component: RouteComponent,
})

function MetaCell({ label, value, mono = false }: { label: string; value: ReactNode; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-2 border-b border-dashed py-5 pe-4 ps-4 md:border-b-0 md:border-s md:py-0 md:first:border-s-0 md:first:ps-0">
      <Eyebrow>{label}</Eyebrow>
      <span className={cn("text-foreground text-base leading-tight font-medium md:text-lg", mono && "font-mono tabular-nums")}>
        {value}
      </span>
    </div>
  )
}

function RoomTile({
  label,
  count,
  icon,
  href,
  index,
  loading,
}: {
  label: string
  count: number
  icon: ReactNode
  href: string
  index: number
  loading: boolean
}) {
  const tileNumber = String(index + 1).padStart(2, "0")
  const className =
    "group/tile bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col justify-between overflow-hidden rounded-2xl p-5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg"
  const isAnchor = href.startsWith("#")
  const Wrapper = ({ children }: { children: ReactNode }) =>
    isAnchor ? (
      <a href={href} className={className}>
        {children}
      </a>
    ) : (
      <Link to={href} className={className}>
        {children}
      </Link>
    )
  return (
    <Wrapper>
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/tile:opacity-100"
      />
      <div className="flex items-start justify-between">
        <div className="bg-muted text-foreground flex size-10 items-center justify-center rounded-xl">{icon}</div>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>
      <div className="mt-8 flex items-end justify-between gap-3">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{label}</Eyebrow>
          {loading ? (
            <Skeleton className="h-9 w-14" />
          ) : (
            <span className="text-4xl font-semibold tracking-tight tabular-nums">{count}</span>
          )}
        </div>
        <span className="text-muted-foreground group-hover/tile:text-foreground text-xs font-medium underline-offset-4 transition-colors group-hover/tile:underline">
          →
        </span>
      </div>
    </Wrapper>
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
        className="pointer-events-none absolute inset-0 -z-10 opacity-[0.05] [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [background-size:48px_48px] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)]"
      />
    </>
  )
}

function JoinAction({ session }: { session: Session }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { can } = useAccess()
  const canStart = can("livesessions:create") || can("livesessions:create_any")
  const canJoin = can("livesessions:join") || can("livesessions:join_any")

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
      disabled={join.isPending || !session.id}
      onClick={() => session.id && join.mutate({ data: { class_session_id: session.id } })}
    >
      <RadioIcon className="size-4" />
      {canStart ? t("org.session.actions.start") : t("org.session.actions.join")}
    </Button>
  )
}

function itemsCount(payload: unknown): number {
  const p = payload as { status?: number; data?: { data?: { items?: unknown[] } } } | undefined
  if (!p || p.status !== 200) return 0
  return p.data?.data?.items?.length ?? 0
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { orgId, classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView: canViewBanks } = useBankPermissions()
  const { canView: canViewQuizzes } = useQuizPermissions()
  const now = useNow(1000)

  const { data: sessionData, isPending: sessionPending, isError: sessionError } =
    useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id

  const { data: classData } = useGetClassesId(classId ?? "", {
    query: { enabled: !!classId },
  })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  const enabled = !!session
  const liveQ = useGetLiveRooms({ class_session_id: classSessionId }, { query: { enabled } })
  const quizQ = useGetQuizzes(
    { class_session_id: classSessionId },
    { query: { enabled: enabled && canViewQuizzes } }
  )
  const practiceQ = useGetPractices({ class_session_id: classSessionId }, { query: { enabled } })
  const offlineQ = useGetOfflines({ class_session_id: classSessionId }, { query: { enabled } })
  const banksQ = useGetQuestionBanks(undefined, { query: { enabled: enabled && canViewBanks } })

  if (!allowed) return null

  if (sessionPending) {
    return (
      <div className="flex flex-col gap-10 py-10">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-16 w-3/4" />
        <Skeleton className="h-32 w-full" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Skeleton className="h-40 w-full" />
          <Skeleton className="h-40 w-full" />
          <Skeleton className="h-40 w-full" />
          <Skeleton className="h-40 w-full" />
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
  const startStr = formatSessionDate(session.start_time, i18n.language, "long")
  const createdStr = formatSessionDate(session.created_at, i18n.language, "long")
  const countdown = formatCountdown(session.start_time, now)
  const shortId = (session.id ?? "").slice(0, 8).toUpperCase()
  const classPath = `/org/${orgId}/classes/${classId ?? ""}`

  const visibleTileCount = 3 + (canViewQuizzes ? 1 : 0) + (canViewBanks ? 1 : 0)
  const roomGridCols =
    visibleTileCount >= 5 ? "xl:grid-cols-5" : visibleTileCount === 4 ? "xl:grid-cols-4" : "xl:grid-cols-3"

  return (
    <div className="relative isolate flex flex-col gap-10 pb-16">
      <DecorativeBackground />

      <div className="flex items-center justify-between pt-6">
        <Link
          to="/org/$orgId/classes/$classId"
          params={{ orgId, classId: classId ?? "" }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {cls?.name ?? t("org.session.backToClass")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">№ {shortId || "—"}</span>
      </div>

      <header className="flex flex-col gap-5">
        <div className="flex flex-wrap items-center gap-3">
          <SessionStatusPill status={status} />
          <Eyebrow>{t("org.session.eyebrow")}</Eyebrow>
        </div>

        <h1 className="max-w-4xl text-4xl leading-tight font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          {session.name}
        </h1>

        {session.description ? (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">{session.description}</p>
        ) : null}

        <div className="mt-2 flex flex-wrap items-center gap-3">
          <JoinAction session={session} />
          <Button
            variant="outline"
            render={<Link to="/org/$orgId/classes/$classId" params={{ orgId, classId: classId ?? "" }} />}
          >
            <SparklesIcon className="size-4" />
            {t("org.session.actions.viewClass")}
          </Button>
        </div>
      </header>

      <section className="bg-card ring-foreground/10 grid grid-cols-1 overflow-hidden rounded-2xl px-4 py-6 ring-1 md:grid-cols-4 md:px-6">
        <MetaCell label={t("org.session.meta.starts")} value={startStr} />
        <MetaCell label={t("org.session.meta.countdown")} value={countdown} mono />
        <MetaCell label={t("org.session.meta.status")} value={t(`status.${status === "live" ? "liveNow" : status}`)} />
        <MetaCell label={t("org.session.meta.created")} value={createdStr} />
      </section>

      <section className="flex flex-col gap-5">
        <div className="flex items-end justify-between">
          <div className="flex flex-col gap-1.5">
            <Eyebrow>{t("org.session.rooms.eyebrow")}</Eyebrow>
            <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.rooms.title")}</h2>
          </div>
          <Eyebrow className="hidden md:inline">{t("org.session.rooms.subtitle")}</Eyebrow>
        </div>

        <div className={cn("grid gap-4 md:grid-cols-2", roomGridCols)}>
          <RoomTile
            index={0}
            label={t("org.session.rooms.live")}
            count={itemsCount(liveQ.data)}
            loading={liveQ.isPending}
            icon={<VideoIcon className="size-5" />}
            href={classPath}
          />
          {canViewQuizzes ? (
            <RoomTile
              index={1}
              label={t("org.session.rooms.quizzes")}
              count={itemsCount(quizQ.data)}
              loading={quizQ.isPending}
              icon={<ClipboardListIcon className="size-5" />}
              href="#quizzes"
            />
          ) : null}
          <RoomTile
            index={2}
            label={t("org.session.rooms.practices")}
            count={itemsCount(practiceQ.data)}
            loading={practiceQ.isPending}
            icon={<DumbbellIcon className="size-5" />}
            href={classPath}
          />
          <RoomTile
            index={3}
            label={t("org.session.rooms.offlines")}
            count={itemsCount(offlineQ.data)}
            loading={offlineQ.isPending}
            icon={<FilmIcon className="size-5" />}
            href={classPath}
          />
          {canViewBanks ? (
            <RoomTile
              index={4}
              label={t("org.session.rooms.questionBanks")}
              count={itemsCount(banksQ.data)}
              loading={banksQ.isPending}
              icon={<LibraryIcon className="size-5" />}
              href="#question-banks"
            />
          ) : null}
        </div>
      </section>

      {classId ? <QuizzesSection classId={classId} classSessionId={classSessionId} /> : null}

      <QuestionBanksSection />

      <footer className="border-t border-dashed pt-6">
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

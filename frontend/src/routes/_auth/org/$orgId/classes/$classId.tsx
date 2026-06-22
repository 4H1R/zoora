import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import {
  ArrowLeftIcon,
  CalendarClockIcon,
  PlusIcon,
  TrophyIcon,
  UserIcon,
  UsersIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useAccess } from "react-access-engine"

import {
  useGetClassesId,
  useGetClassesIdMembers,
  useGetClassesIdSessions,
} from "@/api/classes/classes"
import { SessionCreateModal } from "@/components/admin/sessions/SessionCreateModal"
import { useClassPermissions } from "@/components/org/classes/use-class-permissions"
import { Eyebrow } from "@/components/eyebrow"
import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate, getSessionStatus, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/classes/$classId")({
  head: () => orgHead("org.class.title"),
  component: RouteComponent,
})

function SessionCard({
  session,
  orgId,
  index,
  now,
}: {
  session: Session
  orgId: string
  index: number
  now: number
}) {
  const { t, i18n } = useTranslation()
  const status = getSessionStatus(session.start_time, now)
  const tileNumber = String(index + 1).padStart(2, "0")
  const startStr = formatSessionDate(session.start_time, i18n.language, "short")

  return (
    <Link
      to="/org/$orgId/classes/classsessions/$classSessionId"
      params={{ orgId, classSessionId: session.id! }}
      className="group/tile bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg"
    >
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/tile:opacity-100"
      />

      <div className="flex items-start justify-between gap-2">
        <SessionStatusPill status={status} size="sm" />
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-2">
        <Eyebrow>{t("org.class.sessions.eyebrow")}</Eyebrow>
        <h3 className="line-clamp-2 text-xl leading-snug font-semibold tracking-tight text-balance">
          {session.name}
        </h3>
        {session.description ? (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">{session.description}</p>
        ) : null}
      </div>

      <div className="mt-auto flex items-center justify-between gap-3 border-t border-dashed pt-4">
        <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs">
          <CalendarClockIcon className="size-3.5" />
          {startStr}
        </span>
        <span className="text-muted-foreground group-hover/tile:text-foreground text-xs font-medium underline-offset-4 transition-colors group-hover/tile:underline">
          {t("org.class.sessions.open")} →
        </span>
      </div>
    </Link>
  )
}

function SessionCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex flex-col gap-5 rounded-2xl p-5 ring-1">
      <div className="flex items-center justify-between">
        <Skeleton className="h-5 w-20" />
        <Skeleton className="h-3 w-8" />
      </div>
      <div className="flex flex-col gap-2">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-6 w-4/5" />
        <Skeleton className="h-3 w-3/5" />
      </div>
      <div className="flex items-center justify-between border-t border-dashed pt-4">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-3 w-12" />
      </div>
    </div>
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
        className="pointer-events-none absolute inset-0 -z-10 opacity-[0.04] [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [background-size:48px_48px] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)]"
      />
    </>
  )
}

function StatCell({ label, value, accent }: { label: string; value: number; accent?: boolean }) {
  return (
    <div className="flex flex-col gap-2 px-5 py-5">
      <Eyebrow>{label}</Eyebrow>
      <span
        className={cn(
          "text-3xl font-semibold tracking-tight tabular-nums",
          accent ? "text-destructive" : "text-foreground"
        )}
      >
        {value}
      </span>
    </div>
  )
}

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId, classId } = Route.useParams()
  const { canView, canEdit: canCreateSession } = useClassPermissions()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const now = useNow(30_000)
  const { can, user: accessUser } = useAccess()

  const [formOpen, setFormOpen] = useState(false)

  const { data: classData, isPending: classPending } = useGetClassesId(classId, {
    query: { enabled: canView },
  })
  const { data: sessionsData, isPending: sessionsPending } = useGetClassesIdSessions(
    classId,
    undefined,
    { query: { enabled: canView } }
  )

  const cls = (classData?.status === 200 && classData.data.data) || undefined
  const sessionsResult = (sessionsData?.status === 200 && sessionsData.data.data) || undefined
  const sessions = sessionsResult?.items ?? []
  const total = sessionsResult?.total ?? sessions.length

  // Roster gating mirrors backend canManageClass:
  // admin OR classes:update_any OR caller is class owner.
  const canViewRoster =
    !!cls &&
    (can("classes:update_any") || (!!cls.user_id && cls.user_id === accessUser.id))

  const { data: membersData } = useGetClassesIdMembers(classId, undefined, {
    query: { enabled: canView && canViewRoster },
  })
  const membersResult = (membersData?.status === 200 && membersData.data.data) || undefined
  const studentsTotal = membersResult?.total ?? membersResult?.items?.length ?? 0

  const liveCount = sessions.filter((s) => getSessionStatus(s.start_time, now) === "live").length
  const scheduledCount = sessions.filter((s) => getSessionStatus(s.start_time, now) === "scheduled").length

  if (!allowed) return null

  if (classPending) {
    return (
      <div className="flex flex-col gap-10 py-10">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-16 w-3/4" />
        <Skeleton className="h-20 w-full" />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <SessionCardSkeleton />
          <SessionCardSkeleton />
          <SessionCardSkeleton />
        </div>
      </div>
    )
  }

  const teacherName = cls?.user?.name ?? ""

  return (
    <div className="relative isolate flex flex-col gap-10 pb-16">
      <DecorativeBackground />

      <div className="flex items-center justify-between pt-6">
        <Link
          to="/org/$orgId/classes"
          params={{ orgId }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.class.backToClasses")}
        </Link>
        <Button
          variant="outline"
          size="sm"
          render={<Link to="/org/$orgId/classes/$classId/gradebook" params={{ orgId, classId }} />}
        >
          <TrophyIcon className="size-4" />
          {t("org.class.gradebook.open")}
        </Button>
      </div>

      <header className="flex flex-col gap-5">
        <Eyebrow>{t("org.class.eyebrow")}</Eyebrow>

        <h1 className="max-w-4xl text-4xl leading-tight font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          {cls?.name ?? "—"}
        </h1>

        {cls?.description ? (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">{cls.description}</p>
        ) : null}

        {cls?.user_id ? (
          <div className="text-muted-foreground inline-flex items-center gap-2 text-sm">
            {teacherName ? <UserAvatar name={teacherName} size="md" /> : <UserIcon className="size-4" />}
            <span className="text-foreground font-medium">{teacherName || t("org.class.unknownTeacher")}</span>
            <Eyebrow className="text-muted-foreground">{t("org.class.instructor")}</Eyebrow>
          </div>
        ) : null}
      </header>

      <section className="bg-card ring-foreground/10 grid grid-cols-3 overflow-hidden rounded-2xl ring-1 divide-x divide-dashed rtl:divide-x-reverse">
        <StatCell label={t("org.class.stats.total")} value={total} />
        <StatCell label={t("org.class.stats.live")} value={liveCount} accent={liveCount > 0} />
        <StatCell label={t("org.class.stats.upcoming")} value={scheduledCount} />
      </section>

      <section className="flex flex-col gap-5">
        <div className="flex items-end justify-between gap-4">
          <div className="flex flex-col gap-1.5">
            <Eyebrow>{t("org.class.sessions.eyebrow")}</Eyebrow>
            <h2 className="text-2xl font-semibold tracking-tight">{t("org.class.sessions.title")}</h2>
          </div>

          {canCreateSession ? (
            <Button onClick={() => setFormOpen(true)}>
              <PlusIcon className="size-4" />
              {t("org.class.sessions.newSession")}
            </Button>
          ) : null}
        </div>

        {sessionsPending ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <SessionCardSkeleton />
            <SessionCardSkeleton />
            <SessionCardSkeleton />
          </div>
        ) : sessions.length === 0 ? (
          <EmptyState
            icon={CalendarClockIcon}
            title={t("org.class.sessions.emptyTitle")}
            description={t("org.class.sessions.emptyHint")}
          >
            {canCreateSession ? (
              <Button onClick={() => setFormOpen(true)}>
                <PlusIcon className="size-4" />
                {t("org.class.sessions.newSession")}
              </Button>
            ) : null}
          </EmptyState>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {sessions.map((s, i) => (
              <SessionCard key={s.id} session={s} orgId={orgId} index={i} now={now} />
            ))}
          </div>
        )}
      </section>

      {canViewRoster ? (
        <section className="flex flex-col gap-5">
          <div className="flex flex-col gap-1.5">
            <Eyebrow>{t("org.class.students.eyebrow")}</Eyebrow>
            <h2 className="text-2xl font-semibold tracking-tight">{t("org.class.students.title")}</h2>
          </div>

          <Link
            to="/org/$orgId/classes/$classId/students"
            params={{ orgId, classId }}
            className="group/roster bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex items-center gap-4 overflow-hidden rounded-2xl p-5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg"
          >
            <span className="bg-primary/10 text-primary inline-flex size-12 shrink-0 items-center justify-center rounded-full">
              <UsersIcon className="size-5" />
            </span>
            <div className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span className="text-sm font-semibold tracking-tight">
                {t("org.class.students.manageTitle")}
              </span>
              <span className="text-muted-foreground text-xs">
                {t("org.class.students.manageHint")}
              </span>
            </div>
            <span className="text-muted-foreground font-mono text-lg font-semibold tabular-nums">
              {studentsTotal}
            </span>
            <span className="text-muted-foreground group-hover/roster:text-foreground text-xs font-medium underline-offset-4 transition-colors group-hover/roster:underline">
              {t("org.class.students.open")} →
            </span>
          </Link>
        </section>
      ) : null}

      {canCreateSession ? (
        <SessionCreateModal open={formOpen} onOpenChange={setFormOpen} classId={classId} session={null} />
      ) : null}
    </div>
  )
}

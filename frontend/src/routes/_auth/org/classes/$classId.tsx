import type {
  GithubCom4H1RZooraInternalDomainClassMember as ClassMember,
  GithubCom4H1RZooraInternalDomainClassSession as Session,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { CalendarClockIcon, PlusIcon, TrophyIcon, UserIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetClassesIdMembersQueryKey,
  useDeleteClassesIdMembersUserId,
  useGetClassesId,
  useGetClassesIdMembers,
  useGetClassesIdSessions,
} from "@/api/classes/classes"
import { SessionCreateModal } from "@/components/admin/sessions/SessionCreateModal"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { AttendanceMatrixView } from "@/components/org/classes/AttendanceMatrixView"
import { EnrollMemberModal } from "@/components/org/classes/EnrollMemberModal"
import { useClassPermissions } from "@/components/org/classes/use-class-permissions"
import { useAttendancePermissions } from "@/components/org/livesessions/use-attendance-permissions"
import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { UserAvatar } from "@/components/user-avatar"
import { ViewModeToggle, useViewMode } from "@/components/view-mode-toggle"
import { useOrgGuard } from "@/lib/access"
import { useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { formatRelativeTime, formatSessionDate, getSessionStatus, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

import { useSessionColumns, useStudentColumns } from "./-detail-columns"

// One tab is visible at a time, so the three views share a single set of
// list params (bare search/order_by/order_dir/page/page_size). `tab` selects
// the active view; switching tabs resets the shared params (see
// handleTabChange) so a stale sort token never leaks across tabs.
const CLASS_TABS = ["sessions", "students", "attendance"] as const
type ClassTab = (typeof CLASS_TABS)[number]

const classDetailSearchSchema = z.object({
  tab: z.enum(CLASS_TABS).optional().default("sessions"),
  search: z.string().optional(),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(20),
})

export const Route = createFileRoute("/_auth/org/classes/$classId")({
  head: () => orgHead("org.class.title"),
  validateSearch: classDetailSearchSchema,
  component: RouteComponent,
})

function SessionCard({
  session,
  index,
  now,
  isNext,
}: {
  session: Session
  index: number
  now: number
  isNext: boolean
}) {
  const { t, i18n } = useTranslation()
  const status = getSessionStatus(session.start_time, now)
  const tileNumber = String(index + 1).padStart(2, "0")
  const startStr = formatSessionDate(session.start_time, i18n.language, "short")
  const relativeStr = formatRelativeTime(session.start_time, now, i18n.language)
  const isEnded = status === "ended"

  // Status drives a single accent so the eye lands on what's actionable:
  // live = destructive, the upcoming "next" = primary, the rest stays neutral.
  const accent = isEnded
    ? "ring-foreground/10 hover:ring-foreground/25"
    : status === "live"
      ? "ring-destructive/40 hover:ring-destructive/60"
      : isNext
        ? "ring-primary/45 hover:ring-primary/65"
        : "ring-foreground/10 hover:ring-foreground/30"
  const rail =
    status === "live" ? "bg-destructive" : isNext ? "bg-primary" : isEnded ? "bg-foreground/15" : "bg-foreground/20"

  return (
    <Link
      to="/org/classes/class-sessions/$classSessionId"
      params={{ classSessionId: session.id! }}
      className={cn(
        "group/tile bg-card text-card-foreground relative isolate flex flex-col gap-2.5 overflow-hidden rounded-xl p-3.5 ps-4 ring-1 transition-all",
        accent,
        isNext && !isEnded && "bg-primary/[0.04]",
        isEnded && "opacity-75 hover:opacity-100"
      )}
    >
      {/* Status-keyed accent rail on the inline-start edge (RTL-safe). */}
      <span aria-hidden className={cn("absolute inset-y-0 start-0 w-1", rail)} />

      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/tile:opacity-100"
      />

      {/* Oversized ordinal as a quiet design anchor — gives the grid a lesson-tile cadence. */}
      <span
        aria-hidden
        className="text-foreground/[0.05] pointer-events-none absolute -top-3 end-2 font-mono text-6xl leading-none font-bold tabular-nums select-none"
      >
        {tileNumber}
      </span>

      <div className="flex min-h-6 items-center gap-2">
        {status === "live" ? (
          <SessionStatusPill status="live" size="sm" />
        ) : isEnded ? (
          <SessionStatusPill status="ended" size="sm" />
        ) : isNext ? (
          <span className="text-primary inline-flex items-center gap-1.5 font-mono text-[0.7rem] font-medium tracking-[0.2em] uppercase">
            <span className="bg-primary size-1.5 rounded-full" />
            {t("org.class.sessions.nextUp")}
          </span>
        ) : relativeStr ? (
          <span className="text-muted-foreground inline-flex items-center gap-1.5 text-[0.7rem] font-medium">
            <CalendarClockIcon className="size-3" />
            {relativeStr}
          </span>
        ) : null}
      </div>

      <div className="flex flex-col gap-1">
        <h3 className="line-clamp-2 text-sm leading-snug font-semibold tracking-tight text-balance">{session.name}</h3>
        {session.description && (
          <p className="text-muted-foreground line-clamp-1 text-xs leading-relaxed">{session.description}</p>
        )}
      </div>

      <div className="mt-auto flex items-center justify-between gap-3 border-t border-dashed pt-2.5">
        <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-[0.7rem]">
          <CalendarClockIcon className="size-3" />
          {startStr}
        </span>
        <span className="text-muted-foreground group-hover/tile:text-foreground inline-flex items-center gap-1 text-[0.7rem] font-medium underline-offset-4 transition-colors group-hover/tile:underline">
          {t("org.class.sessions.open")}
          <span className="transition-transform group-hover/tile:-translate-x-0.5 rtl:rotate-180">→</span>
        </span>
      </div>
    </Link>
  )
}

function SessionCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex flex-col gap-2.5 rounded-xl p-3.5 ring-1">
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-3 w-6" />
      </div>
      <div className="flex flex-col gap-1.5">
        <Skeleton className="h-4 w-4/5" />
        <Skeleton className="h-3 w-3/5" />
      </div>
      <div className="flex items-center justify-between border-t border-dashed pt-2.5">
        <Skeleton className="h-3 w-20" />
        <Skeleton className="h-3 w-10" />
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
        className="pointer-events-none absolute inset-0 -z-10 [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)] [background-size:48px_48px] opacity-[0.04]"
      />
    </>
  )
}

function RouteComponent() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { classId } = Route.useParams()
  const { canView, canEdit: canCreateSession } = useClassPermissions()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const now = useNow(30_000)
  const { can, user: accessUser } = useAccess()

  const [formOpen, setFormOpen] = useState(false)
  const [enrollOpen, setEnrollOpen] = useState(false)
  const [removeTarget, setRemoveTarget] = useState<ClassMember | null>(null)

  const removeMutation = useDeleteClassesIdMembersUserId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.class.removeMember.success"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdMembersQueryKey(classId) })
        setRemoveTarget(null)
      },
      onError: (err) => {
        const status = (err as { status?: number })?.status
        if (status === 403) {
          toast.error(t("org.class.removeMember.errorForbidden"))
        } else {
          toast.error(t("org.class.removeMember.errorGeneric"))
        }
      },
    },
  })

  const handleConfirmRemove = () => {
    if (!removeTarget?.user_id) return
    removeMutation.mutate({ id: classId, userId: removeTarget.user_id })
  }

  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const { canEdit: canEditAttendance } = useAttendancePermissions()

  const {
    viewMode: sessionsView,
    setViewMode: setSessionsView,
    isTable: sessionsIsTable,
    isGrid: sessionsIsGrid,
  } = useViewMode()
  const { data: classData, isPending: classPending } = useGetClassesId(classId, {
    query: { enabled: canView },
  })

  const cls = (classData?.status === 200 && classData.data.data) || undefined

  useBreadcrumb([
    { label: t("org.nav.classes"), to: "/org/classes" },
    { label: cls?.name ?? null, loading: !cls },
  ])

  // Roster gating mirrors backend canManageClass:
  // admin OR classes:update_any OR caller is class owner.
  const canViewRoster = !!cls && (can("classes:update_any") || (!!cls.user_id && cls.user_id === accessUser.id))

  // Deep-linking to a gated tab falls back to Sessions.
  const activeTab = !canViewRoster && search.tab !== "sessions" ? "sessions" : search.tab

  const listParams = {
    search: search.search || undefined,
    order_by: search.order_by || undefined,
    order_dir: search.order_dir || undefined,
    page: search.page ?? 1,
    page_size: search.page_size ?? 20,
  }

  const { data: sessionsData, isPending: sessionsPending } = useGetClassesIdSessions(classId, listParams, {
    query: { enabled: canView && activeTab === "sessions" },
  })
  const sessionsResult = (sessionsData?.status === 200 && sessionsData.data.data) || undefined
  const sessions = sessionsResult?.items ?? []
  const total = sessionsResult?.total ?? sessions.length

  const { data: membersData, isPending: membersPending } = useGetClassesIdMembers(classId, listParams, {
    query: { enabled: canView && canViewRoster && activeTab === "students" },
  })
  const membersResult = (membersData?.status === 200 && membersData.data.data) || undefined
  const members = membersResult?.items ?? []
  const studentsTotal = membersResult?.total ?? members.length

  const liveCount = sessions.filter((s) => getSessionStatus(s.start_time, now) === "live").length

  // The soonest upcoming session gets the "next up" highlight (Classroom/Canvas
  // pattern) so the eye lands on what's actionable, not the whole scheduled list.
  const nextSessionId = sessions
    .filter((s) => getSessionStatus(s.start_time, now) === "scheduled" && s.start_time)
    .sort((a, b) => new Date(a.start_time!).getTime() - new Date(b.start_time!).getTime())[0]?.id

  const sorting = search.order_by ? [{ id: search.order_by, desc: search.order_dir === "desc" }] : []

  const sessionColumns = useSessionColumns(now)
  const sessionsTable = useAdminTable({
    data: sessions,
    columns: sessionColumns,
    rowCount: total,
    sorting,
  })

  const studentColumns = useStudentColumns(setRemoveTarget)
  const studentsTable = useAdminTable({
    data: members,
    columns: studentColumns,
    rowCount: studentsTotal,
    sorting,
  })

  // Sort/search/page tokens are tab-specific; reset them on switch so a stale
  // order_by from one tab never leaks into another tab's query.
  const handleTabChange = (tab: string) => {
    navigate({
      search: { tab: tab as ClassTab, page: 1, page_size: search.page_size },
    })
  }

  const handleMatrixPageChange = (page: number) => {
    navigate({ search: { ...search, page } })
  }

  if (!allowed) return null

  if (classPending) {
    return (
      <div className="flex flex-col gap-6 py-6">
        <Skeleton className="h-28 w-full rounded-2xl" />
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
          <SessionCardSkeleton />
          <SessionCardSkeleton />
          <SessionCardSkeleton />
          <SessionCardSkeleton />
        </div>
      </div>
    )
  }

  const teacherName = cls?.user?.name ?? ""

  return (
    <div className="relative isolate flex flex-col gap-6 pb-10">
      <DecorativeBackground />

      <header className="border-foreground/10 bg-card/50 relative mt-5 flex flex-col gap-5 overflow-hidden rounded-2xl border p-4 backdrop-blur-sm md:flex-row md:items-start md:justify-between md:gap-8 md:p-5">
        <div className="flex min-w-0 flex-col gap-2.5">
          <div className="flex flex-wrap items-center gap-2.5">
            <Eyebrow>{t("org.class.eyebrow")}</Eyebrow>
            {liveCount > 0 && <SessionStatusPill status="live" size="sm" />}
          </div>

          <h1 className="max-w-2xl text-2xl leading-tight font-semibold tracking-tight text-balance md:text-3xl">
            {cls?.name ?? "—"}
          </h1>

          {cls?.description && (
            <p className="text-muted-foreground line-clamp-2 max-w-xl text-sm leading-relaxed">{cls.description}</p>
          )}

          {Boolean(cls?.user_id) && (
            <div className="text-muted-foreground mt-0.5 inline-flex items-center gap-2 text-sm">
              {teacherName ? <UserAvatar name={teacherName} size="sm" /> : <UserIcon className="size-4" />}
              <span className="text-foreground font-medium">{teacherName || t("org.class.unknownTeacher")}</span>
              <span className="text-muted-foreground/40">·</span>
              <Eyebrow className="text-muted-foreground text-[0.65rem] rtl:text-xs">
                {t("org.class.instructor")}
              </Eyebrow>
            </div>
          )}
        </div>

        <Button
          variant="outline"
          size="sm"
          className="shrink-0 max-md:self-start"
          render={<Link to="/org/classes/$classId/gradebook" params={{ classId }} />}
        >
          <TrophyIcon className="size-4" />
          {t("org.class.gradebook.open")}
        </Button>
      </header>

      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList variant="line">
          <TabsTrigger value="sessions">{t("org.class.tabs.sessions")}</TabsTrigger>
          {canViewRoster && <TabsTrigger value="students">{t("org.class.tabs.students")}</TabsTrigger>}
          {canViewRoster && <TabsTrigger value="attendance">{t("org.class.tabs.attendance")}</TabsTrigger>}
        </TabsList>

        <TabsContent value="sessions" className="flex flex-col gap-4">
          <div className="flex flex-wrap items-end justify-between gap-3 border-b border-dashed pb-3">
            <div className="flex items-baseline gap-2.5">
              <h2 className="text-lg font-semibold tracking-tight">{t("org.class.sessions.title")}</h2>
              <span className="text-muted-foreground font-mono text-sm tabular-nums">{total}</span>
            </div>

            <div className="flex items-center gap-2">
              <ViewModeToggle value={sessionsView} onChange={setSessionsView} />
              {canCreateSession && (
                <Button size="sm" onClick={() => setFormOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.sessions.newSession")}
                </Button>
              )}
            </div>
          </div>

          <TableFilter
            table={sessionsTable}
            searchPlaceholder={t("org.class.sessions.searchPlaceholder")}
            sortLabel={t("org.class.toolbar.sort")}
            columnsLabel={t("org.class.toolbar.columns")}
            toggleColumnsLabel={t("org.class.toolbar.toggleColumns")}
            showColumnsToggle={sessionsIsTable}
          />

          {sessionsPending && sessionsIsGrid ? (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              <SessionCardSkeleton />
              <SessionCardSkeleton />
              <SessionCardSkeleton />
              <SessionCardSkeleton />
            </div>
          ) : !sessionsPending && sessions.length === 0 ? (
            <EmptyState
              icon={CalendarClockIcon}
              title={t("org.class.sessions.emptyTitle")}
              description={t("org.class.sessions.emptyHint")}
            >
              {canCreateSession && (
                <Button onClick={() => setFormOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.sessions.newSession")}
                </Button>
              )}
            </EmptyState>
          ) : sessionsIsTable ? (
            <Card className="gap-0 overflow-hidden p-0">
              <div className="overflow-x-auto">
                <DataTable
                  table={sessionsTable}
                  isLoading={sessionsPending}
                  emptyIcon={<CalendarClockIcon className="size-5" />}
                  emptyTitle={t("org.class.sessions.emptyTitle")}
                  emptyHint={t("org.class.sessions.emptyHint")}
                />
              </div>
              <DataTablePagination table={sessionsTable} />
            </Card>
          ) : (
            <>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
                {sessions.map((s, i) => (
                  <SessionCard key={s.id} session={s} index={i} now={now} isNext={!!s.id && s.id === nextSessionId} />
                ))}
              </div>
              <DataTablePagination table={sessionsTable} />
            </>
          )}
        </TabsContent>

        {canViewRoster && (
          <TabsContent value="students" className="flex flex-col gap-4">
            <div className="flex flex-wrap items-end justify-between gap-3 border-b border-dashed pb-3">
              <div className="flex items-baseline gap-2.5">
                <h2 className="text-lg font-semibold tracking-tight">{t("org.class.students.title")}</h2>
                <span className="text-muted-foreground font-mono text-sm tabular-nums">{studentsTotal}</span>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => setEnrollOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.students.addMember")}
                </Button>
              </div>
            </div>

            <TableFilter
              table={studentsTable}
              searchPlaceholder={t("org.class.students.searchPlaceholder")}
              sortLabel={t("org.class.toolbar.sort")}
              columnsLabel={t("org.class.toolbar.columns")}
              toggleColumnsLabel={t("org.class.toolbar.toggleColumns")}
              showColumnsToggle
            />

            {!membersPending && members.length === 0 ? (
              <EmptyState
                icon={UsersIcon}
                title={t("org.class.students.emptyTitle")}
                description={t("org.class.students.emptyHint")}
              >
                <Button onClick={() => setEnrollOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.students.addMember")}
                </Button>
              </EmptyState>
            ) : (
              <Card className="gap-0 overflow-hidden p-0">
                <div className="overflow-x-auto">
                  <DataTable
                    table={studentsTable}
                    isLoading={membersPending}
                    emptyIcon={<UsersIcon className="size-5" />}
                    emptyTitle={t("org.class.students.emptyTitle")}
                    emptyHint={t("org.class.students.emptyHint")}
                  />
                </div>
                <DataTablePagination table={studentsTable} />
              </Card>
            )}
          </TabsContent>
        )}

        {canViewRoster && (
          <TabsContent value="attendance" className="flex flex-col gap-4">
            <div className="flex items-baseline gap-2.5 border-b border-dashed pb-3">
              <h2 className="text-lg font-semibold tracking-tight">{t("org.class.attendance.title")}</h2>
            </div>
            <AttendanceMatrixView
              classId={classId}
              canEdit={canEditAttendance}
              page={search.page ?? 1}
              pageSize={search.page_size ?? 20}
              search={search.search}
              orderBy={search.order_by}
              orderDir={search.order_dir}
              onPageChange={handleMatrixPageChange}
            />
          </TabsContent>
        )}
      </Tabs>

      {canCreateSession && (
        <SessionCreateModal open={formOpen} onOpenChange={setFormOpen} classId={classId} session={null} />
      )}

      {canViewRoster && (
        <>
          <EnrollMemberModal open={enrollOpen} onOpenChange={setEnrollOpen} classId={classId} />
          <DeleteConfirmDialog
            open={!!removeTarget}
            onOpenChange={(open) => !open && setRemoveTarget(null)}
            resourceName={removeTarget?.user?.name ?? t("org.class.students.unknownName")}
            onConfirm={handleConfirmRemove}
            isLoading={removeMutation.isPending}
          />
        </>
      )}
    </div>
  )
}

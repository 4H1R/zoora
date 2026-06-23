import type {
  GithubCom4H1RZooraInternalDomainClassMember as ClassMember,
  GithubCom4H1RZooraInternalDomainClassSession as Session,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { CalendarClockIcon, PlusIcon, TrophyIcon, UserIcon, UserMinusIcon, UsersIcon } from "lucide-react"
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
import { formatSessionDate, getSessionStatus, useNow } from "@/lib/session-status"

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
  page_size: z.number().int().positive().optional().default(8),
})

export const Route = createFileRoute("/_auth/org/classes/$classId")({
  head: () => orgHead("org.class.title"),
  validateSearch: classDetailSearchSchema,
  component: RouteComponent,
})

function SessionCard({ session, index, now }: { session: Session; index: number; now: number }) {
  const { t, i18n } = useTranslation()
  const status = getSessionStatus(session.start_time, now)
  const tileNumber = String(index + 1).padStart(2, "0")
  const startStr = formatSessionDate(session.start_time, i18n.language, "short")

  return (
    <Link
      to="/org/classes/classsessions/$classSessionId"
      params={{ classSessionId: session.id! }}
      className="group/tile bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-2.5 overflow-hidden rounded-xl p-3.5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-md"
    >
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/tile:opacity-100"
      />

      <div className="flex items-start justify-between gap-2">
        <SessionStatusPill status={status} size="sm" />
        <span className="text-muted-foreground font-mono text-[0.7rem] tracking-[0.2em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-1">
        <h3 className="line-clamp-2 text-sm leading-snug font-semibold tracking-tight text-balance">{session.name}</h3>
        {session.description ? (
          <p className="text-muted-foreground line-clamp-1 text-xs leading-relaxed">{session.description}</p>
        ) : null}
      </div>

      <div className="mt-auto flex items-center justify-between gap-3 border-t border-dashed pt-2.5">
        <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-[0.7rem]">
          <CalendarClockIcon className="size-3" />
          {startStr}
        </span>
        <span className="text-muted-foreground group-hover/tile:text-foreground text-[0.7rem] font-medium underline-offset-4 transition-colors group-hover/tile:underline">
          {t("org.class.sessions.open")} →
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

function StudentCard({
  member,
  index,
  onRemove,
}: {
  member: ClassMember
  index: number
  onRemove?: (member: ClassMember) => void
}) {
  const { t, i18n } = useTranslation()
  const name = member.user?.name ?? t("org.class.students.unknownName")
  const username = member.user?.username ?? ""
  const tileNumber = String(index + 1).padStart(2, "0")
  const joinedStr = member.created_at ? formatSessionDate(member.created_at, i18n.language, "short") : ""

  return (
    <div className="group/student bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex items-center gap-3 overflow-hidden rounded-xl p-2.5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-md">
      <UserAvatar name={name} size="md" />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm font-semibold tracking-tight">{name}</span>
        <span className="text-muted-foreground inline-flex items-center gap-1.5 truncate text-xs">
          {username ? <span className="font-mono">@{username}</span> : null}
          {username && joinedStr ? <span className="text-muted-foreground/40">·</span> : null}
          {joinedStr ? (
            <span className="inline-flex items-center gap-1">
              <CalendarClockIcon className="size-2.5" />
              {joinedStr}
            </span>
          ) : null}
        </span>
      </div>
      {onRemove ? (
        <button
          type="button"
          onClick={() => onRemove(member)}
          aria-label={t("org.class.students.removeAction")}
          title={t("org.class.students.removeAction")}
          className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive focus-visible:ring-ring inline-flex size-7 shrink-0 items-center justify-center rounded-full opacity-0 transition-all group-hover/student:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:outline-none"
        >
          <UserMinusIcon className="size-3.5" />
        </button>
      ) : (
        <span className="text-muted-foreground font-mono text-[0.7rem] tracking-[0.2em]">/{tileNumber}</span>
      )}
    </div>
  )
}

function StudentCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex items-center gap-3 rounded-xl p-2.5 ring-1">
      <Skeleton className="size-7 rounded-full" />
      <div className="flex flex-1 flex-col gap-1.5">
        <Skeleton className="h-3.5 w-32" />
        <Skeleton className="h-3 w-20" />
      </div>
    </div>
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
  const {
    viewMode: studentsView,
    setViewMode: setStudentsView,
    isTable: studentsIsTable,
    isGrid: studentsIsGrid,
  } = useViewMode()

  const { data: classData, isPending: classPending } = useGetClassesId(classId, {
    query: { enabled: canView },
  })

  const cls = (classData?.status === 200 && classData.data.data) || undefined

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
    page_size: search.page_size ?? 8,
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
            {liveCount > 0 ? <SessionStatusPill status="live" size="sm" /> : null}
          </div>

          <h1 className="max-w-2xl text-2xl leading-tight font-semibold tracking-tight text-balance md:text-3xl">
            {cls?.name ?? "—"}
          </h1>

          {cls?.description ? (
            <p className="text-muted-foreground line-clamp-2 max-w-xl text-sm leading-relaxed">{cls.description}</p>
          ) : null}

          {cls?.user_id ? (
            <div className="text-muted-foreground mt-0.5 inline-flex items-center gap-2 text-sm">
              {teacherName ? <UserAvatar name={teacherName} size="sm" /> : <UserIcon className="size-4" />}
              <span className="text-foreground font-medium">{teacherName || t("org.class.unknownTeacher")}</span>
              <span className="text-muted-foreground/40">·</span>
              <Eyebrow className="text-muted-foreground text-[0.65rem] rtl:text-xs">
                {t("org.class.instructor")}
              </Eyebrow>
            </div>
          ) : null}
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
          {canViewRoster ? <TabsTrigger value="students">{t("org.class.tabs.students")}</TabsTrigger> : null}
          {canViewRoster ? <TabsTrigger value="attendance">{t("org.class.tabs.attendance")}</TabsTrigger> : null}
        </TabsList>

        <TabsContent value="sessions" className="flex flex-col gap-4">
          <div className="flex flex-wrap items-end justify-between gap-3 border-b border-dashed pb-3">
            <div className="flex items-baseline gap-2.5">
              <h2 className="text-lg font-semibold tracking-tight">{t("org.class.sessions.title")}</h2>
              <span className="text-muted-foreground font-mono text-sm tabular-nums">{total}</span>
            </div>

            <div className="flex items-center gap-2">
              <ViewModeToggle value={sessionsView} onChange={setSessionsView} />
              {canCreateSession ? (
                <Button size="sm" onClick={() => setFormOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.sessions.newSession")}
                </Button>
              ) : null}
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
              {canCreateSession ? (
                <Button onClick={() => setFormOpen(true)}>
                  <PlusIcon className="size-4" />
                  {t("org.class.sessions.newSession")}
                </Button>
              ) : null}
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
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              {sessions.map((s, i) => (
                <SessionCard key={s.id} session={s} index={i} now={now} />
              ))}
            </div>
          )}
        </TabsContent>

        {canViewRoster ? (
          <TabsContent value="students" className="flex flex-col gap-4">
            <div className="flex flex-wrap items-end justify-between gap-3 border-b border-dashed pb-3">
              <div className="flex items-baseline gap-2.5">
                <h2 className="text-lg font-semibold tracking-tight">{t("org.class.students.title")}</h2>
                <span className="text-muted-foreground font-mono text-sm tabular-nums">{studentsTotal}</span>
              </div>
              <div className="flex items-center gap-2">
                <ViewModeToggle value={studentsView} onChange={setStudentsView} />
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
              showColumnsToggle={studentsIsTable}
            />

            {membersPending && studentsIsGrid ? (
              <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
                <StudentCardSkeleton />
                <StudentCardSkeleton />
                <StudentCardSkeleton />
                <StudentCardSkeleton />
              </div>
            ) : !membersPending && members.length === 0 ? (
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
            ) : studentsIsTable ? (
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
            ) : (
              <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
                {members.map((m, i) => (
                  <StudentCard key={m.id} member={m} index={i} onRemove={setRemoveTarget} />
                ))}
              </div>
            )}
          </TabsContent>
        ) : null}

        {canViewRoster ? (
          <TabsContent value="attendance" className="flex flex-col gap-4">
            <div className="flex items-baseline gap-2.5 border-b border-dashed pb-3">
              <h2 className="text-lg font-semibold tracking-tight">{t("org.class.attendance.title")}</h2>
            </div>
            <AttendanceMatrixView
              classId={classId}
              canEdit={canEditAttendance}
              page={search.page ?? 1}
              pageSize={search.page_size ?? 8}
              search={search.search}
              orderBy={search.order_by}
              orderDir={search.order_dir}
              onPageChange={handleMatrixPageChange}
            />
          </TabsContent>
        ) : null}
      </Tabs>

      {canCreateSession ? (
        <SessionCreateModal open={formOpen} onOpenChange={setFormOpen} classId={classId} session={null} />
      ) : null}

      {canViewRoster ? (
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
      ) : null}
    </div>
  )
}

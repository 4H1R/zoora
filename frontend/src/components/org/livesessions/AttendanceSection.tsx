import type {
  GithubCom4H1RZooraInternalDomainAttendance as Attendance,
  GetClassesIdSessionsSessionIdAttendanceStatus as AttendanceStatus,
} from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"

import { useQueryClient } from "@tanstack/react-query"
import {
  CheckCircle2Icon,
  ClockIcon,
  PencilIcon,
  ShieldCheckIcon,
  SparklesIcon,
  Trash2Icon,
  UserCheckIcon,
  XCircleIcon,
} from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsSessionIdAttendanceQueryKey,
  useDeleteAttendanceAttendanceId,
  useGetClassesIdSessionsSessionIdAttendance,
  usePostClassesIdSessionsSessionIdAttendanceAutoMark,
} from "@/api/attendance/attendance"
import { useGetClassesId, useGetClassesIdMembers } from "@/api/classes/classes"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { SectionNoResults } from "@/components/org/session/section-no-results"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { SectionToolbar } from "@/components/org/session/section-toolbar"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { DEFAULT_PAGE_SIZE } from "@/lib/list"
import { useSectionList } from "@/lib/use-section-list"
import { cn } from "@/lib/utils"

import { AttendanceEditDialog } from "./AttendanceEditDialog"
import { AttendanceRoster } from "./AttendanceRoster"
import { useAttendancePermissions } from "./use-attendance-permissions"

const STATUS_META: Record<string, { style: string; icon: typeof CheckCircle2Icon }> = {
  present: { style: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400", icon: CheckCircle2Icon },
  absent: { style: "bg-destructive/10 text-destructive", icon: XCircleIcon },
  late: { style: "bg-amber-500/10 text-amber-600 dark:text-amber-400", icon: ClockIcon },
  excused: { style: "bg-primary/10 text-primary", icon: ShieldCheckIcon },
}

interface AutoMarkControlProps {
  classId: string
  classSessionId: string
}

function AutoMarkControl({ classId, classSessionId }: AutoMarkControlProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const autoMark = usePostClassesIdSessionsSessionIdAttendanceAutoMark({
    mutation: {
      onSuccess: (result) => {
        const data = (result.status === 200 && result.data.data) || undefined
        toast.success(
          t("org.session.attendance.autoMark.success", {
            marked: data?.marked ?? 0,
            skipped: data?.skipped ?? 0,
          })
        )
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
      },
    },
  })

  const run = () => {
    autoMark.mutate({
      id: classId,
      sessionId: classSessionId,
      data: { source: "live_room" },
    })
  }

  return (
    <div className="bg-card ring-foreground/10 relative isolate flex flex-col gap-4 overflow-hidden rounded-2xl p-5 ring-1">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)]"
      />
      <div className="flex flex-col gap-1.5">
        <Eyebrow className="inline-flex items-center gap-2">
          <SparklesIcon className="size-3.5" />
          {t("org.session.attendance.autoMark.eyebrow")}
        </Eyebrow>
        <p className="text-muted-foreground text-sm leading-relaxed">{t("org.session.attendance.autoMark.hint")}</p>
      </div>

      <div>
        <Button disabled={autoMark.isPending} onClick={run}>
          <UserCheckIcon className="size-4" />
          {t("org.session.attendance.autoMark.run")}
        </Button>
      </div>
    </div>
  )
}

interface AttendanceRowProps {
  attendance: Attendance
  canEdit: boolean
  canDelete: boolean
  onEdit: (a: Attendance) => void
  onDelete: (a: Attendance) => void
}

function AttendanceRow({ attendance, canEdit, canDelete, onEdit, onDelete }: AttendanceRowProps) {
  const { t } = useTranslation()
  const name = attendance.user?.name ?? "—"
  const status = attendance.status ?? "absent"
  const meta = STATUS_META[status] ?? STATUS_META.absent
  const StatusIcon = meta.icon

  return (
    <div className="group/row bg-card ring-foreground/10 hover:ring-foreground/25 flex items-center gap-3 rounded-2xl px-4 py-3 ring-1 transition-all">
      <Avatar className="size-9 shrink-0">
        <AvatarFallback className={cn("text-xs font-semibold text-white", getEntityColor(name))}>
          {getInitials(name)}
        </AvatarFallback>
      </Avatar>

      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm font-medium">{name}</span>
        {Boolean(attendance.remarks) && (
          <span className="text-muted-foreground truncate text-xs">{attendance.remarks}</span>
        )}
      </div>

      {attendance.is_auto_marked ? (
        <span className="text-muted-foreground hidden items-center gap-1 font-mono text-[10px] tracking-[0.2em] uppercase sm:inline-flex">
          <SparklesIcon className="size-3" />
          {t("org.session.attendance.auto")}
        </span>
      ) : (
        <span className="text-muted-foreground hidden items-center gap-1 font-mono text-[10px] tracking-[0.2em] uppercase sm:inline-flex">
          {t("org.session.attendance.manual")}
        </span>
      )}

      <span
        className={cn(
          "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[10px] tracking-[0.2em] uppercase",
          meta.style
        )}
      >
        <StatusIcon className="size-3" />
        {t(`common.statuses.attendance.${status}`)}
      </span>

      <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/row:opacity-100">
        {canEdit && (
          <Button variant="ghost" size="icon-xs" onClick={() => onEdit(attendance)}>
            <PencilIcon />
          </Button>
        )}
        {canDelete && (
          <Button
            variant="ghost"
            size="icon-xs"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => onDelete(attendance)}
          >
            <Trash2Icon />
          </Button>
        )}
      </div>
    </div>
  )
}

interface AttendanceSectionProps {
  classId: string
  classSessionId: string
}

export function AttendanceSection({ classId, classSessionId }: AttendanceSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { can, user: accessUser } = useAccess()
  const { canView, canCreate, canEdit, canDelete } = useAttendancePermissions()
  const [editing, setEditing] = useState<Attendance | null>(null)
  const [deleting, setDeleting] = useState<Attendance | null>(null)

  const { data: classData, isPending: classPending } = useGetClassesId(classId, {
    query: { enabled: canView && !!classId },
  })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  // Roster visibility mirrors backend canManageClass: classes:update_any OR class owner.
  const canViewRoster = !!cls && (can("classes:update_any") || (!!cls.user_id && cls.user_id === accessUser.id))
  const canMark = canCreate || canEdit

  const membersQuery = useGetClassesIdMembers(classId, undefined, {
    query: { enabled: canView && canViewRoster },
  })
  const members = (membersQuery.data?.status === 200 && membersQuery.data.data.data?.items) || []

  const list = useSectionList({ defaultSort: { id: "status", desc: false } })
  const sortOptions: SortOption[] = [
    { id: "status", label: t("org.session.controls.sortFields.status") },
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "updated_at", label: t("org.session.controls.sortFields.updated_at") },
  ]

  // Roster view shows the full, unfiltered roll; the plain list view honours the
  // search / status / sort / page controls.
  const attendanceParams = canViewRoster
    ? { order_by: "status", order_dir: "asc" }
    : {
        search: list.params.search,
        status: (list.status as AttendanceStatus | undefined) ?? undefined,
        order_by: list.params.order_by ?? "status",
        order_dir: list.params.order_dir ?? "asc",
        page: list.params.page,
      }

  const query = useGetClassesIdSessionsSessionIdAttendance(classId, classSessionId, attendanceParams, {
    query: { enabled: canView && !!classId },
  })
  const attendanceData = (query.data?.status === 200 && query.data.data.data) || undefined
  const records = attendanceData?.items ?? []
  const total = attendanceData?.total ?? 0
  const pageSize = attendanceData?.page_size ?? DEFAULT_PAGE_SIZE

  const deleteMutation = useDeleteAttendanceAttendanceId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.attendance.form.deleteSuccess"))
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
        setDeleting(null)
      },
    },
  })

  if (!canView) return null

  const loading = query.isPending || classPending || (canViewRoster && membersQuery.isPending)

  const statusItems = [
    { value: "all", label: t("org.session.controls.status.all") },
    { value: "present", label: t("common.statuses.attendance.present") },
    { value: "absent", label: t("common.statuses.attendance.absent") },
    { value: "late", label: t("common.statuses.attendance.late") },
    { value: "excused", label: t("common.statuses.attendance.excused") },
  ]

  return (
    <section id="attendance" className="flex scroll-mt-20 flex-col gap-5">
      <div className="flex flex-col gap-1.5">
        <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.attendance.title")}</h2>
      </div>

      {canCreate && <AutoMarkControl classId={classId} classSessionId={classSessionId} />}

      {!canViewRoster && !loading && (records.length > 0 || list.isFiltered) && (
        <SectionToolbar
          searchValue={list.searchInput}
          onSearchChange={list.setSearchInput}
          sortOptions={sortOptions}
          sort={list.sort}
          onSortChange={list.setSort}
        >
          <Select
            items={statusItems}
            value={list.status ?? "all"}
            onValueChange={(v) => list.setStatus(v && v !== "all" ? v : undefined)}
          >
            <SelectTrigger className="h-8 w-auto gap-1.5 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusItems.map((item) => (
                <SelectItem key={item.value} value={item.value} className="text-xs">
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </SectionToolbar>
      )}

      {loading ? (
        <div className="flex flex-col gap-2">
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
        </div>
      ) : canViewRoster ? (
        <AttendanceRoster
          classId={classId}
          classSessionId={classSessionId}
          members={members}
          records={records}
          canMark={canMark}
        />
      ) : records.length === 0 ? (
        list.isFiltered ? (
          <SectionNoResults />
        ) : (
          <EmptyState
            icon={UserCheckIcon}
            title={t("org.session.attendance.emptyTitle")}
            description={
              canCreate ? t("org.session.attendance.emptyHint") : t("org.session.attendance.emptyHintMember")
            }
          />
        )
      ) : (
        <>
          <div className="flex flex-col gap-2">
            {records.map((a) => (
              <AttendanceRow
                key={a.id}
                attendance={a}
                canEdit={canEdit}
                canDelete={canDelete}
                onEdit={setEditing}
                onDelete={setDeleting}
              />
            ))}
          </div>
          <SectionPagination page={list.page} pageSize={pageSize} total={total} onPageChange={list.setPage} />
        </>
      )}

      <AttendanceEditDialog
        attendance={editing}
        open={!!editing}
        onOpenChange={(open) => {
          if (!open) setEditing(null)
        }}
        classId={classId}
        classSessionId={classSessionId}
      />

      <DeleteConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          if (!open) setDeleting(null)
        }}
        resourceName={deleting?.user?.name ?? ""}
        onConfirm={() => {
          if (deleting?.id) deleteMutation.mutate({ attendanceId: deleting.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}

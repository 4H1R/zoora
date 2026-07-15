import type { GithubCom4H1RZooraInternalDomainAttendanceMatrixResult as AttendanceMatrix } from "@/api/model"
import type { AttendanceStatus } from "@/components/org/classes/AttendanceCellPopover"

import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClassesIdAttendanceMatrix } from "@/api/attendance/attendance"
import { AttendanceCellPopover } from "@/components/org/classes/AttendanceCellPopover"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

const STATUS_GLYPH: Record<AttendanceStatus, string> = {
  present: "P",
  absent: "A",
  late: "L",
  excused: "E",
}

const STATUS_CLASS: Record<AttendanceStatus, string> = {
  present: "bg-emerald-500/15 text-emerald-600",
  absent: "bg-rose-500/15 text-rose-600",
  late: "bg-amber-500/15 text-amber-600",
  excused: "bg-sky-500/15 text-sky-600",
}

const LEGEND: AttendanceStatus[] = ["present", "absent", "late", "excused"]

function StatusBadge({ status, className, title }: { status: AttendanceStatus; className?: string; title?: string }) {
  return (
    <span
      className={cn("inline-flex items-center justify-center rounded font-medium", STATUS_CLASS[status], className)}
      title={title}
    >
      {STATUS_GLYPH[status]}
    </span>
  )
}

interface AttendanceMatrixViewProps {
  classId: string
  canEdit: boolean
  page: number
  pageSize: number
  search?: string
  orderBy?: string
  orderDir?: "asc" | "desc"
  onPageChange: (page: number) => void
}

export function AttendanceMatrixView({
  classId,
  canEdit,
  page,
  pageSize,
  search,
  orderBy,
  orderDir,
  onPageChange,
}: AttendanceMatrixViewProps) {
  const { t } = useTranslation()
  const now = Date.now()

  const { data, isLoading } = useGetClassesIdAttendanceMatrix(classId, {
    page,
    page_size: pageSize,
    search: search || undefined,
    order_by: orderBy || undefined,
    order_dir: orderDir || undefined,
  })

  const matrix = data?.data?.data as AttendanceMatrix | undefined
  const students = matrix?.students ?? []
  const total = matrix?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  // Derive the "future" flag once per column instead of re-parsing each
  // session's start_time inside every student row.
  const sessions = (matrix?.sessions ?? []).map((sess) => ({
    ...sess,
    future: sess.start_time ? new Date(sess.start_time).getTime() > now : false,
  }))

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3 text-sm">
        {LEGEND.map((s) => (
          <span key={s} className="flex items-center gap-1.5">
            <StatusBadge status={s} className="size-5 text-xs" />
            {t(`org.class.attendance.legend.${s}`)}
          </span>
        ))}
        <span className="text-muted-foreground flex items-center gap-1.5">
          <span className="inline-flex size-5 items-center justify-center text-xs">—</span>
          {t("org.class.attendance.legend.none")}
        </span>
      </div>

      <div className="overflow-x-auto rounded-lg border">
        {isLoading ? (
          <div className="space-y-2 p-4">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="p-8 text-center">
            <p className="font-medium">{t("org.class.attendance.emptyTitle")}</p>
            <p className="text-muted-foreground text-sm">{t("org.class.attendance.emptyHint")}</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="bg-card z-10 min-w-44 md:sticky md:start-0 md:min-w-56">
                  {t("org.class.attendance.student")}
                </TableHead>
                {sessions.map((sess) => (
                  <TableHead
                    key={sess.id}
                    className={cn("min-w-24 text-center", sess.future && "text-muted-foreground")}
                  >
                    <span className="block truncate" title={sess.name}>
                      {sess.name}
                    </span>
                  </TableHead>
                ))}
                <TableHead className="bg-card z-10 min-w-36 text-center md:sticky md:end-0">
                  {t("org.class.attendance.summary")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {students.map((stu) => (
                <TableRow key={stu.user_id}>
                  <TableCell className="bg-card z-10 md:sticky md:start-0">
                    <div className="flex items-center gap-2">
                      <UserAvatar name={stu.user?.name ?? "?"} size="sm" />
                      <span className="truncate">{stu.user?.name ?? stu.user_id}</span>
                    </div>
                  </TableCell>
                  {sessions.map((sess) => {
                    const cell = sess.id ? stu.cells?.[sess.id] : undefined
                    const status = cell?.status as AttendanceStatus | undefined
                    const glyph = status ? (
                      <StatusBadge
                        status={status}
                        className={cn("size-7 text-xs", cell?.is_auto_marked && "ring-dashed ring-1 ring-current/40")}
                        title={`${t(`common.statuses.attendance.${status}`)}${
                          cell?.is_auto_marked ? ` · ${t("org.class.attendance.autoMarked")}` : ""
                        }`}
                      />
                    ) : (
                      <span
                        className="text-muted-foreground/50 inline-flex size-7 items-center justify-center"
                        title={sess.future ? t("org.class.attendance.future") : t("org.class.attendance.noRecord")}
                      >
                        —
                      </span>
                    )

                    return (
                      <TableCell key={sess.id} className={cn("p-1 text-center", sess.future && "bg-muted/30")}>
                        {sess.future || !sess.id || !stu.user_id ? (
                          glyph
                        ) : (
                          <AttendanceCellPopover
                            classId={classId}
                            sessionId={sess.id}
                            studentId={stu.user_id}
                            attendanceId={cell?.id}
                            status={status}
                            disabled={!canEdit}
                          >
                            {glyph}
                          </AttendanceCellPopover>
                        )}
                      </TableCell>
                    )
                  })}
                  <TableCell className="bg-card z-10 text-center text-sm md:sticky md:end-0">
                    <span className="font-medium">{Math.round((stu.summary?.rate ?? 0) * 100)}%</span>
                    <span className="text-muted-foreground ms-2">
                      {stu.summary?.present ?? 0}/{stu.summary?.absent ?? 0}/{stu.summary?.late ?? 0}/
                      {stu.summary?.excused ?? 0}
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-end gap-2">
          <Button
            variant="outline"
            size="icon"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
            aria-label={t("common.pagination.previous")}
          >
            <ChevronLeftIcon className="size-4 rtl:rotate-180" />
          </Button>
          <span className="text-muted-foreground text-sm">
            {page} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="icon"
            disabled={page >= totalPages}
            onClick={() => onPageChange(page + 1)}
            aria-label={t("common.pagination.next")}
          >
            <ChevronRightIcon className="size-4 rtl:rotate-180" />
          </Button>
        </div>
      )}
    </div>
  )
}

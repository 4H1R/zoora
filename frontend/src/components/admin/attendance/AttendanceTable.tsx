import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { CalendarClockIcon, ClipboardCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { AttendanceActions } from "./AttendanceActions"

type AttendanceStatus = "present" | "absent" | "late" | "excused"

const STATUS_BADGE_VARIANT: Record<AttendanceStatus, "default" | "secondary" | "outline" | "destructive"> = {
  present: "default",
  absent: "destructive",
  late: "outline",
  excused: "secondary",
}

function useAttendanceColumns(onEdit: (a: Attendance) => void): ColumnDef<Attendance>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "user",
      header: t("admin.attendance.student"),
      cell: ({ row }) => {
        const name = row.original.user?.name ?? "—"
        return (
          <div className="flex items-center gap-3">
            <div
              className={cn(
                "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
                getEntityColor(name)
              )}
            >
              {getInitials(name)}
            </div>
            <div className="min-w-0">
              <div className="truncate text-sm font-medium">{name}</div>
              {row.original.user?.username && (
                <div className="text-muted-foreground truncate text-xs">{row.original.user.username}</div>
              )}
            </div>
          </div>
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "class",
      header: t("admin.attendance.class"),
      cell: ({ row }) => (
        <span className="text-sm">{row.original.class?.name ?? <span className="text-muted-foreground">—</span>}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "class_session",
      header: t("admin.attendance.session"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.class_session?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "status",
      header: t("admin.attendance.status"),
      cell: ({ row }) => {
        const status = (row.original.status ?? "absent") as AttendanceStatus
        return (
          <Badge variant={STATUS_BADGE_VARIANT[status] ?? "outline"}>
            {t(`common.statuses.attendance.${status}`, { defaultValue: status })}
          </Badge>
        )
      },
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "is_auto_marked",
      header: t("admin.attendance.source"),
      cell: ({ row }) => (
        <Badge variant="outline">
          {row.original.is_auto_marked ? t("admin.attendance.sources.auto") : t("admin.attendance.sources.manual")}
        </Badge>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "remarks",
      header: t("admin.attendance.remarks"),
      cell: ({ row }) => <span className="text-muted-foreground truncate text-xs">{row.original.remarks || "—"}</span>,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.attendance.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
          <CalendarClockIcon className="size-3.5" />
          {formatDate(row.original.created_at)}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <AttendanceActions attendance={row.original} onEdit={onEdit} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface AttendanceTableProps {
  items: Attendance[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (a: Attendance) => void
}

export function AttendanceTable({ items, total, isLoading, sorting, onEdit }: AttendanceTableProps) {
  const { t } = useTranslation()
  const columns = useAttendanceColumns(onEdit)

  const table = useAdminTable({
    data: items,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.attendance.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ClipboardCheckIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.attendance.noResults")}
            emptyHint={t("admin.attendance.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

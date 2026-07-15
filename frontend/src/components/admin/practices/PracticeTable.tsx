import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { CalendarClockIcon, DumbbellIcon, TargetIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { PracticeActions } from "./PracticeActions"

function usePracticeColumns({ onEdit }: { onEdit: (p: PracticeRoom) => void }): ColumnDef<PracticeRoom>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "title",
      header: t("admin.practices.titleColumn"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.title ?? "")
            )}
          >
            {getInitials(row.original.title ?? "")}
          </div>
          <div className="min-w-0">
            <div className="truncate text-start text-sm font-medium">{row.original.title}</div>
            {row.original.content && (
              <div className="text-muted-foreground truncate text-start text-xs">{row.original.content}</div>
            )}
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "class",
      header: t("admin.practices.class"),
      cell: ({ row }) => (
        <span className="text-sm">{row.original.class?.name ?? <span className="text-muted-foreground">—</span>}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "class_session",
      header: t("admin.practices.session"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.class_session?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "user",
      header: t("admin.practices.teacher"),
      cell: ({ row }) => (
        <span className="text-sm">{row.original.user?.name ?? <span className="text-muted-foreground">—</span>}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "max_score",
      header: t("admin.practices.maxScore"),
      cell: ({ row }) => (
        <span className="inline-flex items-center gap-1.5 text-xs tabular-nums">
          <TargetIcon className="text-muted-foreground size-3.5" />
          {row.original.max_score ?? 0}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "start_time",
      header: t("admin.practices.startTime"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
          <CalendarClockIcon className="size-3.5" />
          {formatDate(row.original.start_time)}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "end_time",
      header: t("admin.practices.endTime"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.end_time)}</span>,
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.practices.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <PracticeActions practice={row.original} onEdit={onEdit} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface PracticeTableProps {
  practices: PracticeRoom[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (p: PracticeRoom) => void
}

export function PracticeTable({ practices, total, isLoading, sorting, onEdit }: PracticeTableProps) {
  const { t } = useTranslation()
  const columns = usePracticeColumns({ onEdit })

  const table = useAdminTable({
    data: practices,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.practices.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<DumbbellIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.practices.noResults")}
            emptyHint={t("admin.practices.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

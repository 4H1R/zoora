import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { CalendarClockIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { useAdminTable } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { SessionActions } from "./SessionActions"

function useSessionColumns({
  classId,
  onEdit,
}: {
  classId: string
  onEdit: (session: Session) => void
}): ColumnDef<Session>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "name",
      header: t("admin.sessions.name"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.name ?? "")
            )}
          >
            {getInitials(row.original.name ?? "")}
          </div>
          <div className="min-w-0">
            <div className="truncate text-start text-sm font-medium">{row.original.name}</div>
            {row.original.description && (
              <div className="text-muted-foreground truncate text-start text-xs">{row.original.description}</div>
            )}
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "start_time",
      header: t("admin.sessions.startTime"),
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
      accessorKey: "created_at",
      header: t("admin.sessions.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <SessionActions session={row.original} classId={classId} onEdit={onEdit} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface SessionTableProps {
  classId: string
  sessions: Session[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (session: Session) => void
}

export function SessionTable({ classId, sessions, total, isLoading, sorting, onEdit }: SessionTableProps) {
  const { t } = useTranslation()
  const columns = useSessionColumns({ classId, onEdit })

  const table = useAdminTable({
    data: sessions,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.sessions.searchPlaceholder")}
        sortLabel={t("admin.sessions.toolbar.sort")}
        columnsLabel={t("admin.sessions.toolbar.columns")}
        toggleColumnsLabel={t("admin.sessions.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<CalendarClockIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.sessions.noResults")}
            emptyHint={t("admin.sessions.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

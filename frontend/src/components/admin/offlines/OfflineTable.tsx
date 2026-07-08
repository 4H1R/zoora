import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { EyeIcon, FileVideoIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { OfflineActions } from "./OfflineActions"

function useOfflineColumns({
  onEdit,
}: {
  onEdit: (room: OfflineRoom) => void
}): ColumnDef<OfflineRoom>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "title",
      header: t("admin.offlines.titleColumn"),
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
            {row.original.description && (
              <div className="text-muted-foreground truncate text-start text-xs">
                {row.original.description}
              </div>
            )}
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "class",
      header: t("admin.offlines.class"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.class?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "creator",
      header: t("admin.offlines.creator"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.creator?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "view_count",
      header: t("admin.offlines.viewCount"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs tabular-nums">
          <EyeIcon className="size-3.5" />
          {row.original.view_count ?? 0}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "published_at",
      header: t("admin.offlines.publishedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.published_at)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.offlines.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <OfflineActions room={row.original} onEdit={onEdit} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface OfflineTableProps {
  rooms: OfflineRoom[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (room: OfflineRoom) => void
}

export function OfflineTable({ rooms, total, isLoading, sorting, onEdit }: OfflineTableProps) {
  const { t } = useTranslation()
  const columns = useOfflineColumns({ onEdit })

  const table = useAdminTable({
    data: rooms,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.offlines.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<FileVideoIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.offlines.noResults")}
            emptyHint={t("admin.offlines.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

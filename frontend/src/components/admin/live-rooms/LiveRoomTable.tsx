import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { CalendarClockIcon, VideoIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { LiveRoomActions } from "./LiveRoomActions"

type LiveRoomStatus = "created" | "active" | "finished"

const STATUS_BADGE_VARIANT: Record<LiveRoomStatus, "default" | "secondary" | "outline"> = {
  created: "outline",
  active: "default",
  finished: "secondary",
}

function useLiveRoomStatusLabel() {
  const { t } = useTranslation()
  return (status?: string) => {
    if (!status) return "—"
    return t(`admin.liveRooms.statuses.${status}`, { defaultValue: status })
  }
}

function useLiveRoomColumns({
  onEnded,
  onDeleted,
}: {
  onEnded: () => void
  onDeleted: () => void
}): ColumnDef<LiveRoom>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const statusLabel = useLiveRoomStatusLabel()

  return [
    {
      accessorKey: "livekit_room_name",
      header: t("admin.liveRooms.name"),
      cell: ({ row }) => {
        const sessionName = row.original.class_session?.name
        const className = row.original.class_session?.class?.name
        const label = sessionName ?? row.original.livekit_room_name ?? ""
        return (
          <div className="flex items-center gap-3">
            <div
              className={cn(
                "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
                getEntityColor(label)
              )}
            >
              {getInitials(label)}
            </div>
            <div className="min-w-0">
              <div className="truncate text-start text-sm font-medium">{label}</div>
              {className && (
                <div className="text-muted-foreground truncate text-start text-xs">{className}</div>
              )}
            </div>
          </div>
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "status",
      header: t("admin.liveRooms.status"),
      cell: ({ row }) => {
        const status = (row.original.status ?? "created") as LiveRoomStatus
        return <Badge variant={STATUS_BADGE_VARIANT[status] ?? "outline"}>{statusLabel(status)}</Badge>
      },
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "actual_start_time",
      header: t("admin.liveRooms.startedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
          <CalendarClockIcon className="size-3.5" />
          {row.original.actual_start_time ? formatDate(row.original.actual_start_time) : "—"}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "actual_end_time",
      header: t("admin.liveRooms.endedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">
          {row.original.actual_end_time ? formatDate(row.original.actual_end_time) : "—"}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.liveRooms.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <LiveRoomActions room={row.original} onEnded={onEnded} onDeleted={onDeleted} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface LiveRoomTableProps {
  rooms: LiveRoom[]
  total: number
  isLoading: boolean
  sorting: SortingState
}

export function LiveRoomTable({ rooms, total, isLoading, sorting }: LiveRoomTableProps) {
  const { t } = useTranslation()

  // After end/delete, callers want feedback; LiveRoomActions invalidates the
  // query itself, but we also accept callbacks for parent-level extensions.
  const noop = () => {}
  const columns = useLiveRoomColumns({ onEnded: noop, onDeleted: noop })

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
        searchPlaceholder={t("admin.liveRooms.searchPlaceholder")}
        sortLabel={t("admin.liveRooms.toolbar.sort")}
        columnsLabel={t("admin.liveRooms.toolbar.columns")}
        toggleColumnsLabel={t("admin.liveRooms.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<VideoIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.liveRooms.noResults")}
            emptyHint={t("admin.liveRooms.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

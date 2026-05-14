import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { CalendarClockIcon, EllipsisVerticalIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

type AttendanceStatus = "present" | "absent" | "late" | "excused"

const STATUS_BADGE_VARIANT: Record<AttendanceStatus, "default" | "secondary" | "outline" | "destructive"> = {
  present: "default",
  absent: "destructive",
  late: "outline",
  excused: "secondary",
}

function useAttendanceStatusLabel() {
  const { t } = useTranslation()
  return (status?: string) => {
    if (!status) return "—"
    return t(`admin.attendance.statuses.${status}`, { defaultValue: status })
  }
}

interface AttendanceRowActionsProps {
  attendance: Attendance
  onEdit: (a: Attendance) => void
  onDelete: (a: Attendance) => void
}

function AttendanceRowActions({ attendance, onEdit, onDelete }: AttendanceRowActionsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(attendance)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => onDelete(attendance)}
      >
        <Trash2Icon />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          <DropdownMenuGroup>
            <DropdownMenuItem onClick={() => onEdit(attendance)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.attendance.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(attendance)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.attendance.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseAttendanceColumnsOptions {
  onEdit: (a: Attendance) => void
  onDelete: (a: Attendance) => void
}

export function useAttendanceColumns({
  onEdit,
  onDelete,
}: UseAttendanceColumnsOptions): ColumnDef<Attendance>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const statusLabel = useAttendanceStatusLabel()

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
        <span className="text-sm">
          {row.original.class?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
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
        return <Badge variant={STATUS_BADGE_VARIANT[status] ?? "outline"}>{statusLabel(status)}</Badge>
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
      cell: ({ row }) => (
        <span className="text-muted-foreground truncate text-xs">{row.original.remarks || "—"}</span>
      ),
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
      cell: ({ row }) => (
        <AttendanceRowActions attendance={row.original} onEdit={onEdit} onDelete={onDelete} />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

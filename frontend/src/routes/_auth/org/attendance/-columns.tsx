import type {
  GithubCom4H1RZooraInternalDomainAttendance as Attendance,
  GithubCom4H1RZooraInternalDomainAttendanceStatus as AttendanceStatus,
} from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { useFormatDate } from "@/lib/data-table"

export function statusBadgeVariant(status: AttendanceStatus | undefined) {
  switch (status) {
    case "present":
      return "default" as const
    case "absent":
      return "destructive" as const
    case "late":
      return "outline" as const
    case "excused":
      return "secondary" as const
    default:
      return "ghost" as const
  }
}

export function useAttendanceColumns(): ColumnDef<Attendance>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      id: "class",
      accessorFn: (a) => a.class?.name ?? "",
      header: t("org.attendance.table.class"),
      cell: ({ getValue }) => <span className="text-sm font-medium">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "session",
      accessorFn: (a) => a.class_session?.name ?? a.class_session?.description ?? "",
      header: t("org.attendance.table.session"),
      cell: ({ getValue }) => <span className="text-muted-foreground text-sm">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "created_at",
      accessorFn: (a) => a.created_at ?? "",
      header: t("org.attendance.table.date"),
      cell: ({ row }) =>
        row.original.created_at ? (
          <span className="text-sm tabular-nums">{formatDate(row.original.created_at, "date")}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "status",
      accessorFn: (a) => a.status ?? "",
      header: t("org.attendance.table.status"),
      cell: ({ row }) => (
        <Badge variant={statusBadgeVariant(row.original.status)}>
          {t(`common.statuses.attendance.${row.original.status}`)}
        </Badge>
      ),
    },
  ]
}

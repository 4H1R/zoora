import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"
import type { RoomWindowStatus } from "./room-window"
import type { ColumnDef } from "@tanstack/react-table"

import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { formatSessionDate } from "@/lib/session-status"

import { roomWindowStatus, surfacedRoom } from "./room-window"

const STATUS_BADGE_VARIANT: Record<RoomWindowStatus, "default" | "secondary" | "outline" | "ghost"> = {
  in_progress: "default",
  not_started: "outline",
  ended: "secondary",
  not_scheduled: "ghost",
}

export function useManagerExamColumns(): ColumnDef<Quiz>[] {
  const { t, i18n } = useTranslation()

  return [
    {
      id: "title",
      accessorFn: (q) => q.title ?? "",
      header: t("org.exams.table.title"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
            <ClipboardListIcon />
          </div>
          <span className="truncate text-sm font-medium">{row.original.title || "—"}</span>
        </div>
      ),
    },
    {
      id: "class_name",
      accessorFn: (q) => q.class?.name ?? "",
      header: t("org.exams.table.class"),
      enableSorting: false,
      cell: ({ getValue }) => <span className="text-sm">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "duration_minutes",
      accessorFn: (q) => q.duration_minutes ?? 0,
      header: t("org.exams.table.duration"),
      cell: ({ row }) =>
        typeof row.original.duration_minutes === "number" ? (
          <span className="text-sm tabular-nums">
            {t("org.exams.duration", { count: row.original.duration_minutes })}
          </span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "total_score",
      accessorFn: (q) => q.total_score ?? 0,
      header: t("org.exams.table.totalScore"),
      enableSorting: false,
      cell: ({ row }) =>
        typeof row.original.total_score === "number" ? (
          <span className="text-sm tabular-nums">{row.original.total_score}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "start_time",
      accessorFn: (q) => surfacedRoom(q)?.started_at ?? "",
      header: t("org.exams.table.start"),
      enableSorting: false,
      cell: ({ row }) => {
        const startedAt = surfacedRoom(row.original)?.started_at
        return startedAt ? (
          <span className="text-sm">{formatSessionDate(startedAt, i18n.language, "short")}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )
      },
    },
    {
      id: "end_time",
      accessorFn: (q) => surfacedRoom(q)?.ended_at ?? "",
      header: t("org.exams.table.end"),
      enableSorting: false,
      cell: ({ row }) => {
        const endedAt = surfacedRoom(row.original)?.ended_at
        return endedAt ? (
          <span className="text-sm">{formatSessionDate(endedAt, i18n.language, "short")}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )
      },
    },
    {
      id: "room_status",
      accessorFn: (q) => roomWindowStatus(q),
      header: t("org.exams.table.state"),
      enableSorting: false,
      cell: ({ row }) => {
        const status = roomWindowStatus(row.original)
        return <Badge variant={STATUS_BADGE_VARIANT[status]}>{t(`org.exams.roomStatus.${status}`)}</Badge>
      },
    },
  ]
}

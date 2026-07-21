import type { GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { Link } from "@tanstack/react-router"
import { NotebookPenIcon, PencilIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"

type PracticeWindowStatus = "upcoming" | "open" | "ended"

// Reads the practice's own window against now — mirrors the backend's
// window filter buckets (upcoming/open/ended) so the status column and
// the window filter speak the same language.
function practiceWindowStatus(view: PracticeRoomView): PracticeWindowStatus {
  const now = Date.now()
  const start = view.start_time ? new Date(view.start_time).getTime() : undefined
  const end = view.end_time ? new Date(view.end_time).getTime() : undefined
  if (start !== undefined && start > now) return "upcoming"
  if (end !== undefined && end <= now) return "ended"
  return "open"
}

const STATUS_BADGE_VARIANT: Record<PracticeWindowStatus, "default" | "secondary" | "outline"> = {
  open: "default",
  upcoming: "outline",
  ended: "secondary",
}

export function usePracticeHubColumns(): ColumnDef<PracticeRoomView>[] {
  const { t, i18n } = useTranslation()

  return [
    {
      id: "title",
      accessorFn: (p) => p.title ?? "",
      header: t("org.practices.title"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
            <NotebookPenIcon />
          </div>
          <span className="truncate text-sm font-medium">{row.original.title || "—"}</span>
        </div>
      ),
      enableHiding: false,
    },
    {
      id: "class_name",
      accessorFn: (p) => p.class?.name ?? "",
      header: t("org.practices.table.class"),
      enableSorting: false,
      cell: ({ getValue }) => <span className="text-sm">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "max_score",
      accessorFn: (p) => p.max_score ?? 0,
      header: t("org.practices.table.maxScore"),
      enableSorting: false,
      cell: ({ row }) =>
        typeof row.original.max_score === "number" ? (
          <span className="text-sm tabular-nums">{formatScore(row.original.max_score)}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "start_time",
      accessorFn: (p) => p.start_time ?? "",
      header: t("org.practices.table.start"),
      enableSorting: false,
      cell: ({ row }) =>
        row.original.start_time ? (
          <span className="text-sm">{formatSessionDate(row.original.start_time, i18n.language, "short")}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "end_time",
      accessorFn: (p) => p.end_time ?? "",
      header: t("org.practices.table.end"),
      enableSorting: false,
      cell: ({ row }) =>
        row.original.end_time ? (
          <span className="text-sm">{formatSessionDate(row.original.end_time, i18n.language, "short")}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "status",
      header: t("org.practices.table.status"),
      enableSorting: false,
      cell: ({ row }) => {
        const status = practiceWindowStatus(row.original)
        return <Badge variant={STATUS_BADGE_VARIANT[status]}>{t(`org.practices.filter.window.${status}`)}</Badge>
      },
    },
    {
      id: "grading",
      header: t("org.practices.actions.enterScores"),
      enableSorting: false,
      cell: ({ row }) => {
        const practice = row.original
        if (!practice.id || !practice.can_grade) return null
        return (
          <Button
            size="sm"
            variant="outline"
            onClick={(e) => e.stopPropagation()}
            render={<Link to="/org/practices/$practiceId/scores" params={{ practiceId: practice.id }} />}
          >
            <PencilIcon data-icon="inline-start" />
            {t("org.practices.actions.enterScores")}
          </Button>
        )
      },
      enableHiding: false,
    },
  ]
}

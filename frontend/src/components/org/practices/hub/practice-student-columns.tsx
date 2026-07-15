import type { GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { NotebookPenIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { useFormatDate } from "@/lib/data-table"
import { formatScore } from "@/lib/score"

import { PracticeStatusBadge } from "./practice-status-badge"

interface Options {
  canSubmit: boolean
  onSubmit: (practice: PracticeRoomView) => void
}

export function usePracticeStudentColumns({ canSubmit, onSubmit }: Options): ColumnDef<PracticeRoomView>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "title",
      header: t("org.practices.title"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
            <NotebookPenIcon />
          </div>
          <span className="truncate text-sm font-medium">{row.original.title}</span>
        </div>
      ),
      enableHiding: false,
    },
    {
      id: "class",
      accessorFn: (p) => p.class?.name ?? "",
      header: t("org.practices.table.class"),
      cell: ({ getValue }) => <span className="text-sm">{(getValue() as string) || "—"}</span>,
      enableSorting: false,
    },
    {
      accessorKey: "end_time",
      header: t("org.practices.table.due"),
      cell: ({ row }) =>
        row.original.end_time ? (
          <span className="text-sm tabular-nums">{formatDate(row.original.end_time, "date")}</span>
        ) : (
          <span className="text-muted-foreground">{t("org.practices.noDueDate")}</span>
        ),
    },
    {
      id: "status",
      accessorFn: (p) => p.status ?? "",
      header: t("org.practices.table.status"),
      cell: ({ row }) => <PracticeStatusBadge status={row.original.status} />,
      enableSorting: false,
    },
    {
      id: "score",
      header: t("org.practices.table.score"),
      enableSorting: false,
      cell: ({ row }) =>
        row.original.status === "graded" ? (
          <span className="text-sm font-semibold tabular-nums">
            {formatScore(row.original.my_submission?.score)}
            <span className="text-muted-foreground font-normal">{` / ${formatScore(row.original.max_score ?? 0)}`}</span>
          </span>
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) =>
        canSubmit &&
        row.original.can_submit && (
          <div className="flex justify-end">
            <Button size="xs" onClick={() => onSubmit(row.original)}>
              {t("org.practices.actions.submit")}
            </Button>
          </div>
        ),
    },
  ]
}

import type { GithubCom4H1RZooraInternalDomainMyExam as MyExam } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ExamAction, examStateBadgeVariant } from "@/components/exam-card"
import { Badge } from "@/components/ui/badge"

export function useExamColumns(): ColumnDef<MyExam>[] {
  const { t } = useTranslation()

  return [
    {
      id: "title",
      accessorFn: (e) => e.title ?? "",
      header: t("org.exams.table.title"),
      cell: ({ row }) => {
        const exam = row.original
        return (
          <div className="flex items-center gap-3">
            <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
              <ClipboardListIcon />
            </div>
            <span className="truncate text-sm font-medium">{exam.title || "—"}</span>
          </div>
        )
      },
    },
    {
      id: "class_name",
      accessorFn: (e) => e.class_name ?? "",
      header: t("org.exams.table.class"),
      cell: ({ getValue }) => <span className="text-sm">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "state",
      accessorFn: (e) => e.state ?? "",
      header: t("org.exams.table.state"),
      cell: ({ row }) => (
        <Badge variant={examStateBadgeVariant(row.original.state)}>{t(`org.exams.state.${row.original.state}`)}</Badge>
      ),
    },
    {
      id: "duration",
      accessorFn: (e) => e.duration_minutes ?? 0,
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
      id: "actions",
      header: "",
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <ExamAction exam={row.original} />
        </div>
      ),
    },
  ]
}

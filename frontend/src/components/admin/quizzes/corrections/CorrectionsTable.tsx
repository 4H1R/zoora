import type { GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { CheckSquareIcon, PencilIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

interface CorrectionsTableProps {
  submissions: QuizSubmission[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onGrade: (s: QuizSubmission) => void
}

export function CorrectionsTable({
  submissions,
  total,
  isLoading,
  sorting,
  onGrade,
}: CorrectionsTableProps) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  const columns: ColumnDef<QuizSubmission>[] = [
    {
      accessorKey: "user",
      header: t("admin.corrections.student"),
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
              <div className="truncate text-start text-sm font-medium">{name}</div>
              {row.original.user?.username && (
                <div className="text-muted-foreground truncate text-start text-xs">
                  {row.original.user.username}
                </div>
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
      header: t("admin.corrections.status"),
      cell: ({ row }) => {
        const status = row.original.status ?? "in_progress"
        const variant =
          status === "graded" ? "default" : status === "submitted" ? "secondary" : "outline"
        return (
          <Badge variant={variant}>
            {t(`admin.corrections.statuses.${status}`)}
          </Badge>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "total_score",
      header: t("admin.corrections.score"),
      cell: ({ row }) => (
        <span className="text-sm font-medium tabular-nums">
          {(row.original.total_score ?? 0).toFixed(2)}
        </span>
      ),
      enableSorting: true,
    },
    {
      accessorKey: "submitted_at",
      header: t("admin.corrections.submittedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">
          {formatDate(row.original.submitted_at)}
        </span>
      ),
      enableSorting: true,
    },
    {
      accessorKey: "started_at",
      header: t("admin.corrections.startedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">
          {formatDate(row.original.started_at)}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const status = row.original.status
        const disabled = status === "in_progress"
        return (
          <div className="flex items-center justify-end">
            <Button
              variant="ghost"
              size="sm"
              disabled={disabled}
              onClick={() => onGrade(row.original)}
            >
              <PencilIcon data-icon="inline-start" />
              {t("admin.corrections.actions.grade")}
            </Button>
          </div>
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
  ]

  const table = useAdminTable({
    data: submissions,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        sortLabel={t("admin.quizzes.toolbar.sort")}
        columnsLabel={t("admin.quizzes.toolbar.columns")}
        toggleColumnsLabel={t("admin.quizzes.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<CheckSquareIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.corrections.noResults")}
            emptyHint={t("admin.corrections.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

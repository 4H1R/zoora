import type {
  GithubCom4H1RZooraInternalDomainSubmissionAntiCheatReport as AntiCheatReport,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"
import type { CellContext, ColumnDef, SortingState } from "@tanstack/react-table"
import type { VariantProps } from "class-variance-authority"

import { CheckSquareIcon, PencilIcon } from "lucide-react"
import { createContext, useContext } from "react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge, badgeVariants } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { TooltipProvider } from "@/components/ui/tooltip"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { formatScore } from "@/lib/score"
import { cn } from "@/lib/utils"

import { IntegrityCell } from "./exam-integrity"

function getBadgeVariant(status: string): VariantProps<typeof badgeVariants>["variant"] {
  if (status === "graded") return "default"
  if (status === "submitted") return "secondary"
  return "outline"
}

// Table-wide values the cell components need but that don't live on a row.
// Threaded through context so each cell can be a stable module-scope component
// (flexRender renders `cell` with only the tanstack CellContext as props).
type CorrectionsCellData = {
  quizMaxScore?: number
  reports?: Map<string, AntiCheatReport>
  onGrade: (s: QuizSubmission) => void
}

const CorrectionsCellCtx = createContext<CorrectionsCellData>({ onGrade: () => {} })

function StudentCell({ row }: CellContext<QuizSubmission, unknown>) {
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
          <div className="text-muted-foreground truncate text-start text-xs">{row.original.user.username}</div>
        )}
      </div>
    </div>
  )
}

function StatusCell({ row }: CellContext<QuizSubmission, unknown>) {
  const { t } = useTranslation()
  const status = row.original.status ?? "in_progress"
  const variant = getBadgeVariant(status)
  return <Badge variant={variant}>{t(`admin.corrections.statuses.${status}`)}</Badge>
}

function ScoreCell({ row }: CellContext<QuizSubmission, unknown>) {
  const { t } = useTranslation()
  const { quizMaxScore } = useContext(CorrectionsCellCtx)
  return (
    <span className="text-sm font-medium tabular-nums">
      {formatScore(row.original.total_score ?? 0)}
      {quizMaxScore != null && quizMaxScore > 0 && (
        <span className="text-muted-foreground ms-1 font-normal">
          {t("admin.corrections.scoreOf", { max: formatScore(quizMaxScore) })}
        </span>
      )}
    </span>
  )
}

function IntegrityColumnCell({ row }: CellContext<QuizSubmission, unknown>) {
  const { reports } = useContext(CorrectionsCellCtx)
  return <IntegrityCell report={reports?.get(row.original.id ?? "")} submission={row.original} />
}

function SubmittedAtCell({ row }: CellContext<QuizSubmission, unknown>) {
  const formatDate = useFormatDate()
  return <span className="text-muted-foreground text-xs">{formatDate(row.original.submitted_at)}</span>
}

function StartedAtCell({ row }: CellContext<QuizSubmission, unknown>) {
  const formatDate = useFormatDate()
  return <span className="text-muted-foreground text-xs">{formatDate(row.original.started_at)}</span>
}

function ActionsCell({ row }: CellContext<QuizSubmission, unknown>) {
  const { t } = useTranslation()
  const { onGrade } = useContext(CorrectionsCellCtx)
  const status = row.original.status
  const disabled = status === "in_progress"
  return (
    <div className="flex items-center justify-end">
      <Button variant="ghost" size="sm" disabled={disabled} onClick={() => onGrade(row.original)}>
        <PencilIcon data-icon="inline-start" />
        {t("admin.corrections.actions.grade")}
      </Button>
    </div>
  )
}

interface CorrectionsTableProps {
  submissions: QuizSubmission[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onGrade: (s: QuizSubmission) => void
  quizMaxScore?: number
  reports?: Map<string, AntiCheatReport>
}

export function CorrectionsTable({
  submissions,
  total,
  isLoading,
  sorting,
  onGrade,
  quizMaxScore,
  reports,
}: CorrectionsTableProps) {
  const { t } = useTranslation()

  const columns: ColumnDef<QuizSubmission>[] = [
    {
      accessorKey: "user",
      header: t("admin.corrections.student"),
      cell: StudentCell,
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "status",
      header: t("admin.corrections.status"),
      cell: StatusCell,
      enableSorting: false,
    },
    {
      accessorKey: "total_score",
      header: t("admin.corrections.score"),
      cell: ScoreCell,
      enableSorting: true,
    },
    {
      id: "integrity",
      header: t("admin.corrections.integrity.column"),
      cell: IntegrityColumnCell,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "submitted_at",
      header: t("admin.corrections.submittedAt"),
      cell: SubmittedAtCell,
      enableSorting: true,
    },
    {
      accessorKey: "started_at",
      header: t("admin.corrections.startedAt"),
      cell: StartedAtCell,
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ActionsCell,
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
    <CorrectionsCellCtx.Provider value={{ quizMaxScore, reports, onGrade }}>
      <TableFilter
        table={table}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <TooltipProvider>
            <DataTable
              table={table}
              isLoading={isLoading}
              emptyIcon={<CheckSquareIcon className="size-8 opacity-40" />}
              emptyTitle={t("admin.corrections.noResults")}
              emptyHint={t("admin.corrections.noResultsHint")}
            />
          </TooltipProvider>
        </div>
        <DataTablePagination table={table} />
      </Card>
    </CorrectionsCellCtx.Provider>
  )
}

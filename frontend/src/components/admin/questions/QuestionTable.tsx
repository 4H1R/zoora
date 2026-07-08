import type { GithubCom4H1RZooraInternalDomainQuestion as Question } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import { HelpCircleIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { useAdminTable, useFormatDate } from "@/lib/data-table"

import { QuestionActions } from "./QuestionActions"

type QType = "descriptive" | "short_answer" | "choice"

const TYPE_BADGE_VARIANT: Record<QType, "default" | "secondary" | "outline"> = {
  choice: "default",
  short_answer: "secondary",
  descriptive: "outline",
}

function useTypeLabel() {
  const { t } = useTranslation()
  return (type?: string) => {
    if (!type) return "—"
    return t(`admin.questions.types.${type}`, { defaultValue: type })
  }
}

function useQuestionColumns({
  onEdit,
}: {
  onEdit: (q: Question) => void
}): ColumnDef<Question>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const typeLabel = useTypeLabel()

  return [
    {
      accessorKey: "text",
      header: t("admin.questions.text"),
      cell: ({ row }) => (
        <div className="min-w-0">
          <div className="line-clamp-2 text-start text-sm font-medium">
            {row.original.text}
          </div>
          {row.original.bank?.name && (
            <div className="text-muted-foreground truncate text-start text-xs">
              {row.original.bank.name}
            </div>
          )}
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "type",
      header: t("admin.questions.type"),
      cell: ({ row }) => {
        const type = (row.original.type ?? "descriptive") as QType
        return (
          <Badge variant={TYPE_BADGE_VARIANT[type] ?? "outline"}>{typeLabel(type)}</Badge>
        )
      },
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "options_count",
      header: t("admin.questions.optionsCount"),
      cell: ({ row }) => (
        <span className="text-xs tabular-nums">{row.original.options?.length ?? 0}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "total_score",
      header: t("admin.questions.totalScore"),
      cell: ({ row }) => {
        const total = (row.original.options ?? []).reduce(
          (sum, o) => sum + (o.score ?? 0),
          0
        )
        return <span className="text-xs tabular-nums">{total}</span>
      },
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.questions.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">
          {formatDate(row.original.created_at)}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <QuestionActions question={row.original} onEdit={onEdit} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface QuestionTableProps {
  questions: Question[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (q: Question) => void
}

export function QuestionTable({
  questions,
  total,
  isLoading,
  sorting,
  onEdit,
}: QuestionTableProps) {
  const { t } = useTranslation()
  const columns = useQuestionColumns({ onEdit })

  const table = useAdminTable({
    data: questions,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.questions.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<HelpCircleIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.questions.noResults")}
            emptyHint={t("admin.questions.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

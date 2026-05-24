import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"
import type { ColumnDef, SortingState } from "@tanstack/react-table"

import {
  ClipboardListIcon,
  ClockIcon,
  LockKeyholeIcon,
  ShuffleIcon,
  TrophyIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { getEntityColor, getInitials, useAdminTable, useFormatDate } from "@/lib/data-table"
import { formatScore } from "@/lib/score"
import { cn } from "@/lib/utils"

import { QuizActions } from "./QuizActions"

function useQuizColumns({
  onEdit,
  onManageQuestions,
}: {
  onEdit: (q: Quiz) => void
  onManageQuestions: (q: Quiz) => void
}): ColumnDef<Quiz>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "title",
      header: t("admin.quizzes.title_col"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.title ?? "")
            )}
          >
            {getInitials(row.original.title ?? "")}
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className="truncate text-start text-sm font-medium">
                {row.original.title}
              </span>
              {row.original.no_back_navigation && (
                <Badge
                  variant="secondary"
                  className="h-4 gap-0.5 px-1 text-[10px] font-normal"
                  title={t("admin.quizzes.flags.noBackNavigation")}
                >
                  <LockKeyholeIcon className="size-2.5" />
                  {t("admin.quizzes.flags.noBackShort")}
                </Badge>
              )}
              {row.original.shuffle_questions && (
                <Badge
                  variant="secondary"
                  className="h-4 gap-0.5 px-1 text-[10px] font-normal"
                  title={t("admin.quizzes.flags.shuffleQuestions")}
                >
                  <ShuffleIcon className="size-2.5" />
                  {t("admin.quizzes.flags.shuffleShort")}
                </Badge>
              )}
            </div>
            {row.original.description && (
              <div className="text-muted-foreground truncate text-start text-xs">
                {row.original.description}
              </div>
            )}
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "class",
      header: t("admin.quizzes.class"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.class?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "duration_minutes",
      header: t("admin.quizzes.duration"),
      cell: ({ row }) => (
        <span className="inline-flex items-center gap-1.5 text-xs tabular-nums">
          <ClockIcon className="text-muted-foreground size-3.5" />
          {row.original.duration_minutes ?? 0} {t("admin.quizzes.minutesShort")}
        </span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "total_score",
      header: t("admin.quizzes.totalScore"),
      cell: ({ row }) => (
        <span className="inline-flex items-center gap-1.5 text-xs tabular-nums">
          <TrophyIcon className="text-muted-foreground size-3.5" />
          {formatScore(row.original.total_score ?? 0)}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "user",
      header: t("admin.quizzes.teacher"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.user?.name ?? <span className="text-muted-foreground">—</span>}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.quizzes.createdAt"),
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
      cell: ({ row }) => (
        <QuizActions
          quiz={row.original}
          onEdit={onEdit}
          onManageQuestions={onManageQuestions}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

interface QuizTableProps {
  quizzes: Quiz[]
  total: number
  isLoading: boolean
  sorting: SortingState
  onEdit: (q: Quiz) => void
  onManageQuestions: (q: Quiz) => void
}

export function QuizTable({
  quizzes,
  total,
  isLoading,
  sorting,
  onEdit,
  onManageQuestions,
}: QuizTableProps) {
  const { t } = useTranslation()
  const columns = useQuizColumns({ onEdit, onManageQuestions })

  const table = useAdminTable({
    data: quizzes,
    columns,
    rowCount: total,
    sorting,
  })

  return (
    <>
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.quizzes.searchPlaceholder")}
        sortLabel={t("admin.quizzes.toolbar.sort")}
        columnsLabel={t("admin.quizzes.toolbar.columns")}
        toggleColumnsLabel={t("admin.quizzes.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ClipboardListIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.quizzes.noResults")}
            emptyHint={t("admin.quizzes.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </>
  )
}

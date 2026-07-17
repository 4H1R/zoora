import { useNavigate } from "@tanstack/react-router"
import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetQuizzesMe } from "@/api/quizzes/quizzes"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { ExamCard, ExamCardSkeleton } from "@/components/exam-card"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { useViewMode, ViewModeToggle } from "@/components/view-mode-toggle"
import { useAdminTable } from "@/lib/data-table"
import { Route } from "@/routes/_auth/org/exams/index"

import { ClassFilterSelect } from "./class-session-filters"
import { useStudentExamColumns } from "./student-exam-columns"

// Ordered by urgency, mirroring the server's default sort.
const STATE_FILTERS = ["all", "open", "upcoming", "submitted", "graded"] as const

export function StudentExamsView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, status, class_id, order_by, order_dir, page, page_size } = Route.useSearch()

  const activeState = status ?? "all"
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const { viewMode, setViewMode, isTable } = useViewMode()

  const { data, isLoading } = useGetQuizzesMe({
    search: search || undefined,
    class_id: class_id || undefined,
    state: activeState === "all" ? undefined : activeState,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
    page: page ?? 1,
    page_size: page_size ?? 20,
  })

  const listData = (data?.status === 200 && data.data.data) || undefined
  const exams = listData?.items ?? []
  const total = listData?.total ?? 0

  const columns = useStudentExamColumns()
  const table = useAdminTable({ data: exams, columns, rowCount: total, sorting })

  const setState = (value: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, status: value === "all" ? undefined : value, page: 1 }) })
  const setClass = (classId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_id: classId, page: 1 }) })

  const emptyTitle = activeState === "all" && !search && !class_id ? t("org.exams.empty") : t("org.exams.noResults")

  const renderContent = () => {
    if (isTable) {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={isLoading}
              emptyIcon={<ClipboardListIcon className="size-8 opacity-40" />}
              emptyTitle={emptyTitle}
              emptyHint={t("org.exams.noResultsHint")}
            />
          </div>
          <DataTablePagination table={table} />
        </Card>
      )
    }

    if (isLoading) {
      return (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <ExamCardSkeleton key={i} />
          ))}
        </div>
      )
    }

    if (exams.length === 0) {
      return (
        <Card className="flex flex-col items-center gap-2 px-4 py-16 text-center">
          <ClipboardListIcon className="text-muted-foreground/40 size-8" />
          <p className="text-sm font-medium">{emptyTitle}</p>
          <p className="text-muted-foreground text-xs">{t("org.exams.noResultsHint")}</p>
        </Card>
      )
    }

    return (
      <div className="flex flex-col gap-3">
        {exams.map((exam) => (
          <ExamCard key={exam.quiz_id} exam={exam} />
        ))}
        {total > exams.length && (
          <Card className="gap-0 overflow-hidden p-0">
            <DataTablePagination table={table} />
          </Card>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.exams.title")} />
        <p className="text-muted-foreground text-sm">{t("org.exams.subtitle")}</p>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("org.exams.searchPlaceholder")}
            sortLabel={t("common.toolbar.sort")}
            columnsLabel={t("common.toolbar.columns")}
            toggleColumnsLabel={t("common.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          >
            <div className="flex flex-wrap gap-1.5">
              {STATE_FILTERS.map((value) => (
                <Button
                  key={value}
                  size="sm"
                  variant={activeState === value ? "default" : "outline"}
                  onClick={() => setState(value)}
                >
                  {value === "all" ? t("org.exams.filter.all") : t(`org.exams.state.${value}`)}
                </Button>
              ))}
            </div>
            <ClassFilterSelect value={class_id} onChange={setClass} />
          </TableFilter>
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}
    </div>
  )
}

import type { GithubCom4H1RZooraInternalDomainMyExam as MyExam } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetQuizzesMe } from "@/api/quizzes/quizzes"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { ExamCard, ExamCardSkeleton } from "@/components/exam-card"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { useViewMode, ViewModeToggle } from "@/components/view-mode-toggle"
import { useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useClientTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useExamColumns } from "./-columns"

export const Route = createFileRoute("/_auth/org/exams/")({
  head: () => orgHead("org.nav.exams"),
  validateSearch: adminSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["quizzes:view", "quizzes:take"])

  const examsQ = useGetQuizzesMe(undefined, { query: { enabled: allowed } })
  const exams: MyExam[] = (examsQ.data?.status === 200 && examsQ.data.data.data?.items) || []
  const loading = examsQ.isPending

  const { viewMode, setViewMode, isTable } = useViewMode()
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useExamColumns()

  const table = useClientTable({
    data: exams,
    columns,
    sorting,
    globalFilter: search,
    page,
    pageSize: page_size ?? 20,
  })

  if (!allowed) return null

  const renderContent = () => {
    if (isTable) {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={loading}
              emptyIcon={<ClipboardListIcon className="size-8 opacity-40" />}
              emptyTitle={t("org.exams.empty")}
            />
          </div>
          <DataTablePagination table={table} />
        </Card>
      )
    }

    if (loading) {
      return (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <ExamCardSkeleton key={i} />
          ))}
        </div>
      )
    }

    const rows = table.getPrePaginationRowModel().rows
    if (rows.length === 0) {
      return <EmptyState icon={ClipboardListIcon} title={t("org.exams.empty")} />
    }

    return (
      <div className="flex flex-col gap-3">
        {rows.map((row) => (
          <ExamCard key={row.original.quiz_id} exam={row.original} />
        ))}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.exams.title")} />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("org.exams.searchPlaceholder")}
            sortLabel={t("common.toolbar.sort")}
            columnsLabel={t("common.toolbar.columns")}
            toggleColumnsLabel={t("common.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          />
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}
    </div>
  )
}

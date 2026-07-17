import { useNavigate } from "@tanstack/react-router"
import { ClipboardListIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { useAdminTable } from "@/lib/data-table"
import { Route } from "@/routes/_auth/org/exams/index"

import { ClassFilterSelect, SessionFilterSelect } from "./class-session-filters"
import { useManagerExamColumns } from "./manager-exam-columns"

export function ManagerExamsView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, class_id, class_session_id, order_by, order_dir, page, page_size } = Route.useSearch()

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const { data, isLoading } = useGetQuizzes({
    search: search || undefined,
    class_id: class_id || undefined,
    class_session_id: class_session_id || undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
    page: page ?? 1,
    page_size: page_size ?? 20,
  })

  const listData = (data?.status === 200 && data.data.data) || undefined
  const rows = listData?.items ?? []
  const total = listData?.total ?? 0

  const columns = useManagerExamColumns()
  const table = useAdminTable({ data: rows, columns, rowCount: total, sorting })

  // Changing class invalidates any chosen session — sessions belong to one class.
  const setClass = (classId?: string) =>
    navigate({
      to: ".",
      search: (prev) => ({ ...prev, class_id: classId, class_session_id: undefined, page: 1 }),
    })
  const setSession = (sessionId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_session_id: sessionId, page: 1 }) })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.exams.manage.title")} />
        <p className="text-muted-foreground text-sm">{t("org.exams.manage.subtitle")}</p>
      </div>

      <TableFilter
        table={table}
        searchPlaceholder={t("org.exams.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      >
        <ClassFilterSelect value={class_id} onChange={setClass} />
        <SessionFilterSelect classId={class_id} value={class_session_id} onChange={setSession} />
      </TableFilter>

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ClipboardListIcon className="size-8 opacity-40" />}
            emptyTitle={t("org.exams.noResults")}
            emptyHint={t("org.exams.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}

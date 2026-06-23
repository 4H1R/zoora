import { createFileRoute } from "@tanstack/react-router"
import { GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetGradebookMe } from "@/api/gradebook/gradebook"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { ViewModeToggle, useViewMode } from "@/components/view-mode-toggle"
import { useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useClientTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useGradeColumns, type GradeRow } from "./-columns"

export const Route = createFileRoute("/_auth/org/$orgId/grades/")({
  head: () => orgHead("org.nav.grades"),
  validateSearch: adminSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["gradebook:view"])

  const gradesQ = useGetGradebookMe({ query: { enabled: allowed } })
  const gradebook = (gradesQ.data?.status === 200 && gradesQ.data.data.data) || undefined
  const classes = gradebook?.classes ?? []
  const loading = gradesQ.isPending

  // Flatten the gradebook matrix to one row per graded item for the table view.
  const rows: GradeRow[] = classes.flatMap((cls) =>
    (cls.columns ?? []).map((col) => ({
      classId: cls.class_id ?? "",
      className: cls.class_name ?? "",
      item: col.title ?? "",
      value: (col.id ? cls.cells?.[col.id] : undefined) ?? "",
    }))
  )

  const { viewMode, setViewMode, isTable } = useViewMode()
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useGradeColumns()
  const table = useClientTable({
    data: rows,
    columns,
    sorting,
    globalFilter: search,
    page,
    pageSize: page_size ?? 8,
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
              emptyIcon={<GraduationCapIcon className="size-8 opacity-40" />}
              emptyTitle={t("org.grades.empty")}
            />
          </div>
          <DataTablePagination table={table} />
        </Card>
      )
    }

    if (loading) {
      return (
        <div className="flex flex-col gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <Card key={i} className="gap-0 overflow-hidden p-0">
              <div className="border-b px-4 py-3">
                <Skeleton className="h-4 w-40" />
              </div>
              <div className="flex flex-col gap-3 p-4">
                {Array.from({ length: 3 }).map((__, j) => (
                  <div key={j} className="flex items-center justify-between">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-4 w-12" />
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </div>
      )
    }

    if (classes.length === 0) {
      return <EmptyState icon={GraduationCapIcon} title={t("org.grades.empty")} />
    }

    return (
      <div className="flex flex-col gap-4">
        {classes.map((cls) => {
          const cols = cls.columns ?? []
          return (
            <Card key={cls.class_id} className="gap-0 overflow-hidden p-0">
              <div className="flex items-center gap-2 border-b px-4 py-3">
                <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
                  <GraduationCapIcon />
                </div>
                <h2 className="truncate text-sm font-semibold">{cls.class_name || "—"}</h2>
              </div>
              {cols.length === 0 ? (
                <p className="text-muted-foreground px-4 py-8 text-center text-sm">{t("org.grades.noColumns")}</p>
              ) : (
                <table className="w-full text-sm">
                  <tbody className="divide-y">
                    {cols.map((col) => {
                      const value = col.id ? cls.cells?.[col.id] : undefined
                      return (
                        <tr key={col.id}>
                          <td className="text-muted-foreground px-4 py-2.5 text-start">{col.title || "—"}</td>
                          <td className="px-4 py-2.5 text-end font-medium tabular-nums">
                            {value && value.trim() ? value : "—"}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              )}
            </Card>
          )
        })}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.grades.title")} />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("org.grades.searchPlaceholder")}
            sortLabel={t("org.grades.toolbar.sort")}
            columnsLabel={t("org.grades.toolbar.columns")}
            toggleColumnsLabel={t("org.grades.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          />
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}
    </div>
  )
}

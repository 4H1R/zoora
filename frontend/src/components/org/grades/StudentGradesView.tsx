import type { GradeRow } from "./grade-columns"

import { useNavigate } from "@tanstack/react-router"
import { GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetGradebookMe } from "@/api/gradebook/gradebook"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { ClassFilterSelect } from "@/components/org/class-session-filters"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useViewMode, ViewModeToggle } from "@/components/view-mode-toggle"
import { useClientTable } from "@/lib/data-table"
import { Route } from "@/routes/_auth/org/grades/index"

import { useGradeColumns } from "./grade-columns"

export function StudentGradesView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, class_id, order_by, order_dir, page, page_size } = Route.useSearch()

  const gradesQ = useGetGradebookMe()
  const gradebook = (gradesQ.data?.status === 200 && gradesQ.data.data.data) || undefined
  const classes = gradebook?.classes ?? []
  const loading = gradesQ.isPending

  // The gradebook payload already holds every class — the class filter is
  // applied client-side, and its options come from the payload itself.
  const visibleClasses = class_id ? classes.filter((cls) => cls.class_id === class_id) : classes

  // Flatten the gradebook matrix to one row per graded item for the table view.
  const rows: GradeRow[] = visibleClasses.flatMap((cls) =>
    (cls.columns ?? []).map((col) => ({
      classId: cls.class_id ?? "",
      className: cls.class_name ?? "",
      item: col.title ?? "",
      value: (col.id ? cls.cells?.[col.id] : undefined) ?? "",
      maxScore: col.max_score,
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
    pageSize: page_size ?? 20,
  })

  const setClass = (classId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_id: classId, page: 1 }) })

  const hasFilters = !!(search || class_id)

  const renderContent = () => {
    if (isTable) {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={loading}
              emptyIcon={<GraduationCapIcon className="size-8 opacity-40" />}
              emptyTitle={hasFilters ? t("org.grades.noResults") : t("org.grades.empty")}
              emptyHint={hasFilters ? t("org.grades.noResultsHint") : undefined}
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

    if (visibleClasses.length === 0) {
      return (
        <EmptyState icon={GraduationCapIcon} title={hasFilters ? t("org.grades.noResults") : t("org.grades.empty")} />
      )
    }

    return (
      <div className="flex flex-col gap-4">
        {visibleClasses.map((cls) => {
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
                            {value && value.trim() ? (
                              <span dir={col.max_score != null ? "ltr" : undefined}>
                                {value}
                                {col.max_score != null && (
                                  <span className="text-muted-foreground font-normal"> / {col.max_score}</span>
                                )}
                              </span>
                            ) : (
                              "—"
                            )}
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
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.grades.title")} />
        <p className="text-muted-foreground text-sm">{t("org.grades.subtitle")}</p>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("org.grades.searchPlaceholder")}
            sortLabel={t("common.toolbar.sort")}
            columnsLabel={t("common.toolbar.columns")}
            toggleColumnsLabel={t("common.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          >
            <ClassFilterSelect
              value={class_id}
              onChange={setClass}
              classes={classes.map((cls) => ({ id: cls.class_id, name: cls.class_name }))}
            />
          </TableFilter>
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}
    </div>
  )
}

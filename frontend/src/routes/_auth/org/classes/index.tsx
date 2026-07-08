import { createFileRoute } from "@tanstack/react-router"
import { LayoutGrid, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetClasses } from "@/api/classes/classes"
import { ClassCard, ClassCardSkeleton } from "@/components/class-card"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { useClassPermissions } from "@/components/org/classes/use-class-permissions"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { ViewModeToggle, useViewMode } from "@/components/view-mode-toggle"
import { useOrgGuard } from "@/lib/access"
import { useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useClassColumns } from "./-columns"
import { CreateClassDialog } from "./-create-class-dialog"

const classesSearchSchema = z.object({
  search: z.string().optional(),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(20),
})

export const Route = createFileRoute("/_auth/org/classes/")({
  head: () => orgHead("org.nav.classes"),
  validateSearch: classesSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const { canView, canCreate } = useClassPermissions()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])

  const currentPage = page ?? 1

  const { data, isPending } = useGetClasses(
    {
      search: search || undefined,
      order_by: order_by || undefined,
      order_dir: order_dir || undefined,
      page: currentPage,
    },
    { query: { enabled: canView } }
  )

  const classesData = (data?.status === 200 && data.data.data) || undefined
  const classes = classesData?.items ?? []
  const total = classesData?.total ?? 0

  const [formOpen, setFormOpen] = useState(false)
  const { viewMode, setViewMode, isTable } = useViewMode()

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useClassColumns()
  const table = useAdminTable({ data: classes, columns, rowCount: total, sorting })

  const renderContent = () => {
    if (isTable) {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={isPending}
              emptyTitle={t("classesPage.noResults")}
              emptyHint={t("classesPage.noResultsHint")}
            />
          </div>
          <DataTablePagination table={table} />
        </Card>
      )
    }

    if (isPending) {
      return (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }, (_, i) => (
            <ClassCardSkeleton key={i} />
          ))}
        </div>
      )
    }

    if (classes.length === 0) {
      return (
        <EmptyState
          icon={LayoutGrid}
          title={t("classesPage.noResults")}
          description={t("classesPage.noResultsHint")}
        />
      )
    }

    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {classes.map((cls) => (
          <ClassCard key={cls.id} cls={cls} />
        ))}
      </div>
    )
  }

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("classesPage.title")}
        actions={
          canCreate && (
            <Button size="sm" onClick={() => setFormOpen(true)}>
              <PlusIcon data-icon="inline-start" />
              {t("classesPage.newClass")}
            </Button>
          )
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("classesPage.searchPlaceholder")}
            sortLabel={t("common.toolbar.sort")}
            columnsLabel={t("common.toolbar.columns")}
            toggleColumnsLabel={t("common.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          />
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}

      <CreateClassDialog open={formOpen} onOpenChange={setFormOpen} />
    </div>
  )
}

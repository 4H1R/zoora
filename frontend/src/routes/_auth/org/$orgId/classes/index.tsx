import type { NavFn } from "@/lib/data-table"

import { createFileRoute } from "@tanstack/react-router"

import { orgHead } from "@/lib/org-head"
import { getCoreRowModel, useReactTable } from "@tanstack/react-table"
import { LayoutGrid, List, PlusIcon, SearchIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"
import { useAccess } from "react-access-engine"
import { z } from "zod"

import { useGetClasses } from "@/api/classes/classes"
import { ClassCard, ClassCardSkeleton } from "@/components/class-card"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { createSortingHandler } from "@/lib/data-table"

import { useClassColumns } from "./-columns"
import { CreateClassDialog } from "./-create-class-dialog"

const classesSearchSchema = z.object({
  search: z.string().optional(),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(8),
})

export const Route = createFileRoute("/_auth/org/$orgId/classes/")({
  head: () => orgHead("org.nav.classes"),
  validateSearch: classesSearchSchema,
  component: RouteComponent,
})

type ViewMode = "grid" | "table"

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId } = Route.useParams()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const navigate = Route.useNavigate() as unknown as NavFn

  const currentPage = page ?? 1

  const { data, isPending } = useGetClasses({
    search: search || undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
    page: currentPage,
  })

  const classesData = (data?.status === 200 && data.data.data) || undefined
  const classes = classesData?.items ?? []
  const total = classesData?.total ?? 0

  const { can } = useAccess()
  const canCreate = can("classes:create") || can("classes:create_any")

  const [formOpen, setFormOpen] = useState(false)
  const [viewMode, setViewMode] = useState<ViewMode>("grid")
  const [localSearch, setLocalSearch] = useState(search ?? "")
  const [debouncedSearch] = useDebounce(localSearch, 300)

  useEffect(() => {
    navigate({ search: (prev) => ({ ...prev, search: debouncedSearch || undefined, page: 1 }) })
  }, [debouncedSearch])

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useClassColumns(orgId)

  const table = useReactTable({
    data: classes,
    columns,
    getCoreRowModel: getCoreRowModel(),
    rowCount: total,
    manualPagination: true,
    manualSorting: true,
    onSortingChange: createSortingHandler(navigate, sorting),
    state: { sorting },
  })

  const renderContent = () => {
    if (viewMode === "table") {
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
        <div className="text-muted-foreground flex flex-col items-center gap-2 py-16 text-center">
          <p className="text-sm font-medium">{t("classesPage.noResults")}</p>
          <p className="text-xs">{t("classesPage.noResultsHint")}</p>
        </div>
      )
    }

    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {classes.map((cls) => (
          <ClassCard key={cls.id} cls={cls} orgId={orgId} />
        ))}
      </div>
    )
  }

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
        <div className="relative max-w-xs flex-1">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2" />
          <Input
            placeholder={t("classesPage.searchPlaceholder")}
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="ps-9"
          />
        </div>

        <ToggleGroup
          value={[viewMode]}
          onValueChange={(values) => {
            const next = values.find((v) => v !== viewMode)
            if (next) setViewMode(next as ViewMode)
          }}
          className="border-border rounded-lg border"
        >
          <ToggleGroupItem value="grid" aria-label={t("classesPage.gridView")} className="px-2.5">
            <LayoutGrid className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem value="table" aria-label={t("classesPage.tableView")} className="px-2.5">
            <List className="size-4" />
          </ToggleGroupItem>
        </ToggleGroup>
      </div>

      {renderContent()}

      <CreateClassDialog open={formOpen} onOpenChange={setFormOpen} />
    </div>
  )
}

import type { GithubCom4H1RZooraInternalDomainPermission as Permission } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { KeyIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetPermissions } from "@/api/roles/roles"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { usePermissionColumns } from "./-columns"

export const Route = createFileRoute("/_admin/admin/permissions/")({
  head: () => adminHead("admin.permissions.title"),
  validateSearch: adminSearchSchema,
  component: PermissionsPage,
})

function PermissionsPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()

  const currentPage = page ?? 1
  const pageSize = page_size ?? 20

  const { data, isLoading } = useGetPermissions()
  const allPermissions: Permission[] = (data?.data as { data?: Permission[] } | undefined)?.data ?? []

  const filtered = (() => {
    const q = search?.toLowerCase().trim()
    if (!q) return allPermissions
    return allPermissions.filter((p) => p.name?.toLowerCase().includes(q))
  })()

  const sorted = (() => {
    if (!order_by) return filtered
    return [...filtered].sort((a, b) => {
      const aVal = (a as Record<string, unknown>)[order_by] as string | undefined
      const bVal = (b as Record<string, unknown>)[order_by] as string | undefined
      const cmp = (aVal ?? "").localeCompare(bVal ?? "")
      return order_dir === "desc" ? -cmp : cmp
    })
  })()

  const total = sorted.length
  const paged = sorted.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = usePermissionColumns()
  const table = useAdminTable({ data: paged, columns, rowCount: total, sorting })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.permissions.title")} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.permissions.searchPlaceholder")}
        sortLabel={t("admin.permissions.toolbar.sort")}
        columnsLabel={t("admin.permissions.toolbar.columns")}
        toggleColumnsLabel={t("admin.permissions.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<KeyIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.permissions.noResults")}
            emptyHint={t("admin.permissions.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}

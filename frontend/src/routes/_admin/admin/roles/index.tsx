import type { GithubCom4H1RZooraInternalDomainRole as Role } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { KeyIcon, PlusIcon, ShieldIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminRolesQueryKey,
  getGetAdminRolesStatsQueryKey,
  useDeleteAdminRolesId,
  useGetAdminRoles,
  useGetAdminRolesStats,
} from "@/api/admin-roles/admin-roles"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { useAdminStore } from "@/stores/admin"

import { useRoleColumns } from "./-columns"
import { RoleFormDialog } from "./-role-form-dialog"

export const Route = createFileRoute("/_admin/admin/roles/")({
  head: () => adminHead("admin.roles.title"),
  validateSearch: adminSearchSchema,
  component: RolesPage,
})

function RolesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const activeOrganizationId = useAdminStore((s) => s.activeOrganizationId)

  const currentPage = page ?? 1

  const [formOpen, setFormOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null)

  const handleEdit = (role: Role) => {
    setEditingRole(role)
    setFormOpen(true)
  }

  const handleDelete = (role: Role) => {
    setDeleteTarget(role)
  }

  const handleCreate = () => {
    setEditingRole(null)
    setFormOpen(true)
  }

  const { data, isLoading } = useGetAdminRoles({
    search: search || undefined,
    organization_id: activeOrganizationId ?? undefined,
    page: currentPage,
    page_size: page_size ?? undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const { data: statsData, isLoading: statsLoading } = useGetAdminRolesStats({
    organization_id: activeOrganizationId ?? undefined,
  })

  const deleteMutation = useDeleteAdminRolesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.roles.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminRolesQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetAdminRolesStatsQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const rolesData = (data?.status === 200 && data.data.data) || undefined
  const roles = rolesData?.items ?? []
  const total = rolesData?.total ?? 0
  const stats = (statsData?.status === 200 && statsData.data.data) || undefined

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useRoleColumns({ onEdit: handleEdit, onDelete: handleDelete })

  const table = useAdminTable({
    data: roles,
    columns,
    rowCount: total,
    sorting,
  })

  const statCards = [
    {
      icon: <ShieldIcon />,
      label: t("admin.roles.stats.total"),
      value: stats?.total_roles,
      loading: statsLoading,
    },
    {
      icon: <KeyIcon />,
      label: t("admin.roles.stats.permissions"),
      value: stats?.total_permissions,
      loading: statsLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.roles.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.roles.newRole")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.roles.searchPlaceholder")}
        sortLabel={t("admin.roles.toolbar.sort")}
        columnsLabel={t("admin.roles.toolbar.columns")}
        toggleColumnsLabel={t("admin.roles.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ShieldIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.roles.noResults")}
            emptyHint={t("admin.roles.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <RoleFormDialog open={formOpen} onOpenChange={setFormOpen} role={editingRole} />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget?.name ?? ""}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}

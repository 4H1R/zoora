import type { GithubCom4H1RZooraInternalDomainRole as Role } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { KeyIcon, PlusIcon, ShieldIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetRolesQueryKey,
  getGetRolesStatsQueryKey,
  useDeleteRolesId,
  useGetRoles,
  useGetRolesStats,
} from "@/api/roles/roles"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { useRoleColumns } from "./-columns"
import { RoleFormDialog } from "./-role-form-dialog"

export const Route = createFileRoute("/_auth/org/$orgId/roles/")({
  validateSearch: adminSearchSchema,
  component: RolesPage,
})

function RolesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { orgId } = Route.useParams()
  const {} = Route.useSearch()

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

  const { data, isLoading } = useGetRoles({
    organization_id: orgId,
  })

  const { data: statsData, isLoading: statsLoading } = useGetRolesStats({
    organization_id: orgId,
  })

  const deleteMutation = useDeleteRolesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.roles.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetRolesQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetRolesStatsQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const roles = (data?.status === 200 && data.data.data) || []
  const stats = (statsData?.status === 200 && statsData.data.data) || undefined

  const columns = useRoleColumns({ onEdit: handleEdit, onDelete: handleDelete })

  const table = useAdminTable({
    data: roles,
    columns,
    rowCount: roles.length,
    sorting: [],
  })

  const statCards = [
    {
      icon: <ShieldIcon />,
      label: t("org.roles.stats.total"),
      value: stats?.total_roles,
      loading: statsLoading,
    },
    {
      icon: <KeyIcon />,
      label: t("org.roles.stats.permissions"),
      value: stats?.total_permissions,
      loading: statsLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("org.roles.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("org.roles.newRole")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("org.roles.searchPlaceholder")}
        sortLabel={t("org.roles.toolbar.sort")}
        columnsLabel={t("org.roles.toolbar.columns")}
        toggleColumnsLabel={t("org.roles.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ShieldIcon className="size-8 opacity-40" />}
            emptyTitle={t("org.roles.noResults")}
            emptyHint={t("org.roles.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <RoleFormDialog open={formOpen} onOpenChange={setFormOpen} role={editingRole} organizationId={orgId} />

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

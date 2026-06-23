import type { GithubCom4H1RZooraInternalDomainRole as Role } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"

import { orgHead } from "@/lib/org-head"
import { useOrgGuard } from "@/lib/access"
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
import { useGetUsersMe } from "@/api/users/users"
import { useRolePermissions } from "@/components/org/roles/use-role-permissions"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { useRoleName } from "@/lib/permissions"

import { useRoleColumns } from "./-columns"
import { RoleFormDialog } from "./-role-form-dialog"

export const Route = createFileRoute("/_auth/org/roles/")({
  head: () => orgHead("org.nav.roles"),
  validateSearch: adminSearchSchema,
  component: RolesPage,
})

function RolesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: meResponse } = useGetUsersMe()
  const orgId = (meResponse?.status === 200 && meResponse.data.data?.organization_id) || ""
  const { search } = Route.useSearch()
  const roleName = useRoleName()
  const { canView, canCreate } = useRolePermissions()
  const allowed = useOrgGuard("roles:view")

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

  const { data, isLoading } = useGetRoles({ query: { enabled: canView } })

  const { data: statsData, isLoading: statsLoading } = useGetRolesStats({
    query: { enabled: canView },
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

  const allRoles = (data?.status === 200 && data.data.data) || []
  const stats = (statsData?.status === 200 && statsData.data.data) || undefined

  const q = (search ?? "").trim().toLowerCase()
  const roles = q
    ? allRoles.filter((r) => {
        const raw = (r.name ?? "").toLowerCase()
        const label = (r.name ? roleName(r.name) : "").toLowerCase()
        return raw.includes(q) || label.includes(q)
      })
    : allRoles

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

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("org.roles.title")}
        actions={
          canCreate ? (
            <Button size="sm" onClick={handleCreate}>
              <PlusIcon data-icon="inline-start" />
              {t("org.roles.newRole")}
            </Button>
          ) : null
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

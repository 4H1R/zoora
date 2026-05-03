import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { Can } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetUsersQueryKey, useDeleteUsersId, useGetUsers } from "@/api/users/users"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { useRolesMap } from "@/hooks/use-roles-map"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { useUserColumns } from "./-columns"
import { UserFormDialog } from "./-user-form-dialog"

export const Route = createFileRoute("/_auth/org/$orgId/users/")({
  validateSearch: adminSearchSchema,
  component: UsersPage,
})

function UsersPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { orgId } = Route.useParams()
  const { page, page_size } = Route.useSearch()

  const currentPage = page ?? 1
  const pageSize = page_size ?? 8

  const [formOpen, setFormOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null)

  const handleEdit = (user: User) => {
    setEditingUser(user)
    setFormOpen(true)
  }

  const handleDelete = (user: User) => {
    setDeleteTarget(user)
  }

  const handleCreate = () => {
    setEditingUser(null)
    setFormOpen(true)
  }

  const { data, isLoading } = useGetUsers({
    page: currentPage,
    page_size: pageSize,
    organization_id: orgId,
  })

  const { rolesMap } = useRolesMap(orgId)

  const deleteMutation = useDeleteUsersId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.users.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetUsersQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const usersData = (data?.status === 200 && data.data.data) || undefined
  const users = usersData?.items ?? []
  const total = usersData?.total ?? 0

  const columns = useUserColumns({ onEdit: handleEdit, onDelete: handleDelete, rolesMap })

  const table = useAdminTable({
    data: users,
    columns,
    rowCount: total,
    sorting: [],
  })

  const statCards = [
    {
      icon: <UsersIcon />,
      label: t("org.users.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("org.users.title")}
        actions={
          <Can perform="users:create">
            <Button size="sm" onClick={handleCreate}>
              <PlusIcon data-icon="inline-start" />
              {t("org.users.newUser")}
            </Button>
          </Can>
        }
      />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("org.users.searchPlaceholder")}
        sortLabel={t("org.users.toolbar.sort")}
        columnsLabel={t("org.users.toolbar.columns")}
        toggleColumnsLabel={t("org.users.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<UsersIcon className="size-8 opacity-40" />}
            emptyTitle={t("org.users.noResults")}
            emptyHint={t("org.users.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <UserFormDialog open={formOpen} onOpenChange={setFormOpen} user={editingUser} organizationId={orgId} />

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

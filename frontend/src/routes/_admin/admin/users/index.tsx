import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetAdminUsersQueryKey, useDeleteAdminUsersId, useGetAdminUsers } from "@/api/admin-users/admin-users"
import { useDisableAdminUser, useEnableAdminUser } from "@/api/users/users-disable"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { DisableConfirmDialog } from "@/components/form/disable-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { useAdminStore } from "@/stores/admin"

import { useUserColumns } from "./-columns"
import { UserFormDialog } from "./-user-form-dialog"

export const Route = createFileRoute("/_admin/admin/users/")({
  head: () => adminHead("admin.users.title"),
  validateSearch: adminSearchSchema,
  component: UsersPage,
})

function UsersPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const activeOrganizationId = useAdminStore((s) => s.activeOrganizationId)

  const currentPage = page ?? 1

  const [formOpen, setFormOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null)
  const [disableTarget, setDisableTarget] = useState<User | null>(null)
  const [disableReason, setDisableReason] = useState("")

  const handleEdit = (user: User) => {
    setEditingUser(user)
    setFormOpen(true)
  }

  const handleDelete = (user: User) => {
    setDeleteTarget(user)
  }

  const handleDisable = (user: User) => {
    setDisableReason("")
    setDisableTarget(user)
  }

  const handleCreate = () => {
    setEditingUser(null)
    setFormOpen(true)
  }

  const { data, isLoading } = useGetAdminUsers({
    search: search || undefined,
    organization_id: activeOrganizationId ?? undefined,
    page: currentPage,
    page_size: page_size ?? undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const deleteMutation = useDeleteAdminUsersId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.users.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const disableMutation = useDisableAdminUser({
    onSuccess: () => {
      toast.success(t("admin.users.form.disableSuccess"))
      queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
      setDisableTarget(null)
    },
  })

  const enableMutation = useEnableAdminUser({
    onSuccess: () => {
      toast.success(t("admin.users.form.enableSuccess"))
      queryClient.invalidateQueries({ queryKey: getGetAdminUsersQueryKey() })
    },
  })

  const handleEnable = (user: User) => {
    if (user.id) enableMutation.mutate({ id: user.id })
  }

  const usersData = (data?.status === 200 && data.data.data) || undefined
  const users = usersData?.items ?? []
  const total = usersData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useUserColumns({
    onEdit: handleEdit,
    onDelete: handleDelete,
    onDisable: handleDisable,
    onEnable: handleEnable,
  })

  const table = useAdminTable({
    data: users,
    columns,
    rowCount: total,
    sorting,
  })

  const statCards = [
    {
      icon: <UsersIcon />,
      label: t("admin.users.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.users.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.users.newUser")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.users.searchPlaceholder")}
        sortLabel={t("admin.users.toolbar.sort")}
        columnsLabel={t("admin.users.toolbar.columns")}
        toggleColumnsLabel={t("admin.users.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<UsersIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.users.noResults")}
            emptyHint={t("admin.users.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <UserFormDialog open={formOpen} onOpenChange={setFormOpen} user={editingUser} />

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

      <DisableConfirmDialog
        open={!!disableTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDisableTarget(null)
        }}
        resourceName={disableTarget?.name ?? ""}
        reason={disableReason}
        onReasonChange={setDisableReason}
        onConfirm={() => {
          if (disableTarget?.id) disableMutation.mutate({ id: disableTarget.id, reason: disableReason })
        }}
        isLoading={disableMutation.isPending}
      />
    </div>
  )
}

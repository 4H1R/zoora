import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetUsersCountsQueryKey, getGetUsersQueryKey, useGetUsers, useGetUsersCounts, useDeleteUsersId } from "@/api/users/users"
import { useDisableUser, useEnableUser } from "@/api/users/users-disable"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { DisableConfirmDialog } from "@/components/form/disable-confirm-dialog"
import { useUserPermissions } from "@/components/org/users/use-user-permissions"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { useRolesMap } from "@/hooks/use-roles-map"
import { useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useUserColumns } from "./-columns"
import { UserFormDialog } from "./-user-form-dialog"

export const Route = createFileRoute("/_auth/org/$orgId/users/")({
  head: () => orgHead("org.nav.users"),
  validateSearch: adminSearchSchema,
  component: UsersPage,
})

function UsersPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { orgId } = Route.useParams()
  const { search, status, order_by, order_dir, page, page_size } = Route.useSearch()
  const { canView, canCreate } = useUserPermissions()
  const allowed = useOrgGuard(["users:view", "users:view_any"])

  const statusFilter = status ?? "all"
  const disabled = statusFilter === "active" ? false : statusFilter === "disabled" ? true : undefined
  const currentPage = page ?? 1
  const pageSize = page_size ?? 8
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

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

  const { data, isLoading } = useGetUsers(
    {
      search: search || undefined,
      order_by: order_by || undefined,
      order_dir: order_dir || undefined,
      disabled,
      page: currentPage,
      page_size: pageSize,
    },
    { query: { enabled: canView } }
  )

  const { data: countsData } = useGetUsersCounts({ query: { enabled: canView } })
  const counts = (countsData?.status === 200 && countsData.data.data) || undefined

  const { rolesMap } = useRolesMap()

  const deleteMutation = useDeleteUsersId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.users.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetUsersQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetUsersCountsQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const disableMutation = useDisableUser({
    onSuccess: () => {
      toast.success(t("org.users.form.disableSuccess"))
      queryClient.invalidateQueries({ queryKey: getGetUsersQueryKey() })
      setDisableTarget(null)
    },
  })

  const enableMutation = useEnableUser({
    onSuccess: () => {
      toast.success(t("org.users.form.enableSuccess"))
      queryClient.invalidateQueries({ queryKey: getGetUsersQueryKey() })
    },
  })

  const handleEnable = (user: User) => {
    if (user.id) enableMutation.mutate({ id: user.id })
  }

  const usersData = (data?.status === 200 && data.data.data) || undefined
  const users = usersData?.items ?? []
  const total = usersData?.total ?? 0

  const columns = useUserColumns({
    onEdit: handleEdit,
    onDelete: handleDelete,
    onDisable: handleDisable,
    onEnable: handleEnable,
    rolesMap,
  })

  const table = useAdminTable({
    data: users,
    columns,
    rowCount: total,
    sorting,
  })

  const statusTabs = [
    { value: "all", label: t("org.users.tabs.all"), count: counts?.all },
    { value: "active", label: t("org.users.tabs.active"), count: counts?.active },
    { value: "disabled", label: t("org.users.tabs.disabled"), count: counts?.disabled },
  ]

  const statCards = [
    {
      icon: <UsersIcon />,
      label: t("org.users.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("org.users.title")}
        actions={
          canCreate ? (
            <Button size="sm" onClick={handleCreate}>
              <PlusIcon data-icon="inline-start" />
              {t("org.users.newUser")}
            </Button>
          ) : null
        }
      />
      <StatCards stats={statCards} />
      <StatusTabs tabs={statusTabs} />
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

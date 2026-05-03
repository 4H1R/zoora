import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { ActivityIcon, Building2Icon, PlusIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminOrganizationsQueryKey,
  getGetAdminOrganizationsStatsQueryKey,
  useDeleteAdminOrganizationsId,
  useGetAdminOrganizations,
  useGetAdminOrganizationsStats,
} from "@/api/admin-organizations/admin-organizations"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { useOrgColumns } from "./-columns"
import { OrgFormDialog } from "./-org-form-dialog"

export const Route = createFileRoute("/_admin/admin/organizations/")({
  head: () => adminHead("admin.orgs.title"),
  validateSearch: adminSearchSchema,
  component: OrganizationsPage,
})

function OrganizationsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, status, order_by, order_dir, page, page_size } = Route.useSearch()

  const statusFilter = status ?? "all"
  const currentPage = page ?? 1

  const [formOpen, setFormOpen] = useState(false)
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Organization | null>(null)

  const handleEdit = (org: Organization) => {
    setEditingOrg(org)
    setFormOpen(true)
  }

  const handleDelete = (org: Organization) => {
    setDeleteTarget(org)
  }

  const handleCreate = () => {
    setEditingOrg(null)
    setFormOpen(true)
  }

  const { data, isLoading } = useGetAdminOrganizations({
    search: search || undefined,
    status: statusFilter !== "all" ? statusFilter : undefined,
    page: currentPage,
    page_size: page_size ?? undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })
  const { data: statsData, isLoading: statsLoading } = useGetAdminOrganizationsStats()

  const deleteMutation = useDeleteAdminOrganizationsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.orgs.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsStatsQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const orgsData = (data?.status === 200 && data.data.data) || undefined
  const organizations = orgsData?.items ?? []
  const total = orgsData?.total ?? 0
  const stats = (statsData?.status === 200 && statsData.data.data) || undefined

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useOrgColumns({ onEdit: handleEdit, onDelete: handleDelete })

  const table = useAdminTable({
    data: organizations,
    columns,
    rowCount: total,
    sorting,
  })

  const statusTabs = [
    { value: "all", label: t("admin.orgs.tabs.all"), count: stats?.total_organizations },
    { value: "active", label: t("admin.orgs.tabs.active"), count: stats?.active_count },
    { value: "trial", label: t("admin.orgs.tabs.trial"), count: stats?.trial_count },
    { value: "suspended", label: t("admin.orgs.tabs.suspended"), count: stats?.suspended_count },
    { value: "archived", label: t("admin.orgs.tabs.archived"), count: stats?.archived_count },
  ]

  const statCards = [
    {
      icon: <Building2Icon />,
      label: t("admin.orgs.stats.total"),
      value: stats?.total_organizations,
      loading: statsLoading,
    },
    {
      icon: <ActivityIcon />,
      label: t("admin.orgs.stats.active"),
      value: stats?.active_count,
      loading: statsLoading,
    },
    {
      icon: <UsersIcon />,
      label: t("admin.orgs.stats.users"),
      value: stats?.total_users,
      loading: statsLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.orgs.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.orgs.newOrg")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <StatusTabs tabs={statusTabs} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.orgs.searchPlaceholder")}
        sortLabel={t("admin.orgs.toolbar.sort")}
        columnsLabel={t("admin.orgs.toolbar.columns")}
        toggleColumnsLabel={t("admin.orgs.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<Building2Icon className="size-8 opacity-40" />}
            emptyTitle={t("admin.orgs.noResults")}
            emptyHint={t("admin.orgs.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <OrgFormDialog open={formOpen} onOpenChange={setFormOpen} organization={editingOrg} />

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

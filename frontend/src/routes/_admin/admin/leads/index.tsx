import type { GithubCom4H1RZooraInternalDomainLead as Lead } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { InboxIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminLeadsQueryKey,
  useDeleteAdminLeadsId,
  useGetAdminLeads,
  usePatchAdminLeadsIdStatus,
} from "@/api/admin-leads/admin-leads"
import { GithubCom4H1RZooraInternalDomainLeadStatus as LeadStatus } from "@/api/model"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { useLeadColumns } from "./-columns"
import { ConvertDialog } from "./-convert-dialog"

export const Route = createFileRoute("/_admin/admin/leads/")({
  head: () => adminHead("admin.leads.title"),
  validateSearch: adminSearchSchema,
  component: LeadsPage,
})

function LeadsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, status, order_by, order_dir, page, page_size } = Route.useSearch()

  const statusFilter = status ?? "all"
  const currentPage = page ?? 1

  const [convertOpen, setConvertOpen] = useState(false)
  const [convertLead, setConvertLead] = useState<Lead | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Lead | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetAdminLeadsQueryKey() })

  const { data, isLoading } = useGetAdminLeads({
    search: search || undefined,
    status: statusFilter !== "all" ? statusFilter : undefined,
    page: currentPage,
    page_size: page_size ?? undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const statusMutation = usePatchAdminLeadsIdStatus({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.leads.statusUpdated"))
        invalidate()
      },
    },
  })

  const deleteMutation = useDeleteAdminLeadsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.leads.deleteSuccess"))
        invalidate()
        setDeleteTarget(null)
      },
    },
  })

  const handleConvert = (lead: Lead) => {
    setConvertLead(lead)
    setConvertOpen(true)
  }

  const handleSetStatus = (lead: Lead, next: string) => {
    if (!lead.id) return
    statusMutation.mutate({ id: lead.id, data: { status: next as LeadStatus } })
  }

  const leadsData = (data?.status === 200 && data.data.data) || undefined
  const leads = leadsData?.items ?? []
  const total = leadsData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useLeadColumns({ onConvert: handleConvert, onSetStatus: handleSetStatus, onDelete: setDeleteTarget })

  const table = useAdminTable({
    data: leads,
    columns,
    rowCount: total,
    sorting,
  })

  const statusTabs = [
    { value: "all", label: t("admin.leads.tabs.all") },
    { value: LeadStatus.LeadStatusNew, label: t("admin.leads.tabs.new") },
    { value: LeadStatus.LeadStatusContacted, label: t("admin.leads.tabs.contacted") },
    { value: LeadStatus.LeadStatusConverted, label: t("admin.leads.tabs.converted") },
    { value: LeadStatus.LeadStatusRejected, label: t("admin.leads.tabs.rejected") },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.leads.title")} />
      <StatusTabs tabs={statusTabs} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.leads.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<InboxIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.leads.noResults")}
            emptyHint={t("admin.leads.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <ConvertDialog open={convertOpen} onOpenChange={setConvertOpen} lead={convertLead} />

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

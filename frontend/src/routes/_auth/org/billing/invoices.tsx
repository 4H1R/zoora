import type { GithubCom4H1RZooraInternalDomainInvoice as Invoice } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ReceiptTextIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetBillingInvoices } from "@/api/billing/billing"
import { GithubCom4H1RZooraInternalDomainInvoiceStatus as InvoiceStatus } from "@/api/model"
import { InvoiceStatusBadge } from "@/components/billing/invoice-status-badge"
import { ReceiptButton } from "@/components/billing/receipt-button"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { useOrgGuard } from "@/lib/access"
import { useFormatToman } from "@/lib/billing"
import { adminSearchSchema, useAdminTable, useFormatDate } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/billing/invoices")({
  head: () => orgHead("billing.invoicesTitle"),
  validateSearch: adminSearchSchema,
  component: InvoicesPage,
})

function useInvoiceColumns(): ColumnDef<Invoice>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const formatToman = useFormatToman()

  return [
    {
      accessorKey: "number",
      header: t("billing.invoiceNumber"),
      cell: ({ row }) => <span className="font-medium tabular-nums">{row.original.number ?? "—"}</span>,
      enableSorting: false,
    },
    {
      accessorKey: "created_at",
      header: t("billing.date"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: false,
    },
    {
      accessorKey: "total",
      header: t("billing.amount"),
      cell: ({ row }) => (
        <span className="tabular-nums">
          {formatToman(row.original.total)} <span className="text-muted-foreground text-xs">{t("billing.toman")}</span>
        </span>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "status",
      header: t("billing.status"),
      cell: ({ row }) => <InvoiceStatusBadge status={row.original.status} />,
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        row.original.status === InvoiceStatus.InvoiceStatusPaid && row.original.id ? (
          <div className="flex justify-end">
            <ReceiptButton invoiceId={row.original.id} />
          </div>
        ) : null,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

function InvoicesPage() {
  const { t } = useTranslation()
  const allowed = useOrgGuard("billing:manage")
  const { status, page } = Route.useSearch()

  const statusFilter = status ?? "all"
  const currentPage = page ?? 1

  const { data, isLoading } = useGetBillingInvoices({
    status: statusFilter !== "all" ? statusFilter : undefined,
    page: currentPage,
  })
  const pageData = (data?.status === 200 && data.data.data) || undefined
  const invoices = pageData?.items ?? []
  const total = pageData?.total ?? 0

  const columns = useInvoiceColumns()
  const table = useAdminTable({ data: invoices, columns, rowCount: total, sorting: [] })

  // Build the status filter from the known invoice statuses so filtering mirrors admin lists.
  const tabs = [
    { value: "all", label: t("admin.orgs.tabs.all") },
    ...Object.values(InvoiceStatus).map((s) => ({ value: s, label: t(`billing.statuses.${s}`) })),
  ]

  if (!allowed) return null

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6">
      <PageHeader
        title={t("billing.invoicesTitle")}
        actions={
          <Link to="/org/billing" className="text-primary text-sm font-medium hover:underline">
            {t("billing.backToPlans")}
          </Link>
        }
      />

      <StatusTabs tabs={tabs} />

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ReceiptTextIcon className="size-8 opacity-40" />}
            emptyTitle={t("billing.noInvoices")}
            emptyHint={t("billing.noInvoicesHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}

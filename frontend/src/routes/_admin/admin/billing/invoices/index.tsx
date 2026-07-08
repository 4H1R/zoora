import type { GithubCom4H1RZooraInternalDomainInvoice as Invoice } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { BanIcon, PlusIcon, ReceiptTextIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminBillingInvoicesQueryKey,
  useGetAdminBillingInvoices,
  usePostAdminBillingInvoicesIdCancel,
  usePostAdminBillingInvoicesIdIssue,
} from "@/api/admin-billing/admin-billing"
import { GithubCom4H1RZooraInternalDomainInvoiceStatus as InvoiceStatus } from "@/api/model"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { PageHeader } from "@/components/page-header"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Spinner } from "@/components/ui/spinner"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { useAdminStore } from "@/stores/admin"

import { useInvoiceColumns } from "./-columns"
import { CreateInvoiceDialog, MarkPaidDialog, RefundDialog } from "./-invoice-dialogs"

export const Route = createFileRoute("/_admin/admin/billing/invoices/")({
  head: () => adminHead("billing.admin.invoicesTitle"),
  validateSearch: adminSearchSchema,
  component: AdminInvoicesPage,
})

function AdminInvoicesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { status, page } = Route.useSearch()
  const activeOrganizationId = useAdminStore((s) => s.activeOrganizationId)

  const statusFilter = status ?? "all"
  const currentPage = page ?? 1

  const [createOpen, setCreateOpen] = useState(false)
  const [markPaidTarget, setMarkPaidTarget] = useState<Invoice | null>(null)
  const [refundTarget, setRefundTarget] = useState<Invoice | null>(null)
  const [cancelTarget, setCancelTarget] = useState<Invoice | null>(null)

  const { data, isLoading } = useGetAdminBillingInvoices({
    organization_id: activeOrganizationId ?? undefined,
    status: statusFilter !== "all" ? statusFilter : undefined,
    page: currentPage,
  })
  const pageData = (data?.status === 200 && data.data.data) || undefined
  const invoices = pageData?.items ?? []
  const total = pageData?.total ?? 0

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetAdminBillingInvoicesQueryKey() })

  const issueMutation = usePostAdminBillingInvoicesIdIssue({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.issued"))
        invalidate()
      },
    },
  })

  const cancelMutation = usePostAdminBillingInvoicesIdCancel({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.canceled"))
        invalidate()
        setCancelTarget(null)
      },
    },
  })

  const columns = useInvoiceColumns({
    onIssue: (inv) => {
      if (inv.id) issueMutation.mutate({ id: inv.id })
    },
    onMarkPaid: (inv) => setMarkPaidTarget(inv),
    onCancel: (inv) => setCancelTarget(inv),
    onRefund: (inv) => setRefundTarget(inv),
  })

  const table = useAdminTable({ data: invoices, columns, rowCount: total, sorting: [] })

  const tabs = [
    { value: "all", label: t("admin.orgs.tabs.all") },
    ...Object.values(InvoiceStatus).map((s) => ({ value: s, label: t(`billing.statuses.${s}`) })),
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("billing.admin.invoicesTitle")}
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon data-icon="inline-start" />
            {t("billing.admin.createInvoice")}
          </Button>
        }
      />

      <StatusTabs tabs={tabs} />

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ReceiptTextIcon className="size-8 opacity-40" />}
            emptyTitle={t("billing.admin.noInvoices")}
            emptyHint={t("billing.admin.noInvoicesHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <CreateInvoiceDialog open={createOpen} onOpenChange={setCreateOpen} />
      <MarkPaidDialog
        open={!!markPaidTarget}
        onOpenChange={(o) => !o && setMarkPaidTarget(null)}
        invoice={markPaidTarget}
      />
      <RefundDialog open={!!refundTarget} onOpenChange={(o) => !o && setRefundTarget(null)} invoice={refundTarget} />

      <AlertDialog open={!!cancelTarget} onOpenChange={(o) => !o && setCancelTarget(null)}>
        <AlertDialogContent onOutsideClick={() => !cancelMutation.isPending && setCancelTarget(null)}>
          <AlertDialogHeader>
            <AlertDialogMedia className="bg-destructive/10 text-destructive">
              <BanIcon />
            </AlertDialogMedia>
            <AlertDialogTitle>{t("billing.admin.cancel")}</AlertDialogTitle>
            <AlertDialogDescription>{t("billing.admin.cancelConfirm")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={cancelMutation.isPending}>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={cancelMutation.isPending}
              onClick={() => {
                if (cancelTarget?.id) cancelMutation.mutate({ id: cancelTarget.id })
              }}
            >
              {cancelMutation.isPending && <Spinner />}
              {t("billing.admin.cancel")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

import type { GithubCom4H1RZooraInternalDomainInvoice as Invoice } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { BanIcon, CheckCircle2Icon, DownloadIcon, EllipsisVerticalIcon, RotateCcwIcon, SendIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getBillingInvoicesIdReceipt } from "@/api/billing/billing"
import { GithubCom4H1RZooraInternalDomainInvoiceStatus as InvoiceStatus } from "@/api/model"
import { InvoiceStatusBadge } from "@/components/billing/invoice-status-badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useFormatToman } from "@/lib/billing"
import { useFormatDate } from "@/lib/data-table"

export interface InvoiceRowActions {
  onIssue: (inv: Invoice) => void
  onMarkPaid: (inv: Invoice) => void
  onCancel: (inv: Invoice) => void
  onRefund: (inv: Invoice) => void
}

function RowActions({ invoice, actions }: { invoice: Invoice; actions: InvoiceRowActions }) {
  const { t } = useTranslation()
  const status = invoice.status
  const isDraft = status === InvoiceStatus.InvoiceStatusDraft
  const isPending = status === InvoiceStatus.InvoiceStatusPending
  const isPaid = status === InvoiceStatus.InvoiceStatusPaid

  const openReceipt = async () => {
    if (!invoice.id) return
    try {
      const res = await getBillingInvoicesIdReceipt(invoice.id)
      const url = res.status === 200 ? res.data.data?.url : undefined
      if (url) window.open(url, "_blank", "noopener,noreferrer")
      else toast.error(t("billing.receiptError"))
    } catch {
      toast.error(t("billing.receiptError"))
    }
  }

  const hasAny = isDraft || isPending || isPaid
  if (!hasAny) return null

  return (
    <div className="flex justify-end">
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          {isDraft && (
            <DropdownMenuItem onClick={() => actions.onIssue(invoice)}>
              <SendIcon data-icon="inline-start" />
              {t("billing.admin.issue")}
            </DropdownMenuItem>
          )}
          {(isDraft || isPending) && (
            <DropdownMenuItem onClick={() => actions.onMarkPaid(invoice)}>
              <CheckCircle2Icon data-icon="inline-start" />
              {t("billing.admin.markPaid")}
            </DropdownMenuItem>
          )}
          {(isDraft || isPending) && (
            <DropdownMenuItem variant="destructive" onClick={() => actions.onCancel(invoice)}>
              <BanIcon data-icon="inline-start" />
              {t("billing.admin.cancel")}
            </DropdownMenuItem>
          )}
          {isPaid && (
            <>
              <DropdownMenuItem onClick={openReceipt}>
                <DownloadIcon data-icon="inline-start" />
                {t("billing.downloadReceipt")}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem variant="destructive" onClick={() => actions.onRefund(invoice)}>
                <RotateCcwIcon data-icon="inline-start" />
                {t("billing.admin.refund")}
              </DropdownMenuItem>
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

export function useInvoiceColumns(actions: InvoiceRowActions): ColumnDef<Invoice>[] {
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
      accessorKey: "organization_id",
      header: t("billing.admin.org"),
      cell: ({ row }) => (
        <code className="bg-muted text-muted-foreground rounded px-1.5 py-0.5 font-mono text-[11px]">
          {row.original.organization_id?.slice(0, 8) ?? "—"}
        </code>
      ),
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
      accessorKey: "created_at",
      header: t("billing.admin.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <RowActions invoice={row.original} actions={actions} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

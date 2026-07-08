import type { GithubCom4H1RZooraInternalDomainPlanPrice as PlanPrice } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { PencilIcon, PlusIcon, TagIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminBillingPricesQueryKey,
  useDeleteAdminBillingPricesId,
  useGetAdminBillingPrices,
} from "@/api/admin-billing/admin-billing"
import { DataTable } from "@/components/data-table/data-table"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { useFormatToman } from "@/lib/billing"
import { useAdminTable } from "@/lib/data-table"
import { planSize, planTier } from "@/lib/plan"

import { PriceFormDialog } from "./-price-form-dialog"

export const Route = createFileRoute("/_admin/admin/billing/prices")({
  head: () => adminHead("billing.admin.pricesTitle"),
  component: PricesPage,
})

function usePriceColumns(opts: {
  onEdit: (p: PlanPrice) => void
  onDelete: (p: PlanPrice) => void
}): ColumnDef<PlanPrice>[] {
  const { t } = useTranslation()
  const formatToman = useFormatToman()

  return [
    {
      accessorKey: "plan",
      header: t("billing.admin.plan"),
      cell: ({ row }) => (
        <Badge variant="default" className="text-[11px]">
          {t(`plans.tiers.${planTier(row.original.plan)}`, { defaultValue: row.original.plan })}
          <span className="tabular-nums">{planSize(row.original.plan)}</span>
        </Badge>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "interval",
      header: t("billing.admin.interval"),
      cell: ({ row }) => (
        <span className="text-sm capitalize">
          {t(`billing.intervals.${row.original.interval}`, { defaultValue: row.original.interval })}
        </span>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "amount",
      header: t("billing.admin.amountToman"),
      cell: ({ row }) => (
        <span className="tabular-nums">
          {formatToman(row.original.amount)} <span className="text-muted-foreground text-xs">{t("billing.toman")}</span>
        </span>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "currency",
      header: t("billing.admin.currency"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{row.original.currency ?? "IRR"}</span>,
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-0.5">
          <Button variant="ghost" size="icon-xs" onClick={() => opts.onEdit(row.original)}>
            <PencilIcon />
          </Button>
          <Button
            variant="ghost"
            size="icon-xs"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => opts.onDelete(row.original)}
          >
            <Trash2Icon />
          </Button>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

function PricesPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<PlanPrice | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<PlanPrice | null>(null)

  const { data, isLoading } = useGetAdminBillingPrices()
  const prices = (data?.status === 200 && data.data.data) || []

  const deleteMutation = useDeleteAdminBillingPricesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.priceDeactivated"))
        queryClient.invalidateQueries({ queryKey: getGetAdminBillingPricesQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const columns = usePriceColumns({
    onEdit: (p) => {
      setEditing(p)
      setFormOpen(true)
    },
    onDelete: (p) => setDeleteTarget(p),
  })

  const table = useAdminTable({ data: prices, columns, rowCount: prices.length, sorting: [] })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("billing.admin.pricesTitle")}
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <PlusIcon data-icon="inline-start" />
            {t("billing.admin.savePrice")}
          </Button>
        }
      />

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<TagIcon className="size-8 opacity-40" />}
            emptyTitle={t("billing.admin.noPrices")}
            emptyHint={t("billing.admin.noPricesHint")}
          />
        </div>
      </Card>

      <PriceFormDialog open={formOpen} onOpenChange={setFormOpen} price={editing} />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget ? `${deleteTarget.plan} · ${deleteTarget.interval}` : ""}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}

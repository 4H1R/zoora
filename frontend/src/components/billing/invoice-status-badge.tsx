import { useTranslation } from "react-i18next"

import { GithubCom4H1RZooraInternalDomainInvoiceStatus as InvoiceStatus } from "@/api/model"
import { Badge } from "@/components/ui/badge"

const STATUS_VARIANT: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  [InvoiceStatus.InvoiceStatusDraft]: "outline",
  [InvoiceStatus.InvoiceStatusPending]: "secondary",
  [InvoiceStatus.InvoiceStatusPaid]: "default",
  [InvoiceStatus.InvoiceStatusCanceled]: "destructive",
  [InvoiceStatus.InvoiceStatusExpired]: "outline",
  [InvoiceStatus.InvoiceStatusRefunded]: "destructive",
}

export function InvoiceStatusBadge({ status }: { status?: string }) {
  const { t } = useTranslation()
  if (!status) return <span className="text-muted-foreground">—</span>
  return (
    <Badge variant={STATUS_VARIANT[status] ?? "secondary"} className="text-[11px] capitalize">
      {t(`billing.statuses.${status}`, { defaultValue: status })}
    </Badge>
  )
}

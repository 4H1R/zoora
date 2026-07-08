import { DownloadIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getBillingInvoicesIdReceipt } from "@/api/billing/billing"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"

// Fetches the presigned receipt URL on demand and opens it in a new tab. Shared
// by the org history table and the admin invoices table. The receipt endpoint is
// org-scoped but the admin table only offers it for the currently-scoped org's
// invoices, so the same fetcher works for both.
export function ReceiptButton({
  invoiceId,
  variant = "ghost",
  size = "icon-xs",
  withLabel = false,
}: {
  invoiceId: string
  variant?: React.ComponentProps<typeof Button>["variant"]
  size?: React.ComponentProps<typeof Button>["size"]
  withLabel?: boolean
}) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)

  const open = async () => {
    setLoading(true)
    try {
      const res = await getBillingInvoicesIdReceipt(invoiceId)
      const url = res.status === 200 ? res.data.data?.url : undefined
      if (url) window.open(url, "_blank", "noopener,noreferrer")
      else toast.error(t("billing.receiptError"))
    } catch {
      toast.error(t("billing.receiptError"))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Button variant={variant} size={size} onClick={open} disabled={loading} aria-label={t("billing.downloadReceipt")}>
      {loading ? <Spinner /> : <DownloadIcon />}
      {withLabel && t("billing.downloadReceipt")}
    </Button>
  )
}

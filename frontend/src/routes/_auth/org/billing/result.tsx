import { createFileRoute, Link } from "@tanstack/react-router"
import { CheckCircle2Icon, ClockIcon, DownloadIcon, XCircleIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getBillingInvoicesIdReceipt } from "@/api/billing/billing"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { orgHead } from "@/lib/org-head"

const searchSchema = z.object({
  status: z.enum(["success", "failed", "pending", "error"]).optional(),
  invoice: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/billing/result")({
  head: () => orgHead("billing.title"),
  validateSearch: searchSchema,
  component: ResultPage,
})

function ResultPage() {
  const { t } = useTranslation()
  const { status, invoice } = Route.useSearch()
  const [downloading, setDownloading] = useState(false)

  const isSuccess = status === "success"

  const openReceipt = async () => {
    if (!invoice) return
    setDownloading(true)
    try {
      const res = await getBillingInvoicesIdReceipt(invoice)
      const url = res.status === 200 ? res.data.data?.url : undefined
      if (url) window.open(url, "_blank", "noopener,noreferrer")
      else toast.error(t("billing.receiptError"))
    } catch {
      toast.error(t("billing.receiptError"))
    } finally {
      setDownloading(false)
    }
  }

  const config = isSuccess
    ? {
        icon: CheckCircle2Icon,
        wrap: "bg-success/10 text-success",
        title: t("billing.result.successTitle"),
        body: t("billing.result.successBody"),
      }
    : status === "pending"
      ? {
          icon: ClockIcon,
          wrap: "bg-warning/10 text-warning",
          title: t("billing.result.pendingTitle"),
          body: t("billing.result.pendingBody"),
        }
      : status === "failed"
        ? {
            icon: XCircleIcon,
            wrap: "bg-destructive/10 text-destructive",
            title: t("billing.result.failedTitle"),
            body: t("billing.result.failedBody"),
          }
        : {
            icon: XCircleIcon,
            wrap: "bg-destructive/10 text-destructive",
            title: t("billing.result.errorTitle"),
            body: t("billing.result.errorBody"),
          }

  const Icon = config.icon

  return (
    <div className="mx-auto flex w-full max-w-md flex-col items-center gap-6 py-12 text-center">
      <div className={`grid size-20 place-items-center rounded-full ${config.wrap}`}>
        <Icon className="size-10" />
      </div>
      <div className="flex flex-col gap-2">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">{config.title}</h1>
        <p className="text-muted-foreground text-sm leading-relaxed">{config.body}</p>
      </div>

      <div className="flex w-full flex-col gap-2">
        {isSuccess ? (
          <>
            {invoice && (
              <Button className="w-full" onClick={openReceipt} disabled={downloading}>
                {downloading ? <Spinner /> : <DownloadIcon />}
                {t("billing.downloadReceipt")}
              </Button>
            )}
            <Button variant="outline" className="w-full" render={<Link to="/org/billing" />}>
              {t("billing.result.backToBilling")}
            </Button>
          </>
        ) : (
          <Button className="w-full" render={<Link to="/org/billing" />}>
            {t("billing.result.tryAgain")}
          </Button>
        )}
      </div>
    </div>
  )
}

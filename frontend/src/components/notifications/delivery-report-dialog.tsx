import { UsersIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetNotificationsIdReport } from "@/api/notifications/notifications"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

const CHANNEL_KEY: Record<string, string> = {
  telegram: "notifications.connectors.telegram",
  bale: "notifications.connectors.bale",
  sms: "notifications.connectors.sms",
  push: "notifications.connectors.push",
}

interface DeliveryReportDialogProps {
  notificationId: string | undefined
  title?: string
  onOpenChange: (open: boolean) => void
}

/** Per-notification delivery breakdown: total recipients plus sent / pending /
 * failed counts for every channel it fanned out to. */
export function DeliveryReportDialog({ notificationId, title, onOpenChange }: DeliveryReportDialogProps) {
  const { t } = useTranslation()
  const open = !!notificationId

  const { data, isLoading } = useGetNotificationsIdReport(notificationId ?? "", {
    query: { enabled: open },
  })
  const report = (data?.status === 200 && data.data.data) || undefined
  const channels = report?.channels ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("notifications.sentHistory.report")}</DialogTitle>
          {title && <DialogDescription className="truncate">{title}</DialogDescription>}
        </DialogHeader>

        {isLoading ? (
          <div className="flex justify-center py-10">
            <Spinner />
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="bg-muted/40 flex items-center gap-3 rounded-xl p-4">
              <div className="bg-primary/10 text-primary grid size-10 place-items-center rounded-lg">
                <UsersIcon className="size-5" />
              </div>
              <div>
                <div className="text-2xl font-semibold tabular-nums">{report?.recipients ?? 0}</div>
                <div className="text-muted-foreground text-xs">
                  {t("notifications.sentHistory.recipients")}
                </div>
              </div>
            </div>

            {channels.length > 0 && (
              <div className="overflow-hidden rounded-xl border">
                <table className="w-full text-sm">
                  <thead className="bg-muted/40 text-muted-foreground text-xs">
                    <tr>
                      <th className="px-3 py-2 text-start font-medium">
                        {t("notifications.sentHistory.channel")}
                      </th>
                      <th className="px-3 py-2 text-end font-medium">
                        {t("notifications.sentHistory.delivered")}
                      </th>
                      <th className="px-3 py-2 text-end font-medium">
                        {t("notifications.sentHistory.pending")}
                      </th>
                      <th className="px-3 py-2 text-end font-medium">
                        {t("notifications.sentHistory.failed")}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-border/60 divide-y">
                    {channels.map((ch, i) => (
                      <tr key={ch.channel ?? i}>
                        <td className="px-3 py-2 font-medium">
                          {ch.channel ? t(CHANNEL_KEY[ch.channel] ?? ch.channel) : "—"}
                        </td>
                        <Cell value={ch.sent} className="text-foreground" />
                        <Cell value={ch.pending} className="text-muted-foreground" />
                        <Cell value={ch.failed} className={cn(ch.failed ? "text-destructive" : "text-muted-foreground")} />
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

function Cell({ value, className }: { value?: number; className?: string }) {
  return <td className={cn("px-3 py-2 text-end tabular-nums", className)}>{value ?? 0}</td>
}

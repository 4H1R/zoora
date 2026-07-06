import type { GithubCom4H1RZooraInternalDomainNotification as SentNotification } from "@/api/model"

import { FileClockIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetNotificationsSent } from "@/api/notifications/notifications"
import { DeliveryReportDialog } from "@/components/notifications/delivery-report-dialog"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Spinner } from "@/components/ui/spinner"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

const PAGE_SIZE = 10

/** Paginated table of the notifications the caller has sent, each row opening a
 * delivery report. Shared by the org composer page and the admin history tab. */
export function SentHistory() {
  const { t, i18n } = useTranslation()
  const [page, setPage] = useState(1)
  const [reportFor, setReportFor] = useState<SentNotification | undefined>()

  const { data, isLoading } = useGetNotificationsSent({ page, page_size: PAGE_SIZE })
  const pageData = (data?.status === 200 && data.data.data) || undefined
  const items = (pageData?.items ?? []) as SentNotification[]
  const total = pageData?.total ?? 0

  const fmtDate = (iso?: string) =>
    iso
      ? new Date(iso).toLocaleDateString(i18n.language, { day: "numeric", month: "short", year: "numeric" })
      : ""

  return (
    <div className="flex flex-col gap-4">
      <h2 className="text-lg font-semibold tracking-tight">{t("notifications.sentHistory.title")}</h2>

      {isLoading ? (
        <div className="flex justify-center py-16">
          <Spinner />
        </div>
      ) : items.length === 0 ? (
        <EmptyState icon={FileClockIcon} title={t("notifications.sentHistory.empty")} />
      ) : (
        <div className="overflow-hidden rounded-xl border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("notifications.send.titleField")}</TableHead>
                <TableHead className="w-32">{t("time.date")}</TableHead>
                <TableHead className="w-24 text-end" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((n) => (
                <TableRow key={n.id}>
                  <TableCell className="max-w-0">
                    <div className="truncate font-medium">{n.title}</div>
                    {n.body && (
                      <div className="text-muted-foreground truncate text-xs">{n.body}</div>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs tabular-nums">
                    {fmtDate(n.created_at)}
                  </TableCell>
                  <TableCell className="text-end">
                    <Button variant="ghost" size="xs" onClick={() => setReportFor(n)}>
                      {t("notifications.sentHistory.report")}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {items.length > 0 && (
        <SectionPagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
      )}

      <DeliveryReportDialog
        notificationId={reportFor?.id}
        title={reportFor?.title}
        onOpenChange={(open) => !open && setReportFor(undefined)}
      />
    </div>
  )
}

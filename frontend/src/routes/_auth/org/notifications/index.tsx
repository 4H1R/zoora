import type { GithubCom4H1RZooraInternalDomainNotificationInboxItem as InboxItem } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { BellIcon, CheckCheckIcon, SendIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import {
  getGetNotificationsQueryKey,
  getGetNotificationsStatusQueryKey,
  useGetNotifications,
  usePostNotificationsIdRead,
  usePostNotificationsMarkAllRead,
} from "@/api/notifications/notifications"
import { NotificationList } from "@/components/notifications/notification-list"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Spinner } from "@/components/ui/spinner"
import { useCanAny } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

const PAGE_SIZE = 20

const searchSchema = z.object({
  page: z.number().int().positive().optional().default(1),
})

export const Route = createFileRoute("/_auth/org/notifications/")({
  head: () => orgHead("notifications.title"),
  validateSearch: searchSchema,
  component: NotificationsInboxPage,
})

function NotificationsInboxPage() {
  const { t } = useTranslation()
  const { page } = Route.useSearch()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const canSend = useCanAny(["notifications:send", "notifications:send_any"])

  const { data, isLoading } = useGetNotifications({ page, page_size: PAGE_SIZE })
  const pageData = (data?.status === 200 && data.data.data) || undefined
  const items = (pageData?.items ?? []) as InboxItem[]
  const total = pageData?.total ?? 0

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetNotificationsStatusQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetNotificationsQueryKey() })
  }

  const markAllRead = usePostNotificationsMarkAllRead({ mutation: { onSuccess: invalidate } })
  const markRead = usePostNotificationsIdRead({ mutation: { onSuccess: invalidate } })

  const handleItemClick = (item: InboxItem) => {
    if (!item.read_at && item.id) markRead.mutate({ id: item.id })
    const url = item.action_url
    if (!url) return
    if (/^https?:\/\//i.test(url)) window.open(url, "_blank", "noopener,noreferrer")
    else navigate({ to: url })
  }

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader
        title={t("notifications.title")}
        actions={
          <>
            <Button variant="outline" size="sm" disabled={markAllRead.isPending} onClick={() => markAllRead.mutate()}>
              <CheckCheckIcon />
              {t("notifications.markAllRead")}
            </Button>
            {canSend && (
              <Button size="sm" render={<Link to="/org/notifications/send" />}>
                <SendIcon />
                {t("notifications.send.title")}
              </Button>
            )}
          </>
        }
      />

      <div className="mt-6">
        {isLoading ? (
          <div className="flex justify-center py-20">
            <Spinner />
          </div>
        ) : items.length === 0 ? (
          <EmptyState icon={BellIcon} title={t("notifications.empty")} description={t("notifications.emptyHint")} />
        ) : (
          <div className="bg-card ring-foreground/10 divide-border/60 divide-y rounded-2xl p-1.5 ring-1">
            <NotificationList items={items} onItemClick={handleItemClick} />
          </div>
        )}
      </div>

      {items.length > 0 && (
        <div className="mt-6">
          <SectionPagination
            page={page}
            pageSize={PAGE_SIZE}
            total={total}
            onPageChange={(next) => navigate({ to: ".", search: { page: next } })}
          />
        </div>
      )}
    </div>
  )
}

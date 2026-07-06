import type { GithubCom4H1RZooraInternalDomainNotificationInboxItem as InboxItem } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { BellIcon, CheckCheckIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import {
  getGetNotificationsQueryKey,
  getGetNotificationsStatusQueryKey,
  useGetNotifications,
  usePostNotificationsIdRead,
  usePostNotificationsMarkAllRead,
} from "@/api/notifications/notifications"
import { NotificationComposer } from "@/components/notifications/notification-composer"
import { NotificationList } from "@/components/notifications/notification-list"
import { SentHistory } from "@/components/notifications/sent-history"
import { PageHeader } from "@/components/page-header"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Spinner } from "@/components/ui/spinner"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

const TABS = ["compose", "history", "inbox"] as const
type Tab = (typeof TABS)[number]

const searchSchema = z.object({
  tab: z.enum(TABS).optional().default("compose"),
})

export const Route = createFileRoute("/_admin/admin/notifications")({
  validateSearch: searchSchema,
  component: AdminNotificationsPage,
})

function AdminNotificationsPage() {
  const { t } = useTranslation()
  const { tab } = Route.useSearch()
  const navigate = useNavigate()

  return (
    <div className="mx-auto w-full max-w-3xl">
      <PageHeader title={t("notifications.title")} />

      <Tabs
        value={tab}
        onValueChange={(value) => navigate({ to: ".", search: { tab: value as Tab } })}
        className="mt-6"
      >
        <TabsList>
          <TabsTrigger value="compose">{t("notifications.send.title")}</TabsTrigger>
          <TabsTrigger value="history">{t("notifications.sentHistory.title")}</TabsTrigger>
          <TabsTrigger value="inbox">{t("notifications.title")}</TabsTrigger>
        </TabsList>

        <TabsContent value="compose" className="mt-6">
          <Card className="p-6">
            <NotificationComposer mode="admin" />
          </Card>
        </TabsContent>

        <TabsContent value="history" className="mt-6">
          <SentHistory />
        </TabsContent>

        <TabsContent value="inbox" className="mt-6">
          <AdminInbox />
        </TabsContent>
      </Tabs>
    </div>
  )
}

const PAGE_SIZE = 20

function AdminInbox() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

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
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button
          variant="outline"
          size="sm"
          disabled={markAllRead.isPending}
          onClick={() => markAllRead.mutate()}
        >
          <CheckCheckIcon />
          {t("notifications.markAllRead")}
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner />
        </div>
      ) : items.length === 0 ? (
        <EmptyState
          icon={BellIcon}
          title={t("notifications.empty")}
          description={t("notifications.emptyHint")}
        />
      ) : (
        <div className="bg-card ring-foreground/10 rounded-2xl p-1.5 ring-1">
          <NotificationList items={items} onItemClick={handleItemClick} />
        </div>
      )}

      {items.length > 0 && (
        <SectionPagination page={page} pageSize={PAGE_SIZE} total={total} onPageChange={setPage} />
      )}
    </div>
  )
}

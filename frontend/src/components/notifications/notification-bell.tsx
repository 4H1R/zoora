import type { GithubCom4H1RZooraInternalDomainNotificationInboxItem as InboxItem } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "@tanstack/react-router"
import { BellIcon, CheckCheckIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getGetNotificationsQueryKey,
  getGetNotificationsStatusQueryKey,
  useGetNotifications,
  useGetNotificationsStatus,
  usePostNotificationsIdRead,
  usePostNotificationsMarkAllRead,
} from "@/api/notifications/notifications"
import { NotificationList } from "@/components/notifications/notification-list"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

interface NotificationBellProps {
  /** Destination for the "view all" footer link — org or admin inbox. */
  to: string
}

/** Header bell shared by the org and admin panels. Polls the unread count every
 * 30s for the badge and reveals the ten most recent notifications on open. */
export function NotificationBell({ to }: NotificationBellProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: statusData } = useGetNotificationsStatus({
    query: { refetchInterval: 30_000, refetchOnWindowFocus: true },
  })
  const status = (statusData?.status === 200 && statusData.data.data) || undefined
  const unread = status?.unread_count ?? 0

  // Only fetch the list while the popover is open — the badge count above is
  // what polls in the background.
  const { data: listData, isLoading } = useGetNotifications(
    { page: 1, page_size: 10 },
    { query: { enabled: open } }
  )
  const items = ((listData?.status === 200 && listData.data.data?.items) || []) as InboxItem[]

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetNotificationsStatusQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetNotificationsQueryKey() })
  }

  const markAllRead = usePostNotificationsMarkAllRead({
    mutation: { onSuccess: invalidate },
  })
  const markRead = usePostNotificationsIdRead({
    mutation: { onSuccess: invalidate },
  })

  const handleItemClick = (item: InboxItem) => {
    if (!item.read_at && item.id) markRead.mutate({ id: item.id })
    setOpen(false)

    const url = item.action_url
    if (!url) return
    if (/^https?:\/\//i.test(url)) {
      window.open(url, "_blank", "noopener,noreferrer")
    } else {
      navigate({ to: url })
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("notifications.title")}
            className="relative"
          >
            <BellIcon />
            {unread > 0 && (
              <span
                className={cn(
                  "bg-primary text-primary-foreground absolute end-0 top-0 flex h-4 min-w-4 -translate-y-1/4 translate-x-1/4 items-center justify-center rounded-full px-1 text-[10px] font-semibold tabular-nums shadow-sm ring-2 ring-background",
                  "rtl:translate-x-[-25%]"
                )}
              >
                {unread > 9 ? "9+" : unread}
              </span>
            )}
          </Button>
        }
      />
      <PopoverContent align="end" className="w-80 gap-0 p-0" sideOffset={8}>
        <header className="flex items-center justify-between gap-2 border-b px-3 py-2.5">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold tracking-tight">
              {t("notifications.title")}
            </span>
            {unread > 0 && (
              <span className="bg-primary/10 text-primary rounded-full px-1.5 py-0.5 text-[10px] font-semibold tabular-nums">
                {unread}
              </span>
            )}
          </div>
          <Button
            variant="ghost"
            size="xs"
            disabled={unread === 0 || markAllRead.isPending}
            onClick={() => markAllRead.mutate()}
          >
            <CheckCheckIcon />
            {t("notifications.markAllRead")}
          </Button>
        </header>

        <div className="max-h-96 overflow-y-auto p-1.5">
          {isLoading ? (
            <div className="flex justify-center py-10">
              <Spinner />
            </div>
          ) : items.length === 0 ? (
            <div className="flex flex-col items-center gap-2 px-4 py-12 text-center">
              <BellIcon className="text-muted-foreground/60 size-6" />
              <p className="text-muted-foreground text-sm">{t("notifications.empty")}</p>
            </div>
          ) : (
            <NotificationList items={items} limit={10} onItemClick={handleItemClick} />
          )}
        </div>

        <footer className="border-t p-1.5">
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-center"
            render={<Link to={to} onClick={() => setOpen(false)} />}
          >
            {t("notifications.viewAll")}
          </Button>
        </footer>
      </PopoverContent>
    </Popover>
  )
}

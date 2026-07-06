import type { GithubCom4H1RZooraInternalDomainNotificationInboxItem as InboxItem } from "@/api/model"

import { useTranslation } from "react-i18next"

import { GithubCom4H1RZooraInternalDomainNotificationCategory as Category } from "@/api/model"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { formatRelativeTime } from "@/lib/relative-time"

// Category → i18n label key. Keeps the badge copy in one place so the dropdown
// and the full inbox never diverge.
const CATEGORY_KEY: Record<string, string> = {
  [Category.NotificationCategorySystem]: "notifications.category.system",
  [Category.NotificationCategoryOrg]: "notifications.category.org",
  [Category.NotificationCategoryClass]: "notifications.category.class",
  [Category.NotificationCategoryReminder]: "notifications.category.reminder",
}

interface NotificationListProps {
  items: InboxItem[]
  /** When set, only the first `limit` rows render (dropdown preview). */
  limit?: number
  onItemClick?: (item: InboxItem) => void
  className?: string
}

/** Shared notification rows used by both the header dropdown and the full inbox
 * page. Each row surfaces an unread accent, title, body preview, category badge
 * and a localized relative time. */
export function NotificationList({ items, limit, onItemClick, className }: NotificationListProps) {
  const { t, i18n } = useTranslation()
  const rows = typeof limit === "number" ? items.slice(0, limit) : items

  return (
    <ul className={cn("flex flex-col", className)}>
      {rows.map((item) => {
        const unread = !item.read_at
        const categoryKey = item.category ? CATEGORY_KEY[item.category] : undefined

        return (
          <li key={item.id}>
            <button
              type="button"
              onClick={() => onItemClick?.(item)}
              className={cn(
                "group relative flex w-full gap-3 rounded-lg px-3 py-3 text-start transition-colors",
                "hover:bg-muted/60 focus-visible:bg-muted/60 focus-visible:outline-hidden",
                unread && "bg-primary/[0.04]"
              )}
            >
              {/* Unread accent — a filled dot; a hollow ring keeps read rows aligned. */}
              <span className="mt-1.5 flex w-2 shrink-0 justify-center" aria-hidden>
                {unread ? (
                  <span className="bg-primary size-2 rounded-full shadow-[0_0_0_3px_color-mix(in_oklch,var(--primary)_18%,transparent)]" />
                ) : (
                  <span className="border-border size-2 rounded-full border" />
                )}
              </span>

              <div className="flex min-w-0 flex-1 flex-col gap-1">
                <div className="flex items-start justify-between gap-2">
                  <p
                    className={cn(
                      "truncate text-sm leading-tight",
                      unread ? "text-foreground font-semibold" : "text-foreground/90 font-medium"
                    )}
                  >
                    {item.title}
                  </p>
                  <time className="text-muted-foreground shrink-0 font-mono text-[11px] tracking-tight tabular-nums">
                    {formatRelativeTime(item.created_at, i18n.language)}
                  </time>
                </div>

                {item.body && (
                  <p className="text-muted-foreground line-clamp-2 text-xs leading-relaxed">
                    {item.body}
                  </p>
                )}

                {categoryKey && (
                  <div className="mt-0.5">
                    <Badge variant="secondary" className="h-4 px-1.5 text-[10px] font-medium">
                      {t(categoryKey)}
                    </Badge>
                  </div>
                )}
              </div>
            </button>
          </li>
        )
      })}
    </ul>
  )
}

import { Badge } from "@/components/ui/badge"

import { useTotalUnread } from "./use-total-unread"

/**
 * Trailing badge for the "Conversations" sidebar nav item. Live app-wide —
 * `useTotalUnread` keeps the conversations query mounted regardless of the
 * current route — and renders nothing when there's nothing unread (or chat
 * isn't entitled, in which case the count is always `0`).
 */
export function ConversationsNavBadge() {
  const unread = useTotalUnread()
  if (unread === 0) return null

  return (
    <Badge className="h-5 min-w-5 shrink-0 justify-center rounded-full px-1.5 text-[11px] font-semibold tabular-nums">
      {unread > 99 ? "99+" : unread}
    </Badge>
  )
}

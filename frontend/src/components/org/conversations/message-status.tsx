import type { ChatMessage } from "./lib/messages"

import { CheckCheckIcon, CheckIcon, ClockIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetConversationsIdMembers } from "@/api/conversations/conversations"
import { cn } from "@/lib/utils"

import { countReaders, isReadByAll } from "./lib/read-receipts"
import { useReadState } from "./use-read-state"

interface MessageStatusProps {
  convId: string
  messageId: string
  /** direct | group | channel — channels never advance past the single tick. */
  conversationType?: string
  /** Optimistic status of the bubble; `"sending"` renders the pending clock. */
  status?: ChatMessage["_status"]
}

/**
 * Telegram-style delivery status on the current user's OWN message, rendered
 * inside the bubble meta row (primary surface, so colours are tuned against
 * `primary-foreground`):
 *
 * - PENDING (`status === "sending"`): a clock — not yet acknowledged by server.
 * - SENT (confirmed, not yet read by all): a single tick.
 * - READ (confirmed, every OTHER member's read cursor has reached it): a blue
 *   double tick. Channels stay at a single tick (broadcast, no all-read signal);
 *   a solo conversation with no other members also stays single.
 *
 * A group/channel bubble that at least one other member has read carries a
 * native `title` ("read by N") so the per-reader count isn't lost.
 *
 * `"failed"` is intentionally NOT handled here — the bubble renders its own
 * retry/discard row for that and gates this component out.
 */
export function MessageStatus({ convId, messageId, conversationType, status }: MessageStatusProps) {
  const { t } = useTranslation()
  const { user } = useAccess()
  const readMap = useReadState(convId)
  const { data: membersData } = useGetConversationsIdMembers(convId)

  if (status === "sending") {
    return (
      <ClockIcon
        className="text-primary-foreground/50 size-3 shrink-0"
        aria-label={t("conversations.receipts.sending")}
      />
    )
  }

  const members = membersData?.status === 200 ? (membersData.data.data ?? []) : []
  const otherIds = members
    .map((m) => m.user_id ?? m.user?.id)
    .filter((id): id is string => !!id && id !== user.id)

  // Read = every OTHER member has read. Channels skip the compute (single tick
  // only); a solo conversation (no other members) also never reaches double.
  const allRead = conversationType !== "channel" && isReadByAll(readMap, otherIds, messageId)

  const Icon = allRead ? CheckCheckIcon : CheckIcon
  const label = t(allRead ? "conversations.receipts.read" : "conversations.receipts.sent")

  const icon = (
    <Icon
      className={cn("size-3 shrink-0", allRead ? "text-sky-300" : "text-primary-foreground/60")}
      aria-label={label}
    />
  )

  // Group/channel: surface the per-reader count on hover once anyone has read.
  const readers = conversationType === "direct" ? 0 : countReaders(readMap, messageId, user.id)
  if (readers > 0) {
    return (
      <span className="inline-flex" title={t("conversations.receipts.readBy", { count: readers })}>
        {icon}
      </span>
    )
  }
  return icon
}

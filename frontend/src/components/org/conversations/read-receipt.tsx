import { CheckCheckIcon, CheckIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetConversationsIdMembers } from "@/api/conversations/conversations"
import { cn } from "@/lib/utils"

import { countReaders, isReadBy } from "./lib/read-receipts"
import { useReadState } from "./use-read-state"

interface ReadReceiptProps {
  convId: string
  messageId: string
  /** direct | group | channel — drives tick (DM) vs "read by N" (group) form. */
  conversationType?: string
  /** Whether this is the user's newest confirmed message (gates the group form). */
  isLatestOwn: boolean
}

/**
 * Read state on the current user's OWN message. Renders inside the bubble meta
 * row (on the primary surface), so colours are tuned against `primary-foreground`.
 *
 * - DIRECT: a single tick (sent) that becomes a coloured double tick once the
 *   other member's read cursor has reached this message.
 * - GROUP/CHANNEL: only the newest own message shows a subtle "read by N", and
 *   only once at least one other member has read it — avoids per-bubble clutter.
 */
export function ReadReceipt({ convId, messageId, conversationType, isLatestOwn }: ReadReceiptProps) {
  const { t } = useTranslation()
  const { user } = useAccess()
  const readMap = useReadState(convId)
  const { data: membersData } = useGetConversationsIdMembers(convId)
  const members = membersData?.status === 200 ? (membersData.data.data ?? []) : []

  if (conversationType === "direct") {
    const other = members.find((m) => (m.user_id ?? m.user?.id) !== user.id)
    const otherId = other?.user_id ?? other?.user?.id
    const read = otherId ? isReadBy(readMap[otherId], messageId) : false
    const Icon = read ? CheckCheckIcon : CheckIcon
    const label = t(read ? "conversations.receipts.read" : "conversations.receipts.sent")
    return (
      <Icon
        className={cn("size-3.5 shrink-0", read ? "text-sky-300" : "text-primary-foreground/60")}
        aria-label={label}
      />
    )
  }

  // group / channel: only the newest own message, and only once someone's read it.
  if (!isLatestOwn) return null
  const readers = countReaders(readMap, messageId, user.id)
  if (readers <= 0) return null
  return (
    <span className="text-primary-foreground/70 inline-flex items-center gap-0.5">
      <CheckCheckIcon className="size-3.5 shrink-0 text-sky-300" aria-hidden />
      {t("conversations.receipts.readBy", { count: readers })}
    </span>
  )
}

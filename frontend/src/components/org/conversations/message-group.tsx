import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { useProfileCard } from "@/stores/profile-card"
import { cn } from "@/lib/utils"

import { avatarTint, initials, nameColor } from "./lib/avatar"
import type { Group } from "./lib/messages"

interface MessageGroupProps {
  group: Group
  /** The group's author is the signed-in user. */
  isOwn: boolean
  /** Show the sender name once atop the group (group/channel, non-own only). */
  showSenderName: boolean
  /** Bubble renderer, injected so the group stays agnostic of bubble internals. */
  renderBubble: (message: Group["messages"][number]) => React.ReactNode
}

/**
 * A run of consecutive same-sender messages: the sender avatar (start side,
 * pinned to the group's last bubble) and name are shown once, then each
 * message's bubble is stacked tight beneath. Own groups sit on the start side
 * and drop the avatar/name for an iMessage-style asymmetry.
 */
export function MessageGroup({ group, isOwn, showSenderName, renderBubble }: MessageGroupProps) {
  const sender = group.messages[0]?.sender
  const senderName = sender?.name ?? ""
  const openCard = useProfileCard((s) => s.open)
  const openSender = () => sender?.id && openCard({ userId: sender.id, name: senderName })

  return (
    <div
      className={cn(
        "flex items-end gap-2 px-4 py-1",
        isOwn ? "flex-row pe-12" : "flex-row-reverse ps-12"
      )}
    >
      {/* Sender avatar (own included), pinned to the group's last bubble. No
          server-side avatar image exists yet, so this is initials-only. */}
      <button type="button" onClick={openSender} className="self-end" aria-label={senderName}>
        <Avatar className="size-8">
          <AvatarFallback className={cn("text-[11px] font-semibold", avatarTint(sender?.id))}>
            {initials(senderName)}
          </AvatarFallback>
        </Avatar>
      </button>

      <div className={cn("flex min-w-0 flex-1 flex-col gap-0.5", isOwn ? "items-start" : "items-end")}>
        {showSenderName && senderName && (
          <button
            type="button"
            onClick={openSender}
            className={cn("px-1 text-xs font-semibold hover:underline", nameColor(sender?.id))}
          >
            {senderName}
          </button>
        )}
        {group.messages.map((message) => (
          <div key={message.id} className="w-full">
            {renderBubble(message)}
          </div>
        ))}
      </div>
    </div>
  )
}

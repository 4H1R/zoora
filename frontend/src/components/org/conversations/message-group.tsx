import { Avatar, AvatarFallback } from "@/components/ui/avatar"
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
 * message's bubble is stacked tight beneath. Own groups flip to the end side
 * and drop the avatar/name for an iMessage-style asymmetry.
 */
export function MessageGroup({ group, isOwn, showSenderName, renderBubble }: MessageGroupProps) {
  const sender = group.messages[0]?.sender
  const senderName = sender?.name ?? ""

  return (
    <div
      className={cn(
        "flex items-end gap-2 px-4 py-1",
        isOwn ? "flex-row-reverse ps-12" : "flex-row pe-12"
      )}
    >
      {isOwn ? (
        // Own messages need no avatar; the end alignment carries authorship.
        <div className="w-8 shrink-0" aria-hidden />
      ) : (
        <Avatar className="size-8 self-end">
          <AvatarFallback className={cn("text-[11px] font-semibold", avatarTint(sender?.id))}>
            {initials(senderName)}
          </AvatarFallback>
        </Avatar>
      )}

      <div className={cn("flex min-w-0 flex-col gap-0.5", isOwn ? "items-end" : "items-start")}>
        {showSenderName && senderName && (
          <span className={cn("px-1 text-xs font-semibold", nameColor(sender?.id))}>
            {senderName}
          </span>
        )}
        {group.messages.map((message) => (
          <div key={message.id}>{renderBubble(message)}</div>
        ))}
      </div>
    </div>
  )
}

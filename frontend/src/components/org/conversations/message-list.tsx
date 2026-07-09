import { useState } from "react"
import { Virtuoso, type VirtuosoHandle } from "react-virtuoso"

import { Spinner } from "@/components/ui/spinner"

import { DayDivider } from "./day-divider"
import type { ChatMessage, Group } from "./lib/messages"
import { groupMessages } from "./lib/messages"
import { MessageBubble } from "./message-bubble"
import { MessageGroup } from "./message-group"

interface MessageListProps {
  messages: ChatMessage[]
  /** Signed-in user id — decides own vs. other alignment. */
  currentUserId: string
  /** Direct DMs hide sender names; groups/channels show them. */
  conversationType?: string
  hasPreviousPage: boolean
  fetchPreviousPage: () => void
  isFetchingPreviousPage: boolean
  hasNextPage: boolean
  fetchNextPage: () => void
  /**
   * Forwarded to `<Virtuoso>` so Phase 5.3 can drive `scrollToIndex` (jump to a
   * pinned / deep-linked message). Owned by the parent thread.
   */
  virtuosoRef: React.Ref<VirtuosoHandle>
}

// Extra context handed to Virtuoso's Header slot so it can render the
// load-older spinner without a second subscription.
type ListContext = { isFetchingPreviousPage: boolean }

// Top slot: a spinner while older history streams in, otherwise a small spacer
// so the first day divider breathes.
function ListHeader({ context }: { context?: ListContext }) {
  return (
    <div className="flex h-8 items-center justify-center">
      {context?.isFetchingPreviousPage && <Spinner className="text-muted-foreground size-4" />}
    </div>
  )
}

/**
 * Reverse-infinite virtualized message list. Newest messages sit at the bottom
 * (`alignToBottom` + `initialTopMostItemIndex: "LAST"`); scrolling UP hits
 * `startReached` to load OLDER history (prepended at top), scrolling DOWN hits
 * `endReached` to load NEWER (only meaningful for an around-seed). `followOutput`
 * auto-sticks to the bottom on new messages ONLY while the user is already at
 * the bottom, so reading history is never yanked.
 *
 * The render item is a GROUP (day divider or a sender run), never a single
 * message — windowing already covers perf, so bubbles are not memoized.
 */
export function MessageList({
  messages,
  currentUserId,
  conversationType,
  hasPreviousPage,
  fetchPreviousPage,
  isFetchingPreviousPage,
  hasNextPage,
  fetchNextPage,
  virtuosoRef,
}: MessageListProps) {
  const [atBottom, setAtBottom] = useState(true)
  const groups = groupMessages(messages)
  const showSenderNames = conversationType !== "direct"

  return (
    <Virtuoso<Group, ListContext>
      ref={virtuosoRef}
      data={groups}
      context={{ isFetchingPreviousPage }}
      className="flex-1"
      computeItemKey={(_index, group) => group.id}
      initialTopMostItemIndex={{ index: "LAST" }}
      alignToBottom
      followOutput={() => (atBottom ? "smooth" : false)}
      startReached={() => {
        if (hasPreviousPage) fetchPreviousPage()
      }}
      endReached={() => {
        if (hasNextPage) fetchNextPage()
      }}
      increaseViewportBy={{ top: 600, bottom: 600 }}
      atBottomThreshold={100}
      atBottomStateChange={setAtBottom}
      components={{ Header: ListHeader }}
      itemContent={(index, group) => {
        if (group.type === "day") {
          // Day dividers carry no message; borrow the timestamp of the first
          // message of the following (same-day) group for an exact date.
          const date = groups[index + 1]?.messages[0]?.created_at
          return <DayDivider date={date} />
        }
        const isOwn = group.senderId === currentUserId
        return (
          <MessageGroup
            group={group}
            isOwn={isOwn}
            showSenderName={showSenderNames && !isOwn}
            renderBubble={(message) => <MessageBubble message={message} isOwn={isOwn} />}
          />
        )
      }}
    />
  )
}

import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon, MessagesSquareIcon } from "lucide-react"
import { useRef } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import type { VirtuosoHandle } from "react-virtuoso"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { cn } from "@/lib/utils"

import { ChatThreadSkeleton } from "./chat-thread.skeleton"
import { conversationTint, initials } from "./lib/avatar"
import { MessageList } from "./message-list"
import { useConversations } from "./use-conversations"
import { useMessages } from "./use-messages"

interface ChatThreadProps {
  convId: string
  /** Deep-link / jump target — seeds the message window around this id. */
  aroundMessageId?: string
}

/**
 * The full message thread column: a header (identity + subtitle, with a slot for
 * presence/actions), the virtualized message list, and — in Phase 6 — a composer
 * mounted in the bottom slot. Owns the `virtuosoRef` so a later phase can scroll
 * to a specific message.
 */
export function ChatThread({ convId, aroundMessageId }: ChatThreadProps) {
  const { t } = useTranslation()
  const { user } = useAccess()
  const { data: conversations } = useConversations()
  const conversation = conversations?.find((c) => c.id === convId)

  const virtuosoRef = useRef<VirtuosoHandle>(null)

  const {
    messages,
    fetchNextPage,
    fetchPreviousPage,
    hasNextPage,
    hasPreviousPage,
    isFetchingPreviousPage,
    isLoading,
  } = useMessages(convId, aroundMessageId)

  const name = conversation?.name ?? convId
  const isDirect = conversation?.type === "direct"
  const memberCount = conversation?.members?.length ?? 0
  const subtitle = isDirect
    ? t("conversations.thread.direct")
    : t("conversations.thread.members", { count: memberCount })

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* Header: identity + subtitle. Presence + actions land in the end slot. */}
      <header className="flex items-center gap-3 border-b px-4 py-3">
        <Button
          variant="ghost"
          size="icon-sm"
          className="md:hidden"
          aria-label={t("common.back")}
          render={<Link to="/org/conversations" />}
        >
          <ArrowLeftIcon className="rtl:rotate-180" />
        </Button>

        <Avatar className="size-9">
          {conversation?.avatar_url && <AvatarImage src={conversation.avatar_url} alt={name} />}
          <AvatarFallback className={cn("text-xs font-semibold", conversationTint(conversation?.color_index))}>
            {initials(name)}
          </AvatarFallback>
        </Avatar>

        <div className="flex min-w-0 flex-col">
          <p className="truncate text-sm font-semibold leading-tight">{name}</p>
          <p className="text-muted-foreground truncate text-xs leading-tight">{subtitle}</p>
        </div>

        {/* Phase later: presence indicator + thread actions mount at the end. */}
        <div className="ms-auto flex items-center gap-1" />
      </header>

      {/* Message region fills the remaining height. */}
      <div className="relative flex min-h-0 flex-1 flex-col">
        {isLoading ? (
          <ChatThreadSkeleton />
        ) : messages.length === 0 ? (
          <div className="flex flex-1 items-center justify-center p-6">
            <EmptyState
              icon={MessagesSquareIcon}
              title={t("conversations.thread.empty.title")}
              description={t("conversations.thread.empty.description")}
              className="border-none bg-transparent ring-0 shadow-none"
            />
          </div>
        ) : (
          <MessageList
            messages={messages}
            currentUserId={user.id}
            conversationType={conversation?.type}
            hasPreviousPage={hasPreviousPage}
            fetchPreviousPage={fetchPreviousPage}
            isFetchingPreviousPage={isFetchingPreviousPage}
            hasNextPage={hasNextPage}
            fetchNextPage={fetchNextPage}
            virtuosoRef={virtuosoRef}
          />
        )}
      </div>

      {/* Phase 6: <MessageInput convId={convId} /> mounts here (bottom composer). */}
    </div>
  )
}

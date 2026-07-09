import { Link, useNavigate } from "@tanstack/react-router"
import { ArrowLeftIcon, MessagesSquareIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import type { VirtuosoHandle } from "react-virtuoso"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { useChatUi } from "@/stores/chat-ui"
import { cn } from "@/lib/utils"

import { ChatThreadSkeleton } from "./chat-thread.skeleton"
import { JumpToMessageProvider } from "./jump-context"
import { conversationTint, initials } from "./lib/avatar"
import { findGroupIndex, groupMessages } from "./lib/messages"
import { MessageList } from "./message-list"
import { useConversations } from "./use-conversations"
import { useMarkRead } from "./use-mark-read"
import { useMessages } from "./use-messages"

// How long a jumped-to bubble keeps its highlight ring before it fades out.
const HIGHLIGHT_MS = 1500

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
  const navigate = useNavigate()
  const { data: conversations } = useConversations()
  const conversation = conversations?.find((c) => c.id === convId)

  const virtuosoRef = useRef<VirtuosoHandle>(null)
  const scrollToMessageId = useChatUi((s) => s.scrollToMessageId)
  const requestScrollTo = useChatUi((s) => s.requestScrollTo)

  // Pinned to the bottom? Lifted from the list so the mark-read hook can gate
  // read receipts on it. Starts true so an initial-load-at-bottom marks read.
  const [atBottom, setAtBottom] = useState(true)
  // Transient jump flash target; cleared by a timeout after HIGHLIGHT_MS.
  const [highlightId, setHighlightId] = useState<string | null>(null)
  // Whether the deep-link (`?msg`) initial jump has already fired for this mount.
  const didInitialJumpRef = useRef(false)

  const {
    messages,
    fetchNextPage,
    fetchPreviousPage,
    hasNextPage,
    hasPreviousPage,
    isFetchingPreviousPage,
    isLoading,
  } = useMessages(convId, aroundMessageId)

  useMarkRead(convId, messages, atBottom)

  // Scroll the virtual list to the group holding `id` and flash the bubble. No-op
  // if the message is not currently loaded (caller decides the fallback).
  function scrollToLoaded(id: string): boolean {
    const index = findGroupIndex(groupMessages(messages), id)
    if (index < 0) return false
    virtuosoRef.current?.scrollToIndex({ index, align: "center", behavior: "smooth" })
    setHighlightId(id)
    return true
  }

  // Reply-preview / mention / pin jump entry point (wired via context for Phase
  // 6). Loaded target → smooth in-thread scroll; unloaded → set `?msg` so the
  // thread re-seeds a window around it (see the deep-link effect below).
  function jumpToMessage(id: string) {
    if (messages.some((m) => m.id === id)) {
      requestScrollTo(id)
    } else {
      navigate({ to: ".", search: (prev) => ({ ...prev, msg: id }) })
    }
  }

  // React to store-driven jump requests (`requestScrollTo`). If the target isn't
  // loaded, fall back to re-seeding via `?msg`. Always clears the request.
  useEffect(() => {
    if (!scrollToMessageId) return
    if (!scrollToLoaded(scrollToMessageId)) {
      navigate({ to: ".", search: (prev) => ({ ...prev, msg: scrollToMessageId }) })
    }
    requestScrollTo(null)
    // scrollToLoaded reads `messages`; re-run when either the request or the
    // loaded window changes so a just-arrived target still resolves.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scrollToMessageId, messages])

  // Deep-link jump: once the `?msg`-seeded window has loaded the target, center
  // and flash it exactly once per mount.
  useEffect(() => {
    if (!aroundMessageId || didInitialJumpRef.current) return
    if (scrollToLoaded(aroundMessageId)) didInitialJumpRef.current = true
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [aroundMessageId, messages])

  // Fade the highlight ring after a beat.
  useEffect(() => {
    if (!highlightId) return
    const timer = setTimeout(() => setHighlightId(null), HIGHLIGHT_MS)
    return () => clearTimeout(timer)
  }, [highlightId])

  const name = conversation?.name ?? convId
  const isDirect = conversation?.type === "direct"
  const memberCount = conversation?.members?.length ?? 0
  const subtitle = isDirect
    ? t("conversations.thread.direct")
    : t("conversations.thread.members", { count: memberCount })

  return (
    <JumpToMessageProvider value={jumpToMessage}>
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
            atBottom={atBottom}
            onAtBottomChange={setAtBottom}
            highlightId={highlightId}
          />
        )}
      </div>

      {/* Phase 6: <MessageInput convId={convId} /> mounts here (bottom composer). */}
    </div>
    </JumpToMessageProvider>
  )
}

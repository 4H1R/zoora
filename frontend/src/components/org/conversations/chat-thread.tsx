import type { VirtuosoHandle } from "react-virtuoso"

import { Link, useNavigate } from "@tanstack/react-router"
import { ArrowLeftIcon, MessagesSquareIcon, MoreVerticalIcon, Settings2Icon, UsersIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetConversationsIdMembers } from "@/api/conversations/conversations"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { EmptyState } from "@/components/ui/empty-state"
import { formatRelativeTime } from "@/lib/relative-time"
import { cn } from "@/lib/utils"
import { useChatUi } from "@/stores/chat"

import { ChatBackground } from "./chat-background"
import { ChatThreadSkeleton } from "./chat-thread.skeleton"
import { ConversationSettings } from "./conversation-settings"
import { InConversationSearch } from "./in-conversation-search"
import { JumpToMessageProvider } from "./jump-context"
import { conversationTint, initials } from "./lib/avatar"
import { conversationTitle } from "./lib/conversation-title"
import { findGroupIndex, groupMessages } from "./lib/messages"
import { MembersSheet } from "./members-sheet"
import { MessageInput } from "./message-input"
import { MessageList } from "./message-list"
import { PinnedBar } from "./pinned-bar"
import { PresenceDot } from "./presence-dot"
import { TypingIndicator } from "./typing-indicator"
import { useConversationPermissions } from "./use-conversation-permissions"
import { useConversations } from "./use-conversations"
import { useMarkRead } from "./use-mark-read"
import { useMessages } from "./use-messages"
import { usePresence } from "./use-presence"
import { useReadStateSync } from "./use-read-state"

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
  const { t, i18n } = useTranslation()
  const { user } = useAccess()
  const { canManage } = useConversationPermissions()
  const navigate = useNavigate()
  const { data: conversations } = useConversations()
  const conversation = conversations?.find((c) => c.id === convId)

  // Roster + settings slide-overs opened from the header actions menu.
  const [membersOpen, setMembersOpen] = useState(false)
  const [settingsOpen, setSettingsOpen] = useState(false)

  // Roster fetched once here for the header count + presence ids. @mention
  // rendering no longer needs members (bubbles tokenize @username directly).
  const { data: membersData } = useGetConversationsIdMembers(convId)
  const memberRows = membersData?.status === 200 ? (membersData.data.data ?? []) : []

  // Presence for everyone in the OPEN conversation (drives the header dot +
  // subtitle). Sidebar DM presence is scoped separately in `ConversationSidebar`.
  const memberUserIds = memberRows.map((m) => m.user_id ?? m.user?.id ?? "").filter(Boolean)
  const getPresence = usePresence(memberUserIds)

  // Seed + live-merge the shared read-cursor store for this thread so own bubbles
  // can show read receipts.
  useReadStateSync(convId)

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

  const name = (conversation && conversationTitle(conversation, user.id)) || convId
  const isDirect = conversation?.type === "direct"
  const memberCount = memberRows.length

  // Header presence. DM → the partner's online/last-seen; group → an online count
  // that INCLUDES the viewer's own session. Self is counted from local state (the
  // viewer is connected by definition) rather than the server snapshot, so the
  // count is right even before the snapshot echoes our own presence back — a
  // fully-online 2-person group reads "2", not "1".
  const partnerId = isDirect ? memberUserIds.find((id) => id !== user.id) : undefined
  const partnerPresence = partnerId ? getPresence(partnerId) : undefined
  const selfOnline = memberUserIds.includes(user.id) ? 1 : 0
  const othersOnline = memberUserIds.filter((id) => id !== user.id && getPresence(id)?.online).length
  const onlineCount = isDirect ? 0 : selfOnline + othersOnline

  let subtitle: string
  if (isDirect) {
    if (partnerPresence?.online) subtitle = t("conversations.presence.online")
    else if (partnerPresence?.lastSeen)
      subtitle = t("conversations.presence.lastSeen", {
        time: formatRelativeTime(partnerPresence.lastSeen, i18n.language),
      })
    else subtitle = t("conversations.thread.direct")
  } else {
    const base = t("conversations.thread.members", { count: memberCount })
    subtitle = onlineCount > 0 ? `${base} · ${t("conversations.presence.membersOnline", { count: onlineCount })}` : base
  }

  return (
    <JumpToMessageProvider value={jumpToMessage}>
      <div className="flex min-h-0 flex-1 flex-col">
        {/* Header: identity + subtitle. Presence + actions land in the end slot.
            `relative` anchors the in-thread search overlay (mobile full-width). */}
        <header className="relative flex items-center gap-3 border-b px-4 py-3">
          <Button
            variant="ghost"
            size="icon-sm"
            className="md:hidden"
            aria-label={t("common.back")}
            render={<Link to="/org/conversations" />}
          >
            <ArrowLeftIcon className="rtl:rotate-180" />
          </Button>

          <div className="relative shrink-0">
            <Avatar className="size-9">
              {conversation?.avatar_url && <AvatarImage src={conversation.avatar_url} alt={name} />}
              <AvatarFallback className={cn("text-xs font-semibold", conversationTint(conversation?.color_index))}>
                {initials(name)}
              </AvatarFallback>
            </Avatar>
            {/* Online dot for DMs once presence is known. */}
            {isDirect && partnerPresence && (
              <PresenceDot online={partnerPresence.online} className="absolute -end-0.5 -bottom-0.5" />
            )}
          </div>

          <div className="flex min-w-0 flex-col">
            <p className="truncate text-sm leading-tight font-semibold">{name}</p>
            <p className="text-muted-foreground truncate text-xs leading-tight">{subtitle}</p>
          </div>

          {/* Thread actions mount at the end; in-thread search is the first. */}
          <div className="ms-auto flex items-center gap-1">
            <InConversationSearch convId={convId} />
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button variant="ghost" size="icon-sm" aria-label={t("conversations.thread.actions")}>
                    <MoreVerticalIcon />
                  </Button>
                }
              />
              <DropdownMenuContent align="end" className="min-w-48">
                <DropdownMenuItem onClick={() => setMembersOpen(true)}>
                  <UsersIcon data-icon="inline-start" />
                  {t("conversations.thread.viewMembers")}
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setSettingsOpen(true)}>
                  <Settings2Icon data-icon="inline-start" />
                  {t("conversations.thread.settings")}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        <MembersSheet
          convId={convId}
          open={membersOpen}
          onOpenChange={setMembersOpen}
          canManage={canManage && !isDirect}
        />
        {conversation && (
          <ConversationSettings
            conversation={conversation}
            open={settingsOpen}
            onOpenChange={setSettingsOpen}
            canManage={canManage}
          />
        )}

        {/* Pinned strip: sits between the header and the list; self-hides when empty. */}
        <PinnedBar convId={convId} />

        {/* Message region fills the remaining height. `isolate` scopes the doodle
            wallpaper's negative z-index so it sits behind the list, not the card. */}
        <div className="relative isolate flex min-h-0 flex-1 flex-col">
          <ChatBackground />
          {isLoading ? (
            <ChatThreadSkeleton />
          ) : messages.length === 0 ? (
            <div className="flex flex-1 items-center justify-center p-6">
              <EmptyState
                icon={MessagesSquareIcon}
                title={t("conversations.thread.empty.title")}
                description={t("conversations.thread.empty.description")}
                className="border-none bg-transparent shadow-none ring-0"
              />
            </div>
          ) : (
            <MessageList
              messages={messages}
              convId={convId}
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

        {/* "X is typing…" strip — fixed height so it never bumps the composer. */}
        <TypingIndicator convId={convId} />

        {/* Bottom composer: auto-grow textarea, @mentions, emoji, reply strip. */}
        <MessageInput convId={convId} />
      </div>
    </JumpToMessageProvider>
  )
}

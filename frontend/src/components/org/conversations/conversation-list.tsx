import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { useParams } from "@tanstack/react-router"
import { MessagesSquareIcon, SearchIcon, SearchXIcon, SquarePenIcon } from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"

import { ConversationItem } from "./conversation-item"
import { ConversationListSkeleton } from "./conversation-list.skeleton"
import { conversationTitle } from "./lib/conversation-title"
import { isMuted, viewerMutedUntil } from "./lib/mute"
import { directPartnerId } from "./lib/presence"
import { NewConversationDialog } from "./new-conversation-dialog"
import { useConversationPermissions } from "./use-conversation-permissions"
import { useConversations } from "./use-conversations"
import { usePresence } from "./use-presence"

// Client-side title/preview filter. The list is a single small page (v1), so
// filtering locally keeps search instant; server search arrives in a later phase.
// Matching runs on the DISPLAY title (DMs are titled after the partner, not the
// empty stored name).
function filterConversations(items: Conversation[], query: string, selfId: string): Conversation[] {
  const q = query.trim().toLowerCase()
  if (!q) return items
  return items.filter(
    (c) =>
      conversationTitle(c, selfId).toLowerCase().includes(q) || c.last_message?.content?.toLowerCase().includes(q)
  )
}

/**
 * Master pane of the conversations screen: a header with a new-conversation
 * affordance, a local search field, and the scrollable conversation list. The
 * active row tracks the current `$conversationId` route param.
 */
export function ConversationSidebar() {
  const { t } = useTranslation()
  const { user } = useAccess()
  const { canManage } = useConversationPermissions()
  const { data: conversations, isLoading } = useConversations()
  const [query, setQuery] = useState("")
  const [newOpen, setNewOpen] = useState(false)

  const params = useParams({ strict: false }) as { conversationId?: string }
  const activeId = params.conversationId

  const items = conversations ?? []
  const filtered = filterConversations(items, query, user.id)
  const isSearching = query.trim().length > 0

  // One batched presence query for every DM partner in the sidebar; each direct
  // row shows an online dot resolved from it. Group/channel rows have no dot.
  const dmPartnerIds = items.map((c) => directPartnerId(c, user.id)).filter((id): id is string => Boolean(id))
  const getPresence = usePresence(dmPartnerIds)

  return (
    <div className="flex h-full min-h-0 w-full flex-col">
      <div className="flex items-center justify-between gap-2 px-4 pt-4 pb-3">
        <h2 className="text-base font-semibold tracking-tight">{t("conversations.sidebar.title")}</h2>
        {canManage && (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={t("conversations.sidebar.newConversation")}
              title={t("conversations.sidebar.newConversation")}
              onClick={() => setNewOpen(true)}
            >
              <SquarePenIcon />
            </Button>
          </div>
        )}
      </div>

      <NewConversationDialog open={newOpen} onOpenChange={setNewOpen} canManage={canManage} />

      <div className="px-3 pb-2">
        <div className="relative">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-2.5 top-1/2 size-4 -translate-y-1/2" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("conversations.sidebar.searchPlaceholder")}
            className="h-9 ps-9"
          />
        </div>
      </div>

      {isLoading ? (
        <ConversationListSkeleton />
      ) : filtered.length === 0 ? (
        <div className="flex flex-1 items-center justify-center p-4">
          <EmptyState
            icon={isSearching ? SearchXIcon : MessagesSquareIcon}
            title={t(isSearching ? "conversations.sidebar.searchEmpty.title" : "conversations.sidebar.empty.title")}
            description={
              isSearching
                ? t("conversations.sidebar.searchEmpty.description", { query: query.trim() })
                : t("conversations.sidebar.empty.description")
            }
            className={cn("border-0 bg-transparent px-4 py-8 ring-0")}
          />
        </div>
      ) : (
        <ScrollArea className="min-h-0 flex-1">
          <div className="flex flex-col gap-0.5 px-2 pt-1 pb-3">
            {filtered.map((conversation) => {
              const partnerId = directPartnerId(conversation, user.id)
              return (
                <ConversationItem
                  key={conversation.id}
                  conversation={conversation}
                  isActive={conversation.id === activeId}
                  presenceOnline={partnerId ? getPresence(partnerId)?.online : undefined}
                  muted={isMuted(viewerMutedUntil(conversation, user.id))}
                />
              )
            })}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}

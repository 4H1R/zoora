import type { ConversationCategory } from "@/stores/conversation-filter"
import type { LucideIcon } from "lucide-react"

import { useParams } from "@tanstack/react-router"
import {
  HashIcon,
  MailIcon,
  MessagesSquareIcon,
  SearchIcon,
  SearchXIcon,
  SquarePenIcon,
  UserIcon,
  UsersIcon,
} from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { EmptyState } from "@/components/ui/empty-state"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { useConversationFilter } from "@/stores/conversation-filter"

import { ConversationItem } from "./conversation-item"
import { ConversationListSkeleton } from "./conversation-list.skeleton"
import { CONVERSATION_CATEGORIES, filterByCategory, filterByQuery } from "./lib/filter"
import { isMuted, viewerMutedUntil } from "./lib/mute"
import { directPartnerId } from "./lib/presence"
import { NewConversationDialog } from "./new-conversation-dialog"
import { NewDirectDialog } from "./new-direct-dialog"
import { useConversationPermissions } from "./use-conversation-permissions"
import { useConversations } from "./use-conversations"
import { usePresence } from "./use-presence"

// Per-category tab icons, keyed by the category value from the filter store.
const CATEGORY_ICONS: Record<ConversationCategory, LucideIcon> = {
  all: MessagesSquareIcon,
  unread: MailIcon,
  direct: UserIcon,
  group: UsersIcon,
  channel: HashIcon,
}

/**
 * Master pane of the conversations screen: a header with a new-conversation
 * affordance, a local search field, a Telegram-style category tab row, and the
 * scrollable conversation list. The active row tracks the current
 * `$conversationId` route param; the category filter is persisted across visits.
 */
export function ConversationSidebar() {
  const { t } = useTranslation()
  const { user } = useAccess()
  const { canManage } = useConversationPermissions()
  const { data: conversations, isLoading } = useConversations()
  const [query, setQuery] = useState("")
  const [newOpen, setNewOpen] = useState(false)
  const [directOpen, setDirectOpen] = useState(false)
  const [newType, setNewType] = useState<"group" | "channel">("group")
  const category = useConversationFilter((s) => s.category)
  const setCategory = useConversationFilter((s) => s.setCategory)

  function openNew(type: "group" | "channel") {
    setNewType(type)
    setNewOpen(true)
  }

  const params = useParams({ strict: false }) as { conversationId?: string }
  const activeId = params.conversationId

  const items = conversations ?? []
  // Category first (tab row), then the free-text search — both are AND-combined.
  const filtered = filterByQuery(filterByCategory(items, category), query, user.id)
  const isSearching = query.trim().length > 0
  const unreadCount = items.filter((c) => (c.unread_count ?? 0) > 0).length

  // One batched presence query for every DM partner in the sidebar; each direct
  // row shows an online dot resolved from it. Group/channel rows have no dot.
  const dmPartnerIds = items.map((c) => directPartnerId(c, user.id)).filter((id): id is string => Boolean(id))
  const getPresence = usePresence(dmPartnerIds)

  return (
    <div className="flex h-full min-h-0 w-full flex-col">
      <div className="flex items-center justify-between gap-2 px-4 pt-4 pb-3">
        <h2 className="text-base font-semibold tracking-tight">{t("conversations.sidebar.title")}</h2>
        {/* Everyone may start a DM; managers additionally get group/channel via a
            menu. Non-managers' single action opens the DM picker directly. */}
        {canManage ? (
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t("conversations.sidebar.newConversation")}
                  title={t("conversations.sidebar.newConversation")}
                >
                  <SquarePenIcon />
                </Button>
              }
            />
            <DropdownMenuContent align="end" className="min-w-48">
              <DropdownMenuItem onClick={() => setDirectOpen(true)}>
                <UserIcon data-icon="inline-start" />
                {t("conversations.new.menu.direct")}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => openNew("group")}>
                <UsersIcon data-icon="inline-start" />
                {t("conversations.new.menu.group")}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => openNew("channel")}>
                <HashIcon data-icon="inline-start" />
                {t("conversations.new.menu.channel")}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t("conversations.new.menu.direct")}
            title={t("conversations.new.menu.direct")}
            onClick={() => setDirectOpen(true)}
          >
            <SquarePenIcon />
          </Button>
        )}
      </div>

      <NewConversationDialog open={newOpen} onOpenChange={setNewOpen} initialType={newType} />
      <NewDirectDialog open={directOpen} onOpenChange={setDirectOpen} />

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

      {/* Category tab row: horizontally scrollable so labels never clip on a
          narrow sidebar. Filters client-side; selection is persisted. */}
      <Tabs
        value={category}
        onValueChange={(value) => setCategory(value as ConversationCategory)}
        className="overflow-x-auto px-3 pb-2"
      >
        <TooltipProvider>
          <TabsList variant="line" className="w-full">
            {CONVERSATION_CATEGORIES.map((cat) => {
              const Icon = CATEGORY_ICONS[cat]
              const label = t(`conversations.sidebar.filter.${cat}`)
              return (
                <Tooltip key={cat}>
                  <TooltipTrigger
                    render={
                      <TabsTrigger value={cat} aria-label={label} className="relative">
                        <Icon />
                        {cat === "unread" && unreadCount > 0 && (
                          <span className="bg-primary text-primary-foreground absolute end-1 top-0.5 min-w-4 rounded-full px-1 text-[10px] leading-4 font-medium tabular-nums">
                            {unreadCount}
                          </span>
                        )}
                      </TabsTrigger>
                    }
                  />
                  <TooltipContent>{label}</TooltipContent>
                </Tooltip>
              )
            })}
          </TabsList>
        </TooltipProvider>
      </Tabs>

      {isLoading ? (
        <ConversationListSkeleton />
      ) : filtered.length === 0 ? (
        <div className="flex flex-1 items-center justify-center p-4">
          {isSearching ? (
            <EmptyState
              icon={SearchXIcon}
              title={t("conversations.sidebar.searchEmpty.title")}
              description={t("conversations.sidebar.searchEmpty.description", { query: query.trim() })}
              className={cn("border-0 bg-transparent px-4 py-8 ring-0")}
            />
          ) : category !== "all" ? (
            <EmptyState
              icon={MessagesSquareIcon}
              title={t(`conversations.sidebar.filter.empty.${category}.title`)}
              description={t(`conversations.sidebar.filter.empty.${category}.description`)}
              className={cn("border-0 bg-transparent px-4 py-8 ring-0")}
            />
          ) : (
            <EmptyState
              icon={MessagesSquareIcon}
              title={t("conversations.sidebar.empty.title")}
              description={t("conversations.sidebar.empty.description")}
              className={cn("border-0 bg-transparent px-4 py-8 ring-0")}
            />
          )}
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

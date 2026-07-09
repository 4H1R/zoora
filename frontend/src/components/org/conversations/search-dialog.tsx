import { useNavigate } from "@tanstack/react-router"
import { HashIcon, MessageSquareTextIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import type {
  GithubCom4H1RZooraInternalDomainConversation as Conversation,
  GithubCom4H1RZooraInternalDomainConversationMessage as ConversationMessage,
} from "@/api/model"
import { useGetConversationsSearch } from "@/api/conversations/conversations"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { cn } from "@/lib/utils"

import { conversationTint, initials } from "./lib/avatar"
import { filterConversationsByQuery } from "./lib/search"
import { useConversations } from "./use-conversations"

// The global endpoint enforces a 3-character minimum; mirror it client-side so
// we don't fire (and 400) on shorter queries and can show a helpful hint.
const MIN_QUERY = 3
const DEBOUNCE_MS = 300
// Keep each group tight so the dialog stays scannable.
const MAX_CONVERSATIONS = 6
const MAX_MESSAGES = 12

interface SearchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * Org-wide command palette for conversations chat. A debounced query hits the
 * global search endpoint (which returns matching MESSAGES); the "Conversations"
 * group is derived locally from the already-loaded sidebar list. Selecting a
 * conversation opens it; selecting a message opens its conversation and jumps to
 * that message via the `?msg` deep-link. cmdk owns keyboard navigation.
 */
export function SearchDialog({ open, onOpenChange }: SearchDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: conversations } = useConversations()
  const items = conversations ?? []

  const [query, setQuery] = useState("")
  const [debounced] = useDebounce(query, DEBOUNCE_MS)
  const trimmed = debounced.trim()
  const enabled = open && trimmed.length >= MIN_QUERY

  const { data, isFetching } = useGetConversationsSearch({ q: trimmed }, { query: { enabled } })
  const messages: ConversationMessage[] =
    data?.status === 200 ? (data.data.data ?? []).slice(0, MAX_MESSAGES) : []

  // Local name/preview matches → the "Conversations" group.
  const convMatches = filterConversationsByQuery(items, trimmed).slice(0, MAX_CONVERSATIONS)
  // Resolve a message's parent conversation for its label/avatar.
  const convById = new Map(items.map((c) => [c.id, c]))

  function close() {
    onOpenChange(false)
    setQuery("")
  }

  function openConversation(id: string) {
    navigate({ to: "/org/conversations/$conversationId", params: { conversationId: id }, search: {} })
    close()
  }

  function openMessage(convId: string, messageId: string) {
    navigate({
      to: "/org/conversations/$conversationId",
      params: { conversationId: convId },
      search: { msg: messageId },
    })
    close()
  }

  const showHint = trimmed.length > 0 && trimmed.length < MIN_QUERY
  const hasResults = convMatches.length > 0 || messages.length > 0

  return (
    <CommandDialog
      open={open}
      onOpenChange={(next) => (next ? onOpenChange(true) : close())}
      title={t("conversations.search.title")}
      description={t("conversations.search.placeholder")}
      className="sm:max-w-xl"
    >
      {/* Server-driven results — disable cmdk's own substring filtering. */}
      <Command shouldFilter={false}>
        <CommandInput
          value={query}
          onValueChange={setQuery}
          placeholder={t("conversations.search.placeholder")}
        />
        <CommandList className="max-h-96">
          {showHint ? (
            <p className="text-muted-foreground px-3 py-6 text-center text-sm">
              {t("conversations.search.hint", { count: MIN_QUERY })}
            </p>
          ) : enabled && !hasResults && !isFetching ? (
            <CommandEmpty>{t("conversations.search.empty", { query: trimmed })}</CommandEmpty>
          ) : null}

          {convMatches.length > 0 && (
            <CommandGroup heading={t("conversations.search.groups.conversations")}>
              {convMatches.map((c) => (
                <ConversationResult key={c.id} conversation={c} onSelect={() => openConversation(c.id ?? "")} />
              ))}
            </CommandGroup>
          )}

          {messages.length > 0 && (
            <CommandGroup heading={t("conversations.search.groups.messages")}>
              {messages.map((m) => (
                <MessageResult
                  key={m.id}
                  message={m}
                  conversation={convById.get(m.conversation_id)}
                  onSelect={() => openMessage(m.conversation_id ?? "", m.id ?? "")}
                />
              ))}
            </CommandGroup>
          )}
        </CommandList>
      </Command>
    </CommandDialog>
  )
}

const TYPE_GLYPH = { group: UsersIcon, channel: HashIcon } as const

function ConversationAvatar({ conversation, name }: { conversation?: Conversation; name: string }) {
  const TypeIcon = conversation?.type ? TYPE_GLYPH[conversation.type as keyof typeof TYPE_GLYPH] : undefined
  return (
    <div className="relative shrink-0">
      <Avatar className="size-8">
        {conversation?.avatar_url && <AvatarImage src={conversation.avatar_url} alt={name} />}
        <AvatarFallback className={cn("text-[11px] font-semibold", conversationTint(conversation?.color_index))}>
          {initials(name)}
        </AvatarFallback>
      </Avatar>
      {TypeIcon && (
        <span className="bg-background ring-background text-muted-foreground absolute -bottom-0.5 -end-0.5 flex size-3.5 items-center justify-center rounded-full ring-2">
          <TypeIcon className="size-2" />
        </span>
      )}
    </div>
  )
}

function ConversationResult({
  conversation,
  onSelect,
}: {
  conversation: Conversation
  onSelect: () => void
}) {
  const name = conversation.name ?? ""
  return (
    <CommandItem value={`conversation-${conversation.id}`} onSelect={onSelect} className="gap-3 py-2">
      <ConversationAvatar conversation={conversation} name={name} />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm font-medium">{name}</span>
        {conversation.last_message?.content && (
          <span className="text-muted-foreground truncate text-xs">{conversation.last_message.content}</span>
        )}
      </div>
    </CommandItem>
  )
}

function MessageResult({
  message,
  conversation,
  onSelect,
}: {
  message: ConversationMessage
  conversation?: Conversation
  onSelect: () => void
}) {
  const { t } = useTranslation()
  const convName = conversation?.name ?? t("conversations.search.unknownConversation")
  const senderName = message.sender?.name
  return (
    <CommandItem value={`message-${message.id}`} onSelect={onSelect} className="items-start gap-3 py-2">
      <MessageSquareTextIcon className="text-muted-foreground mt-0.5" />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm">
          {senderName && <span className="font-medium">{senderName}: </span>}
          {message.content}
        </span>
        <span className="text-muted-foreground truncate text-xs">{convName}</span>
      </div>
    </CommandItem>
  )
}

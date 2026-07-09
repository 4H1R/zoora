import type { QueryClient } from "@tanstack/react-query"

import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { type ChatMessage, reconcileOptimistic } from "./lib/messages"
import { chatKeys } from "./lib/query-keys"
import type { WsEvent } from "./lib/ws-client"

/**
 * Cache shape of the message THREAD query (built in Phase 5). It is a TanStack
 * infinite query whose pages are ASCENDING `ChatMessage[]` arrays — this is the
 * contract the thread query must produce so the reducer can splice into it.
 */
export type MessagesInfinite = {
  pages: ChatMessage[][]
  pageParams: unknown[]
}

/**
 * Append a newly-arrived message to the last (newest) page of an infinite
 * thread cache. No-op (returns `old` unchanged) when the cache is absent — the
 * thread query may not be mounted, and hydrating it here would fabricate a
 * partial, unpaginated cache. Uses `reconcileOptimistic` so an optimistic
 * bubble already in the last page is replaced in place rather than duplicated.
 */
export function appendMessageToInfinite(
  old: MessagesInfinite | undefined,
  msg: ChatMessage
): MessagesInfinite | undefined {
  if (!old || old.pages.length === 0) return old
  const lastIdx = old.pages.length - 1
  const pages = old.pages.slice()
  pages[lastIdx] = reconcileOptimistic(pages[lastIdx], msg)
  return { ...old, pages }
}

/**
 * Replace an edited message wherever it lives across the loaded pages. No-op if
 * the cache is absent or the message is not loaded (we do not append edits for
 * messages the client never fetched).
 */
export function replaceMessageInInfinite(
  old: MessagesInfinite | undefined,
  msg: ChatMessage
): MessagesInfinite | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    const idx = page.findIndex((m) => m.id === msg.id)
    if (idx === -1) return page
    changed = true
    const next = page.slice()
    next[idx] = { ...msg, _status: undefined }
    return next
  })
  return changed ? { ...old, pages } : old
}

/**
 * Remove a deleted message from the loaded pages. No-op if the cache is absent.
 */
export function removeMessageFromInfinite(
  old: MessagesInfinite | undefined,
  id: string
): MessagesInfinite | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    if (!page.some((m) => m.id === id)) return page
    changed = true
    return page.filter((m) => m.id !== id)
  })
  return changed ? { ...old, pages } : old
}

/**
 * Cache shape of the conversation LIST query (built in Phase 4): a flat,
 * newest-activity-first `Conversation[]`.
 */
export function bumpConversationInList(
  old: Conversation[] | undefined,
  args: { convId: string; message?: ChatMessage; incrementUnread: boolean }
): Conversation[] | undefined {
  if (!old) return old
  const idx = old.findIndex((c) => c.id === args.convId)
  // Conversation not in the loaded list — no-op. A brand-new conversation is
  // surfaced by the `conversation_updated`/`member_added` invalidation paths.
  if (idx === -1) return old
  const existing = old[idx]
  const updated: Conversation = {
    ...existing,
    last_message: args.message ?? existing.last_message,
    unread_count: (existing.unread_count ?? 0) + (args.incrementUnread ? 1 : 0),
  }
  const rest = old.filter((_, i) => i !== idx)
  return [updated, ...rest]
}

// ---------------------------------------------------------------------------
// Cross-source dedup: `new_message` arrives BOTH from the joined room (full
// payload) and the per-user firehose (compact payload). A bounded, insertion-
// ordered Set of seen ids drops the second sighting. Module-level so the guard
// spans every ChatProvider/event-handler instance in the tab.
// ---------------------------------------------------------------------------
const SEEN_CAP = 500
const seenMessageIds = new Set<string>()

/** Returns true the FIRST time an id is seen, false on every repeat. */
export function markMessageSeen(id: string): boolean {
  if (seenMessageIds.has(id)) return false
  seenMessageIds.add(id)
  if (seenMessageIds.size > SEEN_CAP) {
    // Set preserves insertion order; evict the oldest id.
    const oldest = seenMessageIds.values().next().value
    if (oldest !== undefined) seenMessageIds.delete(oldest)
  }
  return true
}

/** Test-only: reset the dedup Set between cases. */
export function __clearSeenMessageIds(): void {
  seenMessageIds.clear()
}

/**
 * Build the WS-event → React Query cache reducer. The returned handler is the
 * only thing that mutates chat caches from realtime events; it is otherwise
 * pure (touches nothing but `queryClient` and the module dedup Set).
 *
 * `getFocusedConvId`/`selfUserId` are read lazily so a single long-lived handler
 * always sees the current focused thread and signed-in user.
 */
export function createChatEventHandler(opts: {
  queryClient: QueryClient
  getFocusedConvId: () => string | null
  selfUserId: () => string | null
}): (e: WsEvent) => void {
  const { queryClient, getFocusedConvId, selfUserId } = opts

  return (e: WsEvent) => {
    switch (e.type) {
      case "new_message": {
        const msg = e.data as ChatMessage
        const convId = msg.conversation_id
        const id = msg.id
        if (!convId || !id) return
        // Drop the duplicate firehose/room sighting.
        if (!markMessageSeen(id)) return

        queryClient.setQueryData<MessagesInfinite>(chatKeys.messages(convId), (old) =>
          appendMessageToInfinite(old, msg)
        )

        const incrementUnread = convId !== getFocusedConvId() && msg.sender_id !== selfUserId()
        queryClient.setQueryData<Conversation[]>(chatKeys.conversations(), (old) =>
          bumpConversationInList(old, { convId, message: msg, incrementUnread })
        )
        return
      }

      case "message_updated": {
        const msg = e.data as ChatMessage
        const convId = msg.conversation_id
        if (!convId) return
        queryClient.setQueryData<MessagesInfinite>(chatKeys.messages(convId), (old) =>
          replaceMessageInInfinite(old, msg)
        )
        return
      }

      case "message_deleted": {
        const { id, conversation_id: convId } = e.data as { id?: string; conversation_id?: string }
        if (!convId || !id) return
        queryClient.setQueryData<MessagesInfinite>(chatKeys.messages(convId), (old) =>
          removeMessageFromInfinite(old, id)
        )
        return
      }

      case "reaction_added":
      case "reaction_removed": {
        // TODO(reactions): the backend payload is `{message_id, emoji, user_id,
        // counts}` WITHOUT a conversation_id, so we cannot target a single
        // conv's message cache. Invalidate every loaded thread as the v1
        // fallback; tighten once the payload carries conversation_id.
        queryClient.invalidateQueries({ queryKey: ["chat", "messages"] })
        return
      }

      case "message_read": {
        // TODO(read-receipts): read-receipt UI lands in Phase 7; no-op for now.
        return
      }

      case "member_added":
      case "member_removed": {
        const { conversation_id: convId } = e.data as { conversation_id?: string }
        if (convId) queryClient.invalidateQueries({ queryKey: chatKeys.members(convId) })
        queryClient.invalidateQueries({ queryKey: chatKeys.conversations() })
        return
      }

      case "conversation_updated": {
        queryClient.invalidateQueries({ queryKey: chatKeys.conversations() })
        return
      }

      // user_typing / presence_update are intentionally NOT owned here — Phase 7
      // subscribes to the raw event stream for typing + presence UI.
      default:
        return
    }
  }
}

import type { ChatMessage } from "./lib/messages"
import type { WsEvent } from "./lib/ws-client"
import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"
import type { QueryClient } from "@tanstack/react-query"

import { reconcileOptimistic } from "./lib/messages"
import { applyReactionCounts } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"

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
export function removeMessageFromInfinite(old: MessagesInfinite | undefined, id: string): MessagesInfinite | undefined {
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

/**
 * Compact per-user firehose payload for the sidebar (`conversation_bump`). It
 * carries just enough to render the list row's last-message preview — NOT the
 * full thread message (that arrives on the joined room as `new_message`).
 */
type ConversationBump = {
  conversation_id?: string
  id?: string
  sender_id?: string
  content?: string
  created_at?: string
}

/**
 * Build the WS-event → React Query cache reducer. The returned handler is the
 * only thing that mutates chat caches from realtime events; it is otherwise
 * pure (touches nothing but `queryClient`).
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
        // Full-payload thread message. This event ONLY arrives for the room the
        // client joined (the focused/open conversation), so it drives the thread
        // cache. Appending is idempotent by id (`reconcileOptimistic`), so a
        // repeated sighting replaces in place rather than duplicating.
        const msg = e.data as ChatMessage
        const convId = msg.conversation_id
        const id = msg.id
        if (!convId || !id) return

        queryClient.setQueryData<MessagesInfinite>(chatKeys.messages(convId), (old) =>
          appendMessageToInfinite(old, msg)
        )

        // Bump the list too. Because this is the joined/focused conv (or our own
        // send), unread is suppressed here; the NON-focused unread path is owned
        // exclusively by `conversation_bump`, so no double-count occurs.
        const incrementUnread = convId !== getFocusedConvId() && msg.sender_id !== selfUserId()
        queryClient.setQueryData<Conversation[]>(chatKeys.conversations(), (old) =>
          bumpConversationInList(old, { convId, message: msg, incrementUnread })
        )
        return
      }

      case "conversation_bump": {
        // Per-user sidebar firehose: a compact payload for EVERY conversation the
        // user belongs to (focused or not). Bump the LIST only — never the thread
        // cache (the joined room delivers the full message via `new_message`).
        const bump = e.data as ConversationBump
        const convId = bump.conversation_id
        if (!convId) return

        const incrementUnread = convId !== getFocusedConvId() && bump.sender_id !== selfUserId()
        const preview: ChatMessage | undefined = bump.id
          ? ({
              id: bump.id,
              conversation_id: convId,
              sender_id: bump.sender_id,
              content: bump.content,
              created_at: bump.created_at,
            } as ChatMessage)
          : undefined
        queryClient.setQueryData<Conversation[]>(chatKeys.conversations(), (old) =>
          bumpConversationInList(old, { convId, message: preview, incrementUnread })
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
        // The payload is `{message_id, emoji, user_id, counts}` WITHOUT a
        // conversation_id, so we cannot key straight into one thread cache.
        // Instead of a broad invalidate, walk every loaded thread cache and
        // overwrite the matching message's reactions with the authoritative
        // `counts` map — `applyReactionCounts` no-ops on the caches that don't
        // hold the message, so only the right one changes.
        const { message_id: messageId, counts } = e.data as {
          message_id?: string
          counts?: Record<string, number>
        }
        if (!messageId || !counts) return
        for (const [qKey, data] of queryClient.getQueriesData<MessagesInfinite>({
          queryKey: ["chat", "messages"],
        })) {
          const next = applyReactionCounts(data, messageId, counts)
          if (next !== data) queryClient.setQueryData(qKey, next)
        }
        return
      }

      case "message_read": {
        // Read receipts (Phase 7) are owned by `use-read-state.ts`, which
        // subscribes to the raw stream and advances the `chat-read` store scoped
        // to the OPEN conversation. Handling it here too would double-count, so
        // the central reducer deliberately no-ops it.
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

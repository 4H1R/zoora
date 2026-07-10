import type { ChatMessage } from "./messages"
import type { InfiniteData } from "@tanstack/react-query"

import { reconcileOptimistic } from "./messages"

/**
 * The message-thread cache is a TanStack infinite query whose pages are
 * ASCENDING `ChatMessage[]` arrays (page 0 = oldest/top, last page =
 * newest/bottom). These helpers mirror the WS reducer's cache manipulation in
 * `use-chat-ws.ts` (`appendMessageToInfinite` / `replaceMessageInInfinite`) so
 * optimistic inserts, WS echoes, and server responses all converge by id.
 */
type MessagesCache = InfiniteData<ChatMessage[]>

/**
 * Insert an optimistic (client-created) message into the thread cache.
 *
 * Mirrors the reducer's `appendMessageToInfinite`: reconcile onto the LAST
 * (newest) page so a re-insert with the same id (retry) replaces in place
 * rather than duplicating. UNLIKE the reducer — which no-ops when the cache is
 * absent because a WS echo must not fabricate a partial cache — an optimistic
 * SEND owns the message, so when the thread cache is missing we seed a single
 * page (`pageParams: [{}]` matches `useMessages`' latest-seed initial param).
 */
export function insertOptimistic(old: MessagesCache | undefined, msg: ChatMessage): MessagesCache {
  if (!old || old.pages.length === 0) {
    return { pages: [[msg]], pageParams: [{}] }
  }
  const lastIdx = old.pages.length - 1
  const pages = old.pages.slice()
  pages[lastIdx] = reconcileOptimistic(pages[lastIdx], msg)
  return { ...old, pages }
}

/**
 * Replace a message wherever it lives across the loaded pages, clearing
 * `_status` on the replacement (server-confirmed → no longer optimistic).
 * No-op (returns `old` unchanged) when the cache is absent or the id is not
 * loaded — the WS `new_message` echo appends it in that case. Mirrors the
 * reducer's `replaceMessageInInfinite`.
 *
 * The POST send-response omits `sender` (the backend returns the non-preloaded
 * row, `json:"sender,omitempty"`), while the WS `new_message` echo carries a
 * populated one. Both converge on the same client id, so if the POST response
 * wins the race it must NOT clobber a good sender with an empty one — keep the
 * existing sender whenever the incoming payload lacks a usable name.
 */
export function replaceMessage(old: MessagesCache | undefined, msg: ChatMessage): MessagesCache | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    const idx = page.findIndex((m) => m.id === msg.id)
    if (idx === -1) return page
    changed = true
    const next = page.slice()
    const prev = page[idx]
    const sender = msg.sender?.name?.trim() ? msg.sender : (prev.sender ?? msg.sender)
    next[idx] = { ...msg, sender, _status: undefined }
    return next
  })
  return changed ? { ...old, pages } : old
}

/**
 * Remove the message with `id` from wherever it lives across the loaded pages.
 * No-op (returns `old` unchanged) when the cache is absent or the id is not
 * loaded. Used for optimistic deletes (server-confirmed) and to drop a failed
 * optimistic bubble that never reached the server.
 */
export function removeMessage(old: MessagesCache | undefined, id: string): MessagesCache | undefined {
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
 * Overwrite the `reactions` map of the message with `messageId` wherever it
 * lives across the loaded pages. The server (and the WS `reaction_*` payload)
 * ships an authoritative `{[emoji]: count}` map, so we replace rather than
 * merge. No-op (returns `old` unchanged) when the cache is absent or the id is
 * not loaded — the WS payload lacks a conversation_id, so this is called across
 * every thread cache and must cheaply skip the ones that don't hold the message.
 */
export function applyReactionCounts(
  old: MessagesCache | undefined,
  messageId: string,
  counts: Record<string, number>
): MessagesCache | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    const idx = page.findIndex((m) => m.id === messageId)
    if (idx === -1) return page
    changed = true
    const next = page.slice()
    next[idx] = { ...next[idx], reactions: counts }
    return next
  })
  return changed ? { ...old, pages } : old
}

/**
 * Set the optimistic `_status` on the message with `id`. No-op when the cache
 * is absent or the id is not loaded. Used to flip a send to "failed" on error
 * and back to "sending" on retry.
 */
export function markStatus(
  old: MessagesCache | undefined,
  id: string,
  status: NonNullable<ChatMessage["_status"]>
): MessagesCache | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    const idx = page.findIndex((m) => m.id === id)
    if (idx === -1) return page
    changed = true
    const next = page.slice()
    next[idx] = { ...next[idx], _status: status }
    return next
  })
  return changed ? { ...old, pages } : old
}

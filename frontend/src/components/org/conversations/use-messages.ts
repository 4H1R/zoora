import { useInfiniteQuery } from "@tanstack/react-query"

import {
  getConversationsIdMessages,
  type getConversationsIdMessagesResponse,
} from "@/api/conversations/conversations"
import type { GetConversationsIdMessagesParams } from "@/api/model"

import { type ChatMessage, dedupSortMessages, nextPageParam, prevPageParam } from "./lib/messages"
import { chatKeys } from "./lib/query-keys"

// Backend default page size; also the exhaustion threshold — a page with fewer
// than LIMIT items means that end (older/newer) is fully loaded.
const LIMIT = 50

// One of before/after/around selects a keyset window; `{}` (none) = latest page.
type MessagesPageParam = Pick<GetConversationsIdMessagesParams, "before" | "after" | "around">

/**
 * Unwrap the orval union response into ASCENDING `ChatMessage[]`. The endpoint
 * returns messages newest-first (DESC) under `data.data`; we reverse so every
 * cached page is oldest-first — the ordering contract the WS reducer relies on
 * (it appends new messages to the LAST page). Throws on a non-200 so React
 * Query surfaces the error (the orval fetcher resolves error statuses).
 */
export function unwrapMessagesPage(res: getConversationsIdMessagesResponse): ChatMessage[] {
  if (res.status !== 200) {
    throw new Error(`Failed to load messages (status ${res.status})`)
  }
  const desc = res.data.data ?? []
  return desc.slice().reverse()
}

/**
 * Bidirectional infinite message thread for a conversation.
 *
 * The cache is a TanStack infinite query whose pages are ASCENDING
 * `ChatMessage[]` arrays stored in positional order (page 0 = oldest/top, last
 * page = newest/bottom). This is the exact shape the realtime reducer in
 * `use-chat-ws.ts` reads and writes (`appendMessageToInfinite` targets the last
 * page), so the two stay coherent.
 *
 * - `aroundMessageId` seeds a window centered on a message (deep-link / jump to
 *   a pinned message); omitting it seeds the latest page.
 * - Fetch newer (`fetchNextPage`) is only meaningful for an around-seed — for a
 *   latest-seed the newest page is already loaded and WS appends new messages
 *   live, so `hasNextPage` stays false.
 * - Fetch older (`fetchPreviousPage`) walks history upward until a short page.
 *
 * The final render list comes from `select`: flatten every page, then
 * `dedupSortMessages` (ASC by uuidv7 id, last-write-wins). Optimistic bubbles
 * inserted into the cache pages by later phases flow through here unchanged.
 */
export function useMessages(convId: string, aroundMessageId?: string) {
  const seededAround = Boolean(aroundMessageId)

  const query = useInfiniteQuery({
    queryKey: chatKeys.messages(convId),
    queryFn: async ({ pageParam, signal }) => {
      const res = await getConversationsIdMessages(convId, { limit: LIMIT, ...pageParam }, { signal })
      return unwrapMessagesPage(res)
    },
    initialPageParam: (aroundMessageId ? { around: aroundMessageId } : {}) as MessagesPageParam,
    getNextPageParam: (lastPage, allPages) => nextPageParam(allPages, lastPage, LIMIT, seededAround),
    getPreviousPageParam: (firstPage, allPages) => prevPageParam(allPages, firstPage, LIMIT),
    select: (data) => dedupSortMessages(data.pages.flat()),
    staleTime: Infinity,
    enabled: Boolean(convId),
  })

  return {
    messages: query.data ?? [],
    fetchNextPage: query.fetchNextPage,
    fetchPreviousPage: query.fetchPreviousPage,
    hasNextPage: query.hasNextPage,
    hasPreviousPage: query.hasPreviousPage,
    isFetchingNextPage: query.isFetchingNextPage,
    isFetchingPreviousPage: query.isFetchingPreviousPage,
    isLoading: query.isLoading,
  }
}

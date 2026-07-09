import { useQuery } from "@tanstack/react-query"

import { getConversations, type getConversationsResponse } from "@/api/conversations/conversations"
import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { chatKeys } from "./lib/query-keys"

// v1 fetches a single generous page ordered newest-activity-first. The WS reducer
// keeps this list live (bump unread / move-to-top / last_message) so we don't need
// aggressive refetching; a moderate staleTime avoids refetch storms on navigation.
const PAGE_SIZE = 50
const STALE_MS = 30_000

/**
 * Flatten the paginated list response into the flat `Conversation[]` the chat WS
 * reducer mutates in place. Throws on a non-200 so React Query surfaces the error
 * (the orval fetcher resolves error statuses instead of rejecting).
 */
export function unwrapConversations(res: getConversationsResponse): Conversation[] {
  if (res.status !== 200) {
    throw new Error(`Failed to load conversations (status ${res.status})`)
  }
  return res.data.data?.items ?? []
}

async function fetchConversations(): Promise<Conversation[]> {
  const res = await getConversations({
    order_by: "updated_at",
    order_dir: "desc",
    page: 1,
    page_size: PAGE_SIZE,
  })
  return unwrapConversations(res)
}

/**
 * Conversation list query. MUST key on `chatKeys.conversations()` and resolve to a
 * flat `Conversation[]` — that exact key + shape is the cache the realtime reducer
 * (`use-chat-ws.ts`) writes into on every incoming event.
 *
 * `enabled` (default `true`) lets app-wide consumers — e.g. the nav unread badge
 * via `useTotalUnread` — gate the request on entitlement instead of always
 * firing it for orgs without the chat feature.
 */
export function useConversations(enabled = true) {
  return useQuery({
    queryKey: chatKeys.conversations(),
    queryFn: fetchConversations,
    staleTime: STALE_MS,
    enabled,
  })
}

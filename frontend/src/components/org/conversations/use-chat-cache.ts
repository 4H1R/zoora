import { useQueryClient } from "@tanstack/react-query"

import { getGetConversationsIdMembersQueryKey } from "@/api/conversations/conversations"

import { chatKeys } from "./lib/query-keys"

/**
 * Cache-invalidation helpers shared by the conversation-management mutations
 * (create, members, mute, rename, delete). Centralizes the exact keys so every
 * mutation refreshes the same caches the UI reads:
 *  - `chatKeys.conversations()` — the sidebar list (also mutated live by the WS
 *    reducer, so we invalidate rather than hand-patch).
 *  - members — both the WS convention key (`chatKeys.members`) and the orval
 *    query key the thread + members sheet actually read from.
 */
export function useChatCache() {
  const queryClient = useQueryClient()

  return {
    invalidateConversations: () => queryClient.invalidateQueries({ queryKey: chatKeys.conversations() }),
    invalidateMembers: (convId: string) => {
      queryClient.invalidateQueries({ queryKey: chatKeys.members(convId) })
      queryClient.invalidateQueries({ queryKey: getGetConversationsIdMembersQueryKey(convId) })
    },
  }
}

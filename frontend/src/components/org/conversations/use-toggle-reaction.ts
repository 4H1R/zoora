import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"

import { usePostConversationsMessagesMessageIdReactions } from "@/api/conversations/conversations"
import { useChatReactions } from "@/stores/chat-reactions"

import { applyReactionCounts } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"

type MessagesCache = InfiniteData<ChatMessage[]>

/**
 * Toggle a reaction on a message: add the emoji if the signed-in user has not
 * reacted with it this session, otherwise remove it. The single toggle endpoint
 * (`POST .../reactions`) flips server-side; the client mirrors that.
 *
 * Optimism has two halves that revert together on error:
 *  - the cached message's `reactions` map, bumped/decremented locally via
 *    `applyReactionCounts` (guarded against negatives; the key is dropped at 0);
 *  - the session-scoped own-set flag in the `chat-reactions` store, which drives
 *    pill highlighting.
 *
 * The mutation response returns the authoritative message, so `onSuccess`
 * overwrites the map with the server's counts. The WS `reaction_added/removed`
 * echo reconciles again idempotently (it also sets absolute counts).
 */
export function useToggleReaction(convId: string) {
  const queryClient = useQueryClient()
  const mutation = usePostConversationsMessagesMessageIdReactions()
  const setReacted = useChatReactions((s) => s.setReacted)

  const key = chatKeys.messages(convId)

  function toggle(messageId: string, emoji: string) {
    // Read current state at click time without subscribing this hook to the store.
    const hadReacted = useChatReactions.getState().own[messageId]?.has(emoji) ?? false

    const cache = queryClient.getQueryData<MessagesCache>(key)
    const message = cache?.pages.flat().find((m) => m.id === messageId)
    const currentCounts: Record<string, number> = message?.reactions ?? {}

    // add → +1, remove → -1, never below zero; a count that hits 0 drops the key.
    const nextValue = Math.max(0, (currentCounts[emoji] ?? 0) + (hadReacted ? -1 : 1))
    const nextCounts: Record<string, number> = { ...currentCounts }
    if (nextValue === 0) delete nextCounts[emoji]
    else nextCounts[emoji] = nextValue

    queryClient.setQueryData<MessagesCache>(key, (old) => applyReactionCounts(old, messageId, nextCounts))
    setReacted(messageId, emoji, !hadReacted)

    mutation.mutate(
      { messageId, data: { emoji } },
      {
        onSuccess: (res) => {
          const server = res.status === 200 ? (res.data.data as ChatMessage | undefined) : undefined
          if (!server) return
          queryClient.setQueryData<MessagesCache>(key, (old) =>
            applyReactionCounts(old, messageId, server.reactions ?? {})
          )
        },
        onError: () => {
          // Revert both halves of the optimistic toggle.
          queryClient.setQueryData<MessagesCache>(key, (old) => applyReactionCounts(old, messageId, currentCounts))
          setReacted(messageId, emoji, hadReacted)
        },
      }
    )
  }

  return { toggle }
}

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useRef, useState } from "react"
import { useDebouncedCallback } from "use-debounce"

import { usePostConversationsIdRead } from "@/api/conversations/conversations"
import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { type ChatMessage, newestReadableId } from "./lib/messages"
import { chatKeys } from "./lib/query-keys"

// Wait this long after the newest readable id settles before POSTing — collapses
// a burst of incoming messages / scroll settling into a single receipt.
const READ_DEBOUNCE_MS = 500

/**
 * Track whether the browser tab/window is focused. Marking a thread read while
 * the tab is backgrounded would clear a badge the user never actually saw.
 */
function useWindowFocused(): boolean {
  const [focused, setFocused] = useState(() =>
    typeof document === "undefined" ? true : document.hasFocus()
  )
  useEffect(() => {
    const on = () => setFocused(true)
    const off = () => setFocused(false)
    window.addEventListener("focus", on)
    window.addEventListener("blur", off)
    return () => {
      window.removeEventListener("focus", on)
      window.removeEventListener("blur", off)
    }
  }, [])
  return focused
}

/**
 * Acknowledge the newest visible message as read. Fires a debounced
 * `POST /conversations/:id/read` whenever the thread is at the bottom AND the
 * tab is focused, targeting the newest server-known (non-optimistic) message id.
 *
 * Spam guard: a ref of the last id we've scheduled means each new receipt only
 * fires when a strictly newer message exists (ids are uuidv7, lexicographically
 * time-ordered). On success the conversation's `unread_count` is zeroed directly
 * in the list cache so the sidebar badge clears without a refetch.
 *
 * No-op when the list is empty or holds no non-optimistic message.
 */
export function useMarkRead(convId: string, messages: ChatMessage[], atBottom: boolean) {
  const queryClient = useQueryClient()
  const { mutate } = usePostConversationsIdRead()
  const focused = useWindowFocused()
  const lastMarkedRef = useRef<string | null>(null)

  const post = useDebouncedCallback((messageId: string) => {
    mutate(
      { id: convId, data: { message_id: messageId } },
      {
        onSuccess: () => {
          queryClient.setQueryData<Conversation[]>(chatKeys.conversations(), (old) =>
            old?.map((c) => (c.id === convId ? { ...c, unread_count: 0 } : c))
          )
        },
      }
    )
  }, READ_DEBOUNCE_MS)

  useEffect(() => {
    if (!atBottom || !focused) return
    const id = newestReadableId(messages)
    if (!id) return
    // Skip ids we've already scheduled — only advance on a strictly newer id.
    if (lastMarkedRef.current && id.localeCompare(lastMarkedRef.current) <= 0) return
    lastMarkedRef.current = id
    post(id)
  }, [atBottom, focused, messages, post])
}

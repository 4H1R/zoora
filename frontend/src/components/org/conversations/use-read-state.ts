import { useEffect } from "react"

import { useGetConversationsIdMembers } from "@/api/conversations/conversations"
import { useChatRead } from "@/stores/chat"

import { useChatWs } from "./chat-provider"
import type { WsEvent } from "./lib/ws-client"

const EMPTY: Record<string, string> = {}

/**
 * Keep the shared `chat-read` store current for the OPEN conversation:
 *  - seed each member's `last_read_message_id` (from the members query) as the
 *    initial read cursor, and
 *  - advance cursors live from `message_read` WS frames scoped to `convId`.
 *
 * Seeding is advance-only, so a re-fetch of the member list never regresses a
 * newer live cursor. Mount once per open thread; read the cursors with
 * `useReadState(convId)`.
 */
export function useReadStateSync(convId: string): void {
  const { subscribe } = useChatWs()
  const { data: membersData } = useGetConversationsIdMembers(convId)
  const seed = useChatRead((s) => s.seed)
  const applyRead = useChatRead((s) => s.applyRead)

  // Seed from server read pointers whenever the member list (re)loads.
  useEffect(() => {
    const members = membersData?.status === 200 ? (membersData.data.data ?? []) : []
    const entries: Record<string, string> = {}
    for (const m of members) {
      const uid = m.user_id ?? m.user?.id
      if (uid && m.last_read_message_id) entries[uid] = m.last_read_message_id
    }
    if (Object.keys(entries).length > 0) seed(convId, entries)
  }, [convId, membersData, seed])

  // Advance live from message_read frames for THIS conversation.
  useEffect(() => {
    return subscribe((e: WsEvent) => {
      if (e.type !== "message_read") return
      const d = e.data as { conversation_id?: string; user_id?: string; message_id?: string }
      if (d.conversation_id !== convId || !d.user_id || !d.message_id) return
      applyRead(convId, d.user_id, d.message_id)
    })
  }, [convId, subscribe, applyRead])
}

/** Read cursors (`userId -> latest read message id`) for a conversation. */
export function useReadState(convId: string): Record<string, string> {
  return useChatRead((s) => s.byConv[convId] ?? EMPTY)
}

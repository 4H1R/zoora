import type { TypingExpiryMap } from "./lib/typing"
import type { WsEvent } from "./lib/ws-client"

import { useEffect, useState } from "react"
import { useAccess } from "react-access-engine"

import { useGetConversationsIdMembers } from "@/api/conversations/conversations"

import { useChatWs } from "./chat-provider"
import { activeTypers, markTyping, pruneExpired, TYPING_PRUNE_INTERVAL_MS } from "./lib/typing"

export interface Typer {
  userId: string
  name: string
}

/**
 * Who's currently typing in `convId`, resolved to display names via the
 * conversation's member list. Subscribes to the RAW WS stream for
 * `user_typing` frames — the central cache reducer (`use-chat-ws.ts`)
 * deliberately ignores that event type, leaving it to this Phase-7 hook.
 *
 * Keeps a `{userId: expiresAt}` map: each matching frame refreshes the
 * sender's expiry `TYPING_TTL_MS` out, and a 1s interval sweeps expired
 * entries so the indicator self-clears even if the typer goes idle or closes
 * the tab without a final "stopped typing" signal (there isn't one).
 */
export function useTyping(convId: string): Typer[] {
  const { subscribe } = useChatWs()
  const { user } = useAccess()
  const { data: membersData } = useGetConversationsIdMembers(convId)
  const [map, setMap] = useState<TypingExpiryMap>({})

  // Switching threads — drop any signals carried over from the previous one.
  useEffect(() => {
    setMap({})
  }, [convId])

  useEffect(() => {
    return subscribe((e: WsEvent) => {
      if (e.type !== "user_typing") return
      const data = e.data as { conversation_id?: string; user_id?: string }
      const userId = data.user_id
      if (data.conversation_id !== convId || !userId || userId === user.id) return
      setMap((prev) => markTyping(prev, userId, Date.now()))
    })
  }, [convId, subscribe, user.id])

  useEffect(() => {
    const interval = setInterval(() => {
      setMap((prev) => pruneExpired(prev, Date.now()))
    }, TYPING_PRUNE_INTERVAL_MS)
    return () => clearInterval(interval)
  }, [])

  const members = membersData?.status === 200 ? (membersData.data.data ?? []) : []
  const ids = activeTypers(map, Date.now())

  return ids
    .map((id) => {
      const name = members.find((m) => (m.user_id ?? m.user?.id) === id)?.user?.name
      return name ? { userId: id, name } : null
    })
    .filter((t): t is Typer => t !== null)
}

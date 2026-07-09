import { useEffect, useRef, useState } from "react"

import { useGetConversationsPresence } from "@/api/conversations/conversations"

import { useChatWs } from "./chat-provider"
import type { Presence } from "./lib/presence"
import type { WsEvent } from "./lib/ws-client"

// Presence is cheap to refetch but changes constantly via WS; a moderate
// staleTime avoids refetch storms while navigation remounts consumers.
const PRESENCE_STALE_MS = 30_000

/**
 * Batch-fetch presence for `userIds` and keep it live. The REST snapshot
 * (`GET /conversations/presence?user_ids=a,b,c`) seeds each id; `presence_update`
 * WS frames for any requested id then override the snapshot in a local map (the
 * central cache reducer intentionally ignores presence — it's owned here).
 *
 * The query is skipped while `userIds` is empty. Ids are de-duped and sorted so
 * the query key is stable regardless of caller ordering. Returns a resolver:
 * `getPresence(userId) -> { online, lastSeen } | undefined` (undefined for ids we
 * have no snapshot/live entry for).
 */
export function usePresence(userIds: string[]): (userId: string) => Presence | undefined {
  const { subscribe } = useChatWs()

  const ids = Array.from(new Set(userIds.filter(Boolean))).sort()
  const userIdsParam = ids.join(",")

  const { data } = useGetConversationsPresence(
    { user_ids: userIdsParam },
    { query: { enabled: ids.length > 0, staleTime: PRESENCE_STALE_MS } }
  )

  // Live overrides layered on top of the REST snapshot.
  const [live, setLive] = useState<Record<string, Presence>>({})

  // The single subscriber reads the current id set through a ref so changing
  // `userIds` doesn't tear down / re-add the subscription every render.
  const idsRef = useRef<Set<string>>(new Set(ids))
  idsRef.current = new Set(ids)

  useEffect(() => {
    return subscribe((e: WsEvent) => {
      if (e.type !== "presence_update") return
      const d = e.data as { user_id?: string; online?: boolean; last_seen?: string }
      const userId = d.user_id
      if (!userId || !idsRef.current.has(userId)) return
      setLive((prev) => ({ ...prev, [userId]: { online: !!d.online, lastSeen: d.last_seen } }))
    })
  }, [subscribe])

  const snapshot = data?.status === 200 ? (data.data.data ?? {}) : {}

  return (userId: string): Presence | undefined => {
    const merged = live[userId]
    if (merged) return merged
    const snap = snapshot[userId]
    if (!snap) return undefined
    return { online: !!snap.online, lastSeen: snap.last_seen }
  }
}

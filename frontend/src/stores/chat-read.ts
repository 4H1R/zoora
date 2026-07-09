import { create } from "zustand"

/** `userId -> latest read message id` for one conversation. */
type ReadMap = Record<string, string>

interface ChatReadState {
  /** `convId -> { userId -> latest read message id }`. */
  byConv: Record<string, ReadMap>
  /** Seed/merge known read pointers (e.g. members' `last_read_message_id`). Advance-only. */
  seed: (convId: string, entries: ReadMap) => void
  /** Advance one member's pointer from a live `message_read` frame. Advance-only. */
  applyRead: (convId: string, userId: string, messageId: string) => void
}

// Set a member's pointer only when strictly newer (uuidv7 lexical). Returns the
// same map reference when nothing changes so subscribers don't churn.
function advance(map: ReadMap, userId: string, messageId: string): ReadMap {
  if (!userId || !messageId) return map
  const current = map[userId]
  if (current !== undefined && messageId.localeCompare(current) <= 0) return map
  return { ...map, [userId]: messageId }
}

/**
 * Shared read-cursor store. Both the presence/read sync hook (seeding + live
 * `message_read` merge) and the message-bubble receipt UI read from here, so the
 * WS event is handled in exactly one place (`use-read-state.ts`) — the central
 * cache reducer deliberately leaves `message_read` alone.
 */
export const useChatRead = create<ChatReadState>((set) => ({
  byConv: {},
  seed: (convId, entries) =>
    set((state) => {
      let map = state.byConv[convId] ?? {}
      let changed = false
      for (const [userId, messageId] of Object.entries(entries)) {
        const next = advance(map, userId, messageId)
        if (next !== map) {
          map = next
          changed = true
        }
      }
      if (!changed) return state
      return { byConv: { ...state.byConv, [convId]: map } }
    }),
  applyRead: (convId, userId, messageId) =>
    set((state) => {
      const map = state.byConv[convId] ?? {}
      const next = advance(map, userId, messageId)
      if (next === map) return state
      return { byConv: { ...state.byConv, [convId]: next } }
    }),
}))

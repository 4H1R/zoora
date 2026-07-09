import { create } from "zustand"

/**
 * Client-only tracking of which emojis the CURRENT USER has toggled on for a
 * given message THIS SESSION. The server stores reaction COUNTS only (no
 * per-user info), so "did I react?" cannot survive a reload — this Set is a
 * best-effort, session-scoped highlight for the pills the user just toggled.
 *
 * The reaction bar subscribes to `own[messageId]` to highlight its pills; the
 * toggle hook flips a single (messageId, emoji) on click and reverts it on error.
 */
interface ChatReactionsState {
  /** messageId -> emojis the signed-in user has toggled on this session. */
  own: Record<string, Set<string>>
  /** Add/remove `emoji` from the signed-in user's set for `messageId`. */
  setReacted: (messageId: string, emoji: string, reacted: boolean) => void
}

export const useChatReactions = create<ChatReactionsState>((set) => ({
  own: {},
  setReacted: (messageId, emoji, reacted) =>
    set((state) => {
      const next = new Set(state.own[messageId])
      if (reacted) next.add(emoji)
      else next.delete(emoji)
      return { own: { ...state.own, [messageId]: next } }
    }),
}))

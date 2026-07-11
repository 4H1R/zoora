import { create } from "zustand"
import { persist } from "zustand/middleware"

/**
 * Sidebar category filter. `all`/`unread` are cross-type views; the rest map 1:1
 * onto `Conversation.type`. Persisted so the last-picked filter is restored on
 * return (Telegram-style folders).
 */
export type ConversationCategory = "all" | "unread" | "direct" | "group" | "channel"

interface ConversationFilterState {
  category: ConversationCategory
  setCategory: (category: ConversationCategory) => void
}

export const useConversationFilter = create<ConversationFilterState>()(
  persist(
    (set) => ({
      category: "all",
      setCategory: (category) => set({ category }),
    }),
    { name: "conversation-filter" }
  )
)

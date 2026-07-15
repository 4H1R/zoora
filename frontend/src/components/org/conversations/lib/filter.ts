import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"
import type { ConversationCategory } from "@/stores/conversation-filter"

import { conversationTitle } from "./conversation-title"

// Ordered category list backing the sidebar tab row. `all`/`unread` are
// cross-type; the rest match `Conversation.type` exactly.
export const CONVERSATION_CATEGORIES: ConversationCategory[] = ["all", "unread", "direct", "group", "channel"]

/**
 * Category filter (tab row). `all` passes everything through, `unread` keeps
 * conversations with a positive computed unread count, and the type categories
 * match `Conversation.type`.
 */
export function filterByCategory(items: Conversation[], category: ConversationCategory): Conversation[] {
  switch (category) {
    case "all":
      return items
    case "unread":
      return items.filter((c) => (c.unread_count ?? 0) > 0)
    default:
      return items.filter((c) => c.type === category)
  }
}

/**
 * Client-side title/preview text filter. Matching runs on the DISPLAY title (DMs
 * are titled after the partner, not the empty stored name) and the last-message
 * preview. Empty query passes everything through.
 */
export function filterByQuery(items: Conversation[], query: string, selfId: string): Conversation[] {
  const q = query.trim().toLowerCase()
  if (!q) return items
  return items.filter(
    (c) => conversationTitle(c, selfId).toLowerCase().includes(q) || c.last_message?.content?.toLowerCase().includes(q)
  )
}

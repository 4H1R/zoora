import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

/**
 * Display title for a conversation row/header. Direct conversations are
 * stored NAMELESS on the backend (their identity is the member pair), so a
 * DM is titled after the OTHER member — resolved from the `members` rows the
 * list/get endpoints preload. Falls back to the raw `name` (groups/channels,
 * or a DM whose partner row is missing).
 */
export function conversationTitle(conversation: Conversation, selfId: string): string {
  if (conversation.type === "direct") {
    const partner = (conversation.members ?? []).find((m) => (m.user_id ?? m.user?.id) !== selfId)
    const name = partner?.user?.name
    if (name) return name
  }
  return conversation.name ?? ""
}

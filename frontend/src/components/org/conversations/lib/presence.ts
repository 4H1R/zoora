import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

/** Normalized presence for a single user, resolved from REST snapshot + WS live merge. */
export interface Presence {
  online: boolean
  /** ISO timestamp of the user's last activity; absent while online / unknown. */
  lastSeen?: string
}

/**
 * The OTHER member's user id in a DIRECT conversation, or undefined when the
 * conversation isn't a DM or its members aren't loaded yet. Relies on the
 * preloaded `conversation.members` (2 rows for a DM).
 */
export function directPartnerId(conversation: Conversation, selfId: string): string | undefined {
  if (conversation.type !== "direct") return undefined
  const other = (conversation.members ?? []).find((m) => (m.user_id ?? m.user?.id) !== selfId)
  return other?.user_id ?? other?.user?.id
}

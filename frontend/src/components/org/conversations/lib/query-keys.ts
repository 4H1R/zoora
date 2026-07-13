/**
 * Central React Query cache-key factory for the conversations chat feature.
 *
 * The WS-event reducer (`use-chat-ws.ts`) and every chat query (built in later
 * phases with orval's PLAIN fetcher functions — not the generated hooks) must
 * agree on these keys so the reducer can `setQueryData`/`invalidateQueries`
 * against the exact caches those queries populate. Keep this the single source
 * of truth for chat keys.
 */
export const chatKeys = {
  conversations: () => ["chat", "conversations"] as const,
  messages: (convId: string) => ["chat", "messages", convId] as const,
  messagesAround: (convId: string, msgId: string) => ["chat", "messages", convId, "around", msgId] as const,
  pins: (convId: string) => ["chat", "pins", convId] as const,
  members: (convId: string) => ["chat", "members", convId] as const,
}

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
  // Jump-to-message ("around") views get their OWN cache entry, keyed by the
  // target id, so a warm base-thread cache (with `staleTime: Infinity`) can't
  // swallow the `{around}` initial page param. The WS reducer only writes the
  // base `messages(convId)` cache, so this transient view does not receive live
  // appends — acceptable for a read-only history jump.
  messagesAround: (convId: string, msgId: string) => ["chat", "messages", convId, "around", msgId] as const,
  pins: (convId: string) => ["chat", "pins", convId] as const,
  members: (convId: string) => ["chat", "members", convId] as const,
}

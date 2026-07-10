/** userId -> ms-epoch when they'll be considered done typing. */
export type TypingExpiryMap = Record<string, number>

/** How long a single `user_typing` signal keeps a user "typing" without a refresh. */
export const TYPING_TTL_MS = 5000

/** How often the indicator sweeps the map for expired entries. */
export const TYPING_PRUNE_INTERVAL_MS = 1000

/**
 * Immutably record/refresh a typing signal for `userId`, expiring at
 * `now + TYPING_TTL_MS`. Called only on an actual incoming `user_typing` frame
 * (self-exclusion happens upstream — the caller never inserts its own user id).
 */
export function markTyping(map: TypingExpiryMap, userId: string, now: number): TypingExpiryMap {
  return { ...map, [userId]: now + TYPING_TTL_MS }
}

/**
 * Drop entries whose expiry has passed `now`. Returns the SAME reference when
 * nothing changed, so a `setState` in an interval can skip a re-render when
 * nothing actually expired.
 */
export function pruneExpired(map: TypingExpiryMap, now: number): TypingExpiryMap {
  let changed = false
  const next: TypingExpiryMap = {}
  for (const [id, expiresAt] of Object.entries(map)) {
    if (expiresAt > now) {
      next[id] = expiresAt
    } else {
      changed = true
    }
  }
  return changed ? next : map
}

/**
 * User ids still "typing" at `now`, in the order they started (insertion
 * order of `map`) — stable so the indicator's word order doesn't shuffle on
 * every render.
 */
export function activeTypers(map: TypingExpiryMap, now: number): string[] {
  return Object.entries(map)
    .filter(([, expiresAt]) => expiresAt > now)
    .map(([id]) => id)
}

// ---------------------------------------------------------------------------
// Copy selection — "X is typing…" / "X and Y are typing…" / "Several people…"
// ---------------------------------------------------------------------------

export type TypingCopy =
  | { key: "conversations.typing.one"; params: { name: string } }
  | { key: "conversations.typing.two"; params: { name1: string; name2: string } }
  | { key: "conversations.typing.many" }

/**
 * Pick the i18n key (+ interpolation params, where relevant) for the
 * currently-typing display names. `null` when no one is typing. Three-plus
 * typers collapse to a generic "several people" line rather than growing an
 * ever-longer name list.
 */
export function typingCopy(names: string[]): TypingCopy | null {
  if (names.length === 0) return null
  if (names.length === 1) return { key: "conversations.typing.one", params: { name: names[0] } }
  if (names.length === 2) return { key: "conversations.typing.two", params: { name1: names[0], name2: names[1] } }
  return { key: "conversations.typing.many" }
}

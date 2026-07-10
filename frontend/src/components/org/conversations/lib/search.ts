import type { GithubCom4H1RZooraInternalDomainConversationMessage as ConversationMessage } from "@/api/model"

/**
 * Cycle a match cursor by one step with wraparound. `dir` is +1 (next) or -1
 * (prev); the index stays inside `[0, len)`. Returns -1 when there is nothing to
 * cycle so callers can no-op. Callers seed `current` to 0 on each fresh result
 * set, so it is always a valid in-range index by the time this is called.
 */
export function nextMatchIndex(current: number, len: number, dir: 1 | -1): number {
  if (len <= 0) return -1
  return (((current + dir) % len) + len) % len
}

/**
 * A stable identity for a message list — the ordered ids joined. Effects that
 * must re-run when the *set* of matches changes (not on every render) can depend
 * on this instead of the array reference.
 */
export function matchesKey(messages: ConversationMessage[]): string {
  return messages.map((m) => m.id ?? "").join(",")
}

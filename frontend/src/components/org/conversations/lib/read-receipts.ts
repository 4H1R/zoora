/**
 * Read-receipt logic. A member's read pointer is the newest message id they have
 * acknowledged. Message ids are uuidv7 — lexicographically ordered by creation
 * time — so "has user X read message M" reduces to a string `pointer >= M`
 * comparison. These are the load-bearing bits, so they live here (pure + tested)
 * rather than inline in the receipt UI.
 */

/**
 * True when `readPointer` is at or past `messageId` (uuidv7 lexical `>=`). False
 * when the pointer is missing/empty (the member has never read anything) or when
 * `messageId` is empty.
 */
export function isReadBy(readPointer: string | undefined, messageId: string): boolean {
  if (!readPointer || !messageId) return false
  return readPointer.localeCompare(messageId) >= 0
}

/**
 * How many of the supplied read pointers have reached `messageId`, excluding
 * `excludeUserId` (the message author never counts as one of its own readers).
 * Missing / behind pointers do not count.
 */
export function countReaders(
  readPointers: Record<string, string | undefined>,
  messageId: string,
  excludeUserId: string
): number {
  let count = 0
  for (const [userId, pointer] of Object.entries(readPointers)) {
    if (userId === excludeUserId) continue
    if (isReadBy(pointer, messageId)) count++
  }
  return count
}

/**
 * The newest own, server-confirmed message id in an ASCENDING list — the only
 * message that carries the group "read by N" affordance (keeps the thread from
 * sprouting a receipt under every own bubble). Skips optimistic bubbles
 * (`_status` set) and messages without an id. Null when the user has no
 * confirmed message loaded.
 */
export function lastOwnMessageId(
  messages: { id?: string; sender_id?: string; _status?: unknown }[],
  selfUserId: string
): string | null {
  for (let i = messages.length - 1; i >= 0; i--) {
    const m = messages[i]
    if (m.sender_id === selfUserId && m._status === undefined && m.id) return m.id
  }
  return null
}

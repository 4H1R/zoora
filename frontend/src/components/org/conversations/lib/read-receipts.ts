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
 * True when EVERY id in `otherIds` has a read pointer at or past `messageId` —
 * the "read by all" signal that flips a sent bubble to the double tick. False
 * when `otherIds` is empty (a solo conversation with no other members never
 * reaches "read"), guarding against a vacuous all-true.
 */
export function isReadByAll(
  readPointers: Record<string, string | undefined>,
  otherIds: string[],
  messageId: string
): boolean {
  if (otherIds.length === 0) return false
  return otherIds.every((id) => isReadBy(readPointers[id], messageId))
}

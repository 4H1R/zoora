/**
 * Pure @mention helpers for the composer. Kept free of React so the
 * detection/insertion/resolution rules stay unit-testable in isolation — this
 * is where the token-boundary and prefix-name edge cases hide.
 */

/** A member the composer can @mention: a stable user id + a display name. */
export interface MentionCandidate {
  id: string
  name: string
}

/**
 * A rendered slice of message content: either plain text or a resolved
 * `@<DisplayName>` mention (carrying the matched member's `userId`).
 */
export interface MentionSegment {
  text: string
  isMention: boolean
  userId?: string
}

/** An active @mention query detected just before the caret. */
export interface MentionQuery {
  /** Word chars typed after the `@` (possibly empty right after typing `@`). */
  token: string
  /** Index of the `@` character within the full text (start of the span). */
  atIndex: number
}

// A mention token is an `@` at a word boundary (start of text or after
// whitespace) followed by zero-or-more word chars, anchored to the caret.
const MENTION_RE = /(^|\s)@(\w*)$/

/**
 * Detect an in-progress `@token` immediately before the caret. `textBeforeCaret`
 * is `value.slice(0, caret)`. Returns the token (chars after `@`) and the index
 * of the `@`, or null when the caret is not inside a mention token.
 */
export function detectMention(textBeforeCaret: string): MentionQuery | null {
  const match = MENTION_RE.exec(textBeforeCaret)
  if (!match) return null
  const token = match[2]
  // The `@` sits exactly `token.length + 1` chars back from the caret.
  return { token, atIndex: textBeforeCaret.length - token.length - 1 }
}

/**
 * Replace the in-progress `@token` (spanning `query.atIndex`..`caret`) with a
 * finished `@<DisplayName> ` mention. Returns the new text and the caret offset
 * that should follow it (just past the trailing space).
 */
export function insertMention(
  value: string,
  query: MentionQuery,
  caret: number,
  name: string
): { value: string; caret: number } {
  const insert = `@${name} `
  return {
    value: value.slice(0, query.atIndex) + insert + value.slice(caret),
    caret: query.atIndex + insert.length,
  }
}

/** Insert `text` at `caret`, returning the new value and caret past the insert. */
export function insertAtCaret(value: string, caret: number, text: string): { value: string; caret: number } {
  return {
    value: value.slice(0, caret) + text + value.slice(caret),
    caret: caret + text.length,
  }
}

/**
 * Re-derive the mention user-ids that survive in the final content. A candidate
 * counts only if its literal `@<DisplayName>` substring is still present, so
 * deleting the inserted text drops the mention. Candidates are matched
 * LONGEST-NAME-FIRST and each match is blanked out before the next scan, so a
 * name that is a prefix of another (e.g. `@Ali` vs `@Ali Alizadeh`) never
 * double-counts against the same span. Ids are returned once each, in match
 * order.
 */
export function resolveMentions(content: string, members: MentionCandidate[]): string[] {
  const sorted = members.filter((m) => m.id && m.name).sort((a, b) => b.name.length - a.name.length)

  let remaining = content
  const ids: string[] = []
  for (const member of sorted) {
    const needle = `@${member.name}`
    const idx = remaining.indexOf(needle)
    if (idx === -1) continue
    if (!ids.includes(member.id)) ids.push(member.id)
    // Blank the matched span so a shorter prefix name can't re-match it.
    remaining = remaining.slice(0, idx) + " ".repeat(needle.length) + remaining.slice(idx + needle.length)
  }
  return ids
}

/**
 * Split rendered `content` into plain + mention segments for display. Best-effort
 * companion to `resolveMentions`: every `@<DisplayName>` occurrence is claimed
 * LONGEST-NAME-FIRST so a prefix name (`@Ali`) never steals a span that belongs
 * to a longer one (`@Ali Alizadeh`); a claimed range blocks any overlapping
 * shorter match. ALL non-overlapping occurrences of a name are highlighted (not
 * just the first), and plain text between claims is preserved verbatim so the
 * caller can render it with `whitespace-pre-wrap`. Returns a single plain
 * segment when nothing matches; an empty array only for empty content.
 */
export function highlightMentions(content: string, members: MentionCandidate[]): MentionSegment[] {
  const sorted = members.filter((m) => m.id && m.name).sort((a, b) => b.name.length - a.name.length)

  // Claimed [start, end) ranges, each owned by the member that matched it.
  const claims: Array<{ start: number; end: number; userId: string }> = []
  const overlaps = (start: number, end: number) => claims.some((c) => start < c.end && c.start < end)

  for (const member of sorted) {
    const needle = `@${member.name}`
    let from = 0
    for (;;) {
      const idx = content.indexOf(needle, from)
      if (idx === -1) break
      const end = idx + needle.length
      if (!overlaps(idx, end)) claims.push({ start: idx, end, userId: member.id })
      from = idx + 1
    }
  }

  claims.sort((a, b) => a.start - b.start)

  const segments: MentionSegment[] = []
  let cursor = 0
  for (const claim of claims) {
    if (claim.start > cursor) {
      segments.push({ text: content.slice(cursor, claim.start), isMention: false })
    }
    segments.push({ text: content.slice(claim.start, claim.end), isMention: true, userId: claim.userId })
    cursor = claim.end
  }
  if (cursor < content.length) {
    segments.push({ text: content.slice(cursor), isMention: false })
  }
  return segments
}

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
  const sorted = members
    .filter((m) => m.id && m.name)
    .sort((a, b) => b.name.length - a.name.length)

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

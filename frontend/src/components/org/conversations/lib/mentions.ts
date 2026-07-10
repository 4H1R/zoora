/**
 * Pure @mention helpers. Mentions are now real `@username` tokens (Telegram
 * style): the composer inserts `@username`, the renderer highlights any
 * `@username` and resolution is by exact username. Kept React-free for tests.
 */

/** A member the composer can @mention: id + display name + username. */
export interface MentionCandidate {
  id: string
  name: string
  username: string
}

/** A rendered slice of message content: plain text or an `@username` mention. */
export interface MentionSegment {
  text: string
  isMention: boolean
  username?: string
}

/** An active @mention query detected just before the caret. */
export interface MentionQuery {
  /** Username chars typed after `@` (possibly empty right after typing `@`). */
  token: string
  /** Index of the `@` within the full text. */
  atIndex: number
}

// Username charset must mirror the backend rule `^[a-z0-9_.]{3,30}$`.
const USERNAME_CHARS = "a-z0-9_."
// In-progress token before the caret: `@` at a word boundary + username chars.
const MENTION_RE = new RegExp(`(^|\\s)@([${USERNAME_CHARS}]*)$`)
// A finished, resolvable mention token anywhere in content (min length 3).
const MENTION_TOKEN_RE = new RegExp(`@([${USERNAME_CHARS}]{3,30})`, "g")

/** Detect an in-progress `@token` immediately before the caret. */
export function detectMention(textBeforeCaret: string): MentionQuery | null {
  const match = MENTION_RE.exec(textBeforeCaret)
  if (!match) return null
  const token = match[2]
  return { token, atIndex: textBeforeCaret.length - token.length - 1 }
}

/** Replace the in-progress `@token` with a finished `@username ` mention. */
export function insertMention(
  value: string,
  query: MentionQuery,
  caret: number,
  username: string
): { value: string; caret: number } {
  const insert = `@${username} `
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
 * Map `@username` tokens present in `content` to member ids (for notify).
 * Non-member and too-short tokens are ignored. Ids returned once each, in
 * first-appearance order.
 */
export function resolveMentions(content: string, members: MentionCandidate[]): string[] {
  const byUsername = new Map(members.filter((m) => m.id && m.username).map((m) => [m.username, m.id]))
  const ids: string[] = []
  for (const match of content.matchAll(MENTION_TOKEN_RE)) {
    const id = byUsername.get(match[1])
    if (id && !ids.includes(id)) ids.push(id)
  }
  return ids
}

/**
 * Split rendered `content` into plain + mention segments. Members-agnostic:
 * every `@username` (3-30 charset) becomes a clickable mention segment; whether
 * it resolves to a real user is decided lazily on click. Whitespace preserved.
 */
export function splitMentions(content: string): MentionSegment[] {
  const segments: MentionSegment[] = []
  let cursor = 0
  for (const match of content.matchAll(MENTION_TOKEN_RE)) {
    const start = match.index ?? 0
    if (start > cursor) segments.push({ text: content.slice(cursor, start), isMention: false })
    segments.push({ text: match[0], isMention: true, username: match[1] })
    cursor = start + match[0].length
  }
  if (cursor < content.length) segments.push({ text: content.slice(cursor), isMention: false })
  return segments
}

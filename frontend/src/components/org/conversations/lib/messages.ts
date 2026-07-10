import type { GithubCom4H1RZooraInternalDomainConversationMessage } from "@/api/model"

/**
 * Client-only, pre-confirmation view of a single attachment on an optimistic
 * bubble. It drives the Telegram-style render (blob preview → progress ring →
 * real media) BEFORE the server ever assigns a `media_ids` entry.
 *
 * - `localId` is a client uuid, stable for the lifetime of the optimistic bubble
 *   (progress/status updates and cancel/retry all key off it).
 * - `blobUrl` is an `URL.createObjectURL` handle for inline-previewable media
 *   (images, audio, video); it is revoked once the confirmed message swaps in
 *   (see `use-send-attachments`).
 * - `mediaId` is filled in once the file's upload resolves.
 */
export interface LocalAttachment {
  localId: string
  name: string
  contentType: string
  size: number
  blobUrl?: string
  blurhash?: string | null
  width?: number
  height?: number
  /** 0..1 upload fraction. */
  progress: number
  status: "uploading" | "done" | "error"
  mediaId?: string
}

/**
 * A conversation message enriched with client-only optimistic fields. The
 * `_status` field only exists for locally-created bubbles that have not yet
 * been confirmed by the server; server-reconciled messages clear it.
 * `_attachments` carries the pre-confirmation attachment previews so a bubble
 * can render its images/files before `media_ids` exist.
 */
export type ChatMessage = GithubCom4H1RZooraInternalDomainConversationMessage & {
  _status?: "sending" | "failed"
  _attachments?: LocalAttachment[]
}

/**
 * Backend stores conversation `media_ids` as uuid strings (and the OpenAPI
 * model now types them `string[]`). Keep a defensive normalization anyway —
 * WS payloads carry the raw JSON column, so this is the single choke point
 * whatever the runtime shape.
 */
export function mediaIdStrings(message: ChatMessage): string[] {
  const ids = message.media_ids as Array<string | number> | undefined
  if (!ids) return []
  return ids.map((id) => String(id)).filter((id) => id.length > 0)
}

export type Group = {
  id: string
  type: "day" | "messages"
  senderId?: string
  messages: ChatMessage[]
}

// Messages within this window from the previous message (same sender) stay in
// the same visual group; a larger gap starts a fresh group.
const GROUP_GAP_MS = 5 * 60 * 1000

/**
 * Dedup messages by `id` (last occurrence wins) and return them sorted
 * ascending by `id`. Ids are uuidv7 which sort lexicographically by time.
 */
export function dedupSortMessages(msgs: ChatMessage[]): ChatMessage[] {
  const byId = new Map<string, ChatMessage>()
  for (const m of msgs) {
    byId.set(m.id ?? "", m)
  }
  return Array.from(byId.values()).sort((a, b) => (a.id ?? "").localeCompare(b.id ?? ""))
}

/**
 * Derive pagination cursors from an ASCENDING list: `before` is the oldest
 * (first) id, `after` is the newest (last) id. Both null when empty.
 */
export function deriveCursors(msgs: ChatMessage[]): { before: string | null; after: string | null } {
  if (msgs.length === 0) return { before: null, after: null }
  return {
    before: msgs[0].id ?? null,
    after: msgs[msgs.length - 1].id ?? null,
  }
}

/**
 * Decide the pageParam for the NEXT (newer, appended-at-bottom) page of the
 * bidirectional message thread. Pure so the direction/exhaustion logic is
 * testable in isolation — this is where off-by-one/direction bugs hide.
 *
 * - Latest-seed (`seededAround === false`): the bottom page is already the
 *   newest page and realtime WS events append new messages live, so there is
 *   never a "newer" page to fetch → always `undefined`.
 * - Around-seed: newer messages may exist beyond the window. Allow an `after`
 *   fetch while the newest-position page comes back FULL; a short page means the
 *   newer end is exhausted.
 *
 * `lastPage` is the newest-position (bottom) page; the cursor is the overall
 * newest id across every loaded page (robust to per-page storage order).
 */
export function nextPageParam(
  allPages: ChatMessage[][],
  lastPage: ChatMessage[],
  limit: number,
  seededAround: boolean
): { after: string } | undefined {
  if (!seededAround) return undefined
  if (lastPage.length < limit) return undefined
  const { after } = deriveCursors(dedupSortMessages(allPages.flat()))
  return after ? { after } : undefined
}

/**
 * Decide the pageParam for the PREVIOUS (older, prepended-at-top) page. Allow a
 * `before` fetch while the oldest-position page comes back FULL; a short (or
 * empty) first page means the older end is exhausted → `undefined`.
 *
 * `firstPage` is the oldest-position (top) page; the cursor is the overall
 * oldest id across every loaded page.
 */
export function prevPageParam(
  allPages: ChatMessage[][],
  firstPage: ChatMessage[],
  limit: number
): { before: string } | undefined {
  if (firstPage.length < limit) return undefined
  const { before } = deriveCursors(dedupSortMessages(allPages.flat()))
  return before ? { before } : undefined
}

function isoDate(created?: string): string {
  if (!created) return "unknown"
  // Local calendar day, YYYY-MM-DD.
  const d = new Date(created)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

function timeMs(created?: string): number {
  return created ? new Date(created).getTime() : 0
}

/**
 * Walk an ASCENDING message list, producing render groups. A `day` divider is
 * emitted whenever the calendar day changes; consecutive same-sender messages
 * within GROUP_GAP_MS collapse into a single `messages` group.
 */
export function groupMessages(msgs: ChatMessage[]): Group[] {
  const groups: Group[] = []
  let current: Group | null = null
  let currentDay: string | null = null
  let lastTime = 0

  for (const m of msgs) {
    const day = isoDate(m.created_at)
    const t = timeMs(m.created_at)

    if (day !== currentDay) {
      groups.push({ id: `day-${day}`, type: "day", messages: [] })
      currentDay = day
      current = null
    }

    const sameSender = current && current.senderId === m.sender_id
    const withinGap = current && t - lastTime <= GROUP_GAP_MS
    if (!current || !sameSender || !withinGap) {
      current = { id: m.id ?? "", type: "messages", senderId: m.sender_id, messages: [] }
      groups.push(current)
    }

    current.messages.push(m)
    lastTime = t
  }

  return groups
}

/**
 * Index of the render GROUP that contains `messageId`, or -1 if no loaded group
 * holds it. The index is in the SAME space `<Virtuoso>` renders (day dividers
 * included), so it can be handed straight to `scrollToIndex`. Day dividers carry
 * no messages, so they never match — only `messages` groups are searched.
 */
export function findGroupIndex(groups: Group[], messageId: string): number {
  return groups.findIndex((g) => g.type === "messages" && g.messages.some((m) => m.id === messageId))
}

/**
 * Newest message id that is safe to acknowledge as read: the last (newest, since
 * the list is ASCENDING) message that is NOT an unconfirmed optimistic bubble
 * (`_status` set) and carries a real id. Returns null when the list is empty or
 * every message is still optimistic — nothing server-known to mark read.
 */
export function newestReadableId(messages: ChatMessage[]): string | null {
  for (let i = messages.length - 1; i >= 0; i--) {
    const m = messages[i]
    if (m._status === undefined && m.id) return m.id
  }
  return null
}

/**
 * Merge an incoming (server-confirmed) message into an existing ASCENDING list.
 * If the id already exists, replace it in place and clear `_status`; otherwise
 * append. Always returns a new array.
 */
export function reconcileOptimistic(existing: ChatMessage[], incoming: ChatMessage): ChatMessage[] {
  const idx = existing.findIndex((m) => m.id === incoming.id)
  const merged: ChatMessage = { ...incoming, _status: undefined }
  if (idx === -1) {
    return [...existing, merged]
  }
  const out = existing.slice()
  out[idx] = merged
  return out
}

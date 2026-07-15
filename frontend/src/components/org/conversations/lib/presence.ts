import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

/** Normalized presence for a single user, resolved from REST snapshot + WS live merge. */
export interface Presence {
  online: boolean
  /** ISO timestamp of the user's last activity; absent while online / unknown. */
  lastSeen?: string
}

/**
 * Pick whichever of `live` (an accumulated WS `presence_update` delta) or
 * `snapshot` (an entry from the REST `GetConversationsPresence` result) is
 * fresher, comparing `lastSeen` timestamps. A live entry only "wins" while
 * it's actually newer than (or as new as) the snapshot — once a REST
 * refetch produces a newer snapshot (e.g. after a WS gap on reconnect), the
 * fresh snapshot supersedes a stale live entry instead of being shadowed
 * forever.
 *
 * When only one of `live`/`snapshot` is present, that one is returned; when
 * neither is present, returns undefined. When a `lastSeen` is missing or
 * unparseable, that side is treated as not comparable: it loses to the
 * other side if that side has a parseable timestamp (concrete data beats
 * unknown), and `live` wins if neither timestamp is comparable (including a
 * true tie).
 */
export function pickFreshestStatus(live: Presence | undefined, snapshot: Presence | undefined): Presence | undefined {
  if (!live) return snapshot
  if (!snapshot) return live

  const liveTime = parseTimestamp(live.lastSeen)
  const snapshotTime = parseTimestamp(snapshot.lastSeen)

  if (snapshotTime !== undefined && (liveTime === undefined || snapshotTime > liveTime)) {
    return snapshot
  }
  return live
}

function parseTimestamp(value: string | undefined): number | undefined {
  if (!value) return undefined
  const ms = Date.parse(value)
  return Number.isNaN(ms) ? undefined : ms
}

/**
 * The OTHER member's user id in a DIRECT conversation, or undefined when the
 * conversation isn't a DM or its members aren't loaded yet. Relies on the
 * preloaded `conversation.members` (2 rows for a DM).
 */
export function directPartnerId(conversation: Conversation, selfId: string): string | undefined {
  if (conversation.type !== "direct") return undefined
  const other = (conversation.members ?? []).find((m) => (m.user_id ?? m.user?.id) !== selfId)
  return other?.user_id ?? other?.user?.id
}

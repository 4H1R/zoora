import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

// Mute-duration presets offered in the conversation settings. A `null`
// muted_until clears the mute (unmute); "forever" uses a far-future timestamp so
// the backend keeps the conversation muted until the user explicitly turns it
// back on (the API stores an absolute `muted_until`, so there is no sentinel).
export const MUTE_FOREVER_ISO = "2999-12-31T23:59:59.000Z"

export type MuteDuration = "1h" | "8h" | "1w" | "forever"

const DURATION_MS: Record<Exclude<MuteDuration, "forever">, number> = {
  "1h": 60 * 60 * 1000,
  "8h": 8 * 60 * 60 * 1000,
  "1w": 7 * 24 * 60 * 60 * 1000,
}

/** RFC3339 timestamp to mute until, for a given preset. */
export function muteUntilISO(duration: MuteDuration): string {
  if (duration === "forever") return MUTE_FOREVER_ISO
  return new Date(Date.now() + DURATION_MS[duration]).toISOString()
}

/** True while `muted_until` is a valid timestamp still in the future. */
export function isMuted(mutedUntil?: string | null): boolean {
  if (!mutedUntil) return false
  const t = new Date(mutedUntil).getTime()
  return Number.isFinite(t) && t > Date.now()
}

/**
 * The viewer's `muted_until` for a conversation, read from its preloaded members
 * (when available). Returns undefined when the member row isn't present — the
 * list endpoint doesn't always preload members, so callers treat that as "not
 * muted" rather than an error.
 */
export function viewerMutedUntil(conversation: Conversation, userId: string): string | undefined {
  const member = (conversation.members ?? []).find((m) => (m.user_id ?? m.user?.id) === userId)
  return member?.muted_until
}

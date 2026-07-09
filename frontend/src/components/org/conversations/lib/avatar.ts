// Message sender avatars have no server-assigned `color_index` (unlike
// conversations), so we derive a stable palette slot from the sender id. A given
// user therefore keeps the same accent across every message. Tailwind scale
// only — no hashing into arbitrary hex.

const AVATAR_TINTS = [
  "bg-rose-100 text-rose-700 dark:bg-rose-500/20 dark:text-rose-200",
  "bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-200",
  "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-200",
  "bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200",
  "bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-200",
  "bg-cyan-100 text-cyan-700 dark:bg-cyan-500/20 dark:text-cyan-200",
  "bg-pink-100 text-pink-700 dark:bg-pink-500/20 dark:text-pink-200",
  "bg-indigo-100 text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-200",
] as const

const NAME_TINTS = [
  "text-rose-600 dark:text-rose-300",
  "text-sky-600 dark:text-sky-300",
  "text-emerald-600 dark:text-emerald-300",
  "text-amber-600 dark:text-amber-300",
  "text-violet-600 dark:text-violet-300",
  "text-cyan-600 dark:text-cyan-300",
  "text-pink-600 dark:text-pink-300",
  "text-indigo-600 dark:text-indigo-300",
] as const

// Stable non-negative palette slot from a key (sender id). Simple char sum is
// enough to spread users across the fixed palette.
function slotFor(key: string, len: number): number {
  let sum = 0
  for (let i = 0; i < key.length; i++) sum += key.charCodeAt(i)
  return sum % len
}

export function avatarTint(key: string | undefined): string {
  return AVATAR_TINTS[slotFor(key ?? "", AVATAR_TINTS.length)]
}

// Conversations carry a server-assigned `color_index`; map it straight onto the
// palette (matching the sidebar) so a conversation keeps one accent everywhere.
export function conversationTint(index: number | undefined): string {
  const len = AVATAR_TINTS.length
  return AVATAR_TINTS[(((index ?? 0) % len) + len) % len]
}

export function nameColor(key: string | undefined): string {
  return NAME_TINTS[slotFor(key ?? "", NAME_TINTS.length)]
}

// Up to two initials (first + last token) for the avatar fallback. Works for
// Persian and Latin names alike; toUpperCase is a no-op on scripts without case.
export function initials(name?: string): string {
  const parts = (name ?? "").trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return "؟"
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

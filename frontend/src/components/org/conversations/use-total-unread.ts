import { FEATURE, useHasFeature } from "@/lib/entitlements"

import { useConversations } from "./use-conversations"

/**
 * App-wide summed unread count across every conversation, driven by the same
 * `chatKeys.conversations()` cache the WS reducer (`use-chat-ws.ts`) keeps live.
 * Mounting this hook is what keeps that query running even off the chat page —
 * e.g. for the sidebar nav badge — but only for orgs entitled to chat; other
 * orgs never issue the request and always read back `0`.
 */
export function useTotalUnread(): number {
  const { enabled: chatEnabled } = useHasFeature(FEATURE.chat)
  const { data: conversations } = useConversations(chatEnabled)

  if (!chatEnabled || !conversations) return 0
  return conversations.reduce((sum, c) => sum + (c.unread_count ?? 0), 0)
}

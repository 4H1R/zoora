import { useAccess } from "react-access-engine"

// Mirrors the backend split (internal/conversations/service.go):
//  - view   → participate in conversations you belong to (DMs, groups, reactions…)
//  - manage → org chat admin: create groups/channels, moderate, post in channels
// Both still require the org plan to include the `chat` feature — gated
// separately via useHasFeature(FEATURE.chat).
export function useConversationPermissions() {
  const { can } = useAccess()
  return {
    canView: can("conversations:view") || can("conversations:manage"),
    canManage: can("conversations:manage"),
    // Whoever can actually purchase a plan — drives the upgrade CTA.
    canUpgrade: can("billing:manage"),
  }
}

import { useAccess } from "react-access-engine"

export function useLivesessionPermissions() {
  const { can } = useAccess()
  return {
    canView: can("live_sessions:view") || can("live_sessions:view_any"),
    canCreate: can("live_sessions:create"),
    canEdit: can("live_sessions:update") || can("live_sessions:update_any"),
    canDelete: can("live_sessions:manage") || can("live_sessions:manage_any"),
    canJoin: can("live_sessions:join") || can("live_sessions:join_any"),
    canManage: can("live_sessions:manage") || can("live_sessions:manage_any"),
  }
}

import { useAccess } from "react-access-engine"

export function useLivesessionPermissions() {
  const { can } = useAccess()
  return {
    canView: can("livesessions:view") || can("livesessions:view_any"),
    canCreate: can("livesessions:create") || can("livesessions:create_any"),
    canEdit: can("livesessions:update") || can("livesessions:update_any"),
    canDelete: can("livesessions:delete") || can("livesessions:delete_any"),
    canJoin: can("livesessions:join") || can("livesessions:join_any"),
    canManage: can("livesessions:manage") || can("livesessions:manage_any"),
  }
}

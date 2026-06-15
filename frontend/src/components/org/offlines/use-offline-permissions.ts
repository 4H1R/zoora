import { useAccess } from "react-access-engine"

export function useOfflinePermissions() {
  const { can } = useAccess()
  return {
    canView: can("offlines:view") || can("offlines:view_any"),
    canCreate: can("offlines:create") || can("offlines:create_any"),
    canEdit: can("offlines:update") || can("offlines:update_any"),
    canDelete: can("offlines:delete") || can("offlines:delete_any"),
  }
}

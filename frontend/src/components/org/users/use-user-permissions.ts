import { useAccess } from "react-access-engine"

export function useUserPermissions() {
  const { can } = useAccess()
  return {
    canView: can("users:view") || can("users:view_any"),
    canCreate: can("users:create"),
    canEdit: can("users:update") || can("users:update_any"),
    canDelete: can("users:delete") || can("users:delete_any"),
  }
}

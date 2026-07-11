import { useAccess } from "react-access-engine"

export function useClassPermissions() {
  const { can } = useAccess()
  return {
    canView: can("classes:view") || can("classes:view_any"),
    canCreate: can("classes:create") || can("classes:create_any"),
    canCreateAny: can("classes:create_any"),
    canEdit: can("classes:update") || can("classes:update_any"),
    canDelete: can("classes:delete") || can("classes:delete_any"),
    canJoin: can("classes:join"),
  }
}

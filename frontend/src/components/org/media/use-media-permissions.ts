import { useAccess } from "react-access-engine"

export function useMediaPermissions() {
  const { can } = useAccess()
  return {
    canView: can("media:view") || can("media:view_any"),
    canCreate: can("media:create") || can("media:create_any"),
    canDelete: can("media:delete") || can("media:delete_any"),
  }
}

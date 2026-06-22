import { useAccess } from "react-access-engine"

export function useMediaPermissions() {
  const { can } = useAccess()
  return {
    canView: can("media:view"),
    canCreate: can("media:create"),
    canDelete: can("media:delete") || can("media:delete_any"),
  }
}

import { useAccess } from "react-access-engine"

export function useRolePermissions() {
  const { can } = useAccess()
  return {
    canView: can("roles:view"),
    canCreate: can("roles:create"),
    canEdit: can("roles:update"),
    canDelete: can("roles:delete"),
  }
}

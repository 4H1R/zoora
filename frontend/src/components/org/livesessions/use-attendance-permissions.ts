import { useAccess } from "react-access-engine"

export function useAttendancePermissions() {
  const { can } = useAccess()
  return {
    canView: can("attendance:view") || can("attendance:view_any"),
    canCreate: can("attendance:create") || can("attendance:create_any"),
    canEdit: can("attendance:update") || can("attendance:update_any"),
    canDelete: can("attendance:delete"),
  }
}

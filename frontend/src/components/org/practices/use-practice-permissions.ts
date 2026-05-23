import { useAccess } from "react-access-engine"

export function usePracticePermissions() {
  const { can } = useAccess()
  return {
    canView: can("practices:view") || can("practices:view_any"),
    canCreate: can("practices:create") || can("practices:create_any"),
    canEdit: can("practices:update") || can("practices:update_any"),
    canDelete: can("practices:delete") || can("practices:delete_any"),
    canSubmit: can("practices:submit"),
    canGrade: can("practices:grade"),
  }
}

import { useAccess } from "react-access-engine"

export function useBankPermissions() {
  const { can } = useAccess()
  return {
    canView: can("question_banks:view") || can("question_banks:view_any"),
    canCreate: can("question_banks:create") || can("question_banks:create_any"),
    canEdit: can("question_banks:update") || can("question_banks:update_any"),
    canDelete: can("question_banks:delete") || can("question_banks:delete_any"),
  }
}

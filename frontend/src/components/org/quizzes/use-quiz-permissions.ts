import { useAccess } from "react-access-engine"

export function useQuizPermissions() {
  const { can } = useAccess()
  return {
    canView: can("quizzes:view") || can("quizzes:view_any"),
    canCreate: can("quizzes:create"),
    canEdit: can("quizzes:update") || can("quizzes:update_any"),
    canDelete: can("quizzes:delete") || can("quizzes:delete_any"),
  }
}

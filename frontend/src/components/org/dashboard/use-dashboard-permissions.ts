import { useAccess } from "react-access-engine"

export function useDashboardPermissions() {
  const { can } = useAccess()
  return {
    canViewClasses: can("classes:view") || can("classes:view_any"),
    canViewQuizzes: can("quizzes:view") || can("quizzes:view_any"),
    canViewUsers: can("users:view") || can("users:view_any"),
    canCreateClass: can("classes:create"),
  }
}

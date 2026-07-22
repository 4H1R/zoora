import { useAccess } from "react-access-engine"

export function useExamPermissions() {
  const { can } = useAccess()
  return {
    // quizzes:view marks staff (teachers/managers); students only hold quizzes:take.
    canManage: can("quizzes:view"),
    canTake: can("quizzes:take"),
  }
}

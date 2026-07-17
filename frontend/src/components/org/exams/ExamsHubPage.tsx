import { useOrgGuard } from "@/lib/access"

import { ManagerExamsView } from "./ManagerExamsView"
import { StudentExamsView } from "./StudentExamsView"
import { useExamPermissions } from "./use-exam-permissions"

export function ExamsHubPage() {
  const allowed = useOrgGuard(["quizzes:view", "quizzes:take"])
  const { canManage } = useExamPermissions()

  if (!allowed) return null

  // Staff see the org-wide oversight table; students see their own exam list.
  return canManage ? <ManagerExamsView /> : <StudentExamsView />
}

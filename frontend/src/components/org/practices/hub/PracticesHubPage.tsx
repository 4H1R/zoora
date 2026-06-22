import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { useOrgGuard } from "@/lib/access"

import { ManagerPracticesView } from "./ManagerPracticesView"
import { StudentPracticesView } from "./StudentPracticesView"

export function PracticesHubPage() {
  const allowed = useOrgGuard(["practices:view", "practices:view_any"])
  const { canGrade, canViewAny } = usePracticePermissions()

  if (!allowed) return null

  // Managers (org-wide viewers or graders) get the oversight data-table; everyone
  // else gets the personal "my homework" list.
  const isManager = canGrade || canViewAny
  return isManager ? <ManagerPracticesView /> : <StudentPracticesView />
}

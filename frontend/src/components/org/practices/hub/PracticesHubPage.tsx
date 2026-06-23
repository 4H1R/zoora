import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { useOrgGuard } from "@/lib/access"

import { ManagerPracticesView } from "./ManagerPracticesView"
import { StudentPracticesView } from "./StudentPracticesView"

export function PracticesHubPage() {
  const allowed = useOrgGuard(["practices:view", "practices:view_any"])
  const { canGrade, canViewAny } = usePracticePermissions()

  if (!allowed) return null

  // Managers (graders or org-wide viewers) see the oversight table; others see their own list.
  const isManager = canGrade || canViewAny
  return isManager ? <ManagerPracticesView /> : <StudentPracticesView />
}

import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { ManagerGradesView } from "@/components/org/grades/ManagerGradesView"
import { StudentGradesView } from "@/components/org/grades/StudentGradesView"
import { useCanAny, useOrgGuard } from "@/lib/access"
import { adminSearchSchema } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

const gradesSearchSchema = adminSearchSchema.extend({
  class_id: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/grades/")({
  head: () => orgHead("org.nav.grades"),
  validateSearch: gradesSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const allowed = useOrgGuard(["gradebook:view", "gradebook:view_any", "gradebook:create"])
  // Graders (column creators or org-wide viewers) manage class gradebooks;
  // everyone else sees their own report card.
  const isManager = useCanAny(["gradebook:view_any", "gradebook:create"])

  if (!allowed) return null

  return isManager ? <ManagerGradesView /> : <StudentGradesView />
}

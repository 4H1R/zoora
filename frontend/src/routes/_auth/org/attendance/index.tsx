import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { ManagerAttendanceView } from "@/components/org/attendance/ManagerAttendanceView"
import { StudentAttendanceView } from "@/components/org/attendance/StudentAttendanceView"
import { useCanAny, useOrgGuard } from "@/lib/access"
import { adminSearchSchema } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

const attendanceSearchSchema = adminSearchSchema.extend({
  class_id: z.string().optional(),
  class_session_id: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/attendance/")({
  head: () => orgHead("org.nav.attendance"),
  validateSearch: attendanceSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const allowed = useOrgGuard(["attendance:view", "attendance:view_any", "attendance:create"])
  // Attendance takers (markers or org-wide viewers) manage class matrices;
  // everyone else sees their own record.
  const isManager = useCanAny(["attendance:view_any", "attendance:create"])

  if (!allowed) return null

  return isManager ? <ManagerAttendanceView /> : <StudentAttendanceView />
}

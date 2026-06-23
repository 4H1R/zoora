import { createFileRoute } from "@tanstack/react-router"
import { useAccess } from "react-access-engine"

import { AdminDashboard } from "@/components/org/dashboard/admin-dashboard"
import { StudentDashboard } from "@/components/org/dashboard/student-dashboard"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/dashboard")({
  head: () => orgHead("org.nav.dashboard"),
  component: RouteComponent,
})

function RouteComponent() {
  const { can } = useAccess()
  const isManager = can("classes:create") || can("users:view") || can("users:view_any")
  return isManager ? <AdminDashboard /> : <StudentDashboard />
}

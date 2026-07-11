import { useTranslation } from "react-i18next"

import { SidebarBreadcrumb } from "@/components/layout/sidebar-breadcrumb"

const SEGMENT_KEYS: Record<string, string> = {
  dashboard: "admin.dashboard.title",
  account: "account.title",
  changelog: "admin.changelog.title",
  tutorials: "admin.tutorials.title",
  classes: "admin.nav.classes",
  corrections: "admin.corrections.title",
  organizations: "admin.organizations",
  leads: "admin.leads.title",
  permissions: "admin.permissions.title",
  roles: "admin.nav.roles",
  users: "admin.nav.users",
  sessions: "admin.sessions.title",
  quizzes: "admin.quizzes.title",
  practices: "admin.practices.title",
  offlines: "admin.offlines.title",
  "live-rooms": "admin.liveRooms.title",
  attendance: "admin.attendance.title",
  questions: "admin.questions.title",
  gradebook: "admin.gradebook.title",
}

export function AdminBreadcrumb({ className }: { className?: string }) {
  const { t } = useTranslation()

  return (
    <SidebarBreadcrumb
      className={className}
      prefixLabel={t("admin.panel")}
      pathPrefix={/^\/admin\/?/}
      segmentKeys={SEGMENT_KEYS}
    />
  )
}

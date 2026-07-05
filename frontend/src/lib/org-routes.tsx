import type { AppPermission } from "@/lib/access"

import {
  CalendarCheckIcon,
  CalendarIcon,
  ClipboardListIcon,
  FileIcon,
  GraduationCapIcon,
  LayoutDashboardIcon,
  NotebookPenIcon,
  SchoolIcon,
  SettingsIcon,
  ShieldIcon,
  UsersIcon,
  VideoIcon,
} from "lucide-react"

// OrgRouteKey identifies every org-scoped route that appears in either the
// sidebar nav or the dashboard launcher tiles.
export type OrgRouteKey =
  | "dashboard"
  | "calendar"
  | "classes"
  | "online-classes"
  | "exams"
  | "practices"
  | "grades"
  | "attendance"
  | "users"
  | "roles"
  | "settings"
  | "files"

// OrgRouteSpec is the single source of truth for the per-route metadata shared
// between the sidebar nav (org-nav.tsx) and the dashboard tiles
// (use-dashboard-tiles.tsx): icon, i18n key, the path segment under
// /org/$orgId, and the permissions that gate visibility.
//
// Each consumer decides WHICH routes to show and in what order/grouping; this
// table only owns the metadata that must stay in sync across both.
export type OrgRouteSpec = {
  i18nKey: string
  icon: React.ReactNode
  segment: string
  // perms gate visibility; undefined means always visible to org members.
  perms?: AppPermission[]
}

export const ORG_ROUTES: Record<OrgRouteKey, OrgRouteSpec> = {
  dashboard: {
    i18nKey: "org.nav.dashboard",
    icon: <LayoutDashboardIcon />,
    segment: "dashboard",
  },
  calendar: {
    i18nKey: "org.nav.calendar",
    icon: <CalendarIcon />,
    segment: "calendar",
  },
  classes: {
    i18nKey: "org.nav.classes",
    icon: <SchoolIcon />,
    segment: "classes",
    perms: ["classes:view", "classes:view_any"],
  },
  "online-classes": {
    i18nKey: "org.nav.onlineClasses",
    icon: <VideoIcon />,
    segment: "online-classes",
    perms: ["live_sessions:view", "live_sessions:view_any"],
  },
  exams: {
    i18nKey: "org.nav.exams",
    icon: <ClipboardListIcon />,
    segment: "exams",
    perms: ["quizzes:view", "quizzes:take"],
  },
  practices: {
    i18nKey: "org.nav.practices",
    icon: <NotebookPenIcon />,
    segment: "practices",
    perms: ["practices:view", "practices:view_any"],
  },
  grades: {
    i18nKey: "org.nav.grades",
    icon: <GraduationCapIcon />,
    segment: "grades",
    perms: ["gradebook:view"],
  },
  attendance: {
    i18nKey: "org.nav.attendance",
    icon: <CalendarCheckIcon />,
    segment: "attendance",
    perms: ["attendance:view"],
  },
  users: {
    i18nKey: "org.nav.users",
    icon: <UsersIcon />,
    segment: "users",
    perms: ["users:view", "users:view_any"],
  },
  roles: {
    i18nKey: "org.nav.roles",
    icon: <ShieldIcon />,
    segment: "roles",
    perms: ["roles:view"],
  },
  settings: {
    i18nKey: "org.nav.settings",
    icon: <SettingsIcon />,
    segment: "settings",
    perms: ["organizations:update"],
  },
  files: {
    i18nKey: "org.nav.files",
    icon: <FileIcon />,
    segment: "files",
    perms: ["media:view_any"],
  },
}

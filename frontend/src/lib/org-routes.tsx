import type { AppPermission } from "@/lib/access"
import type { FeatureKey } from "@/lib/entitlements"

import {
  BellIcon,
  CalendarCheckIcon,
  CalendarIcon,
  ClipboardListIcon,
  CreditCardIcon,
  FileIcon,
  GraduationCapIcon,
  LayoutDashboardIcon,
  LibraryIcon,
  MessageCircleIcon,
  NotebookPenIcon,
  SchoolIcon,
  SettingsIcon,
  ShieldIcon,
  TicketIcon,
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
  | "question-banks"
  | "practices"
  | "grades"
  | "attendance"
  | "users"
  | "roles"
  | "settings"
  | "billing"
  | "files"
  | "notifications"
  | "conversations"
  | "tickets"

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
  // feature gates visibility on the org's plan, in ADDITION to perms. When the
  // plan lacks it, the item is hidden — except for holders of featureExemptPerms.
  feature?: FeatureKey
  // featureExemptPerms still see the item when the plan lacks `feature`; they
  // land on an upgrade page. Surfaces plan-locked features to the people who can
  // actually upgrade (e.g. conversations:manage → org admins).
  featureExemptPerms?: AppPermission[]
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
  "question-banks": {
    i18nKey: "org.nav.questionBanks",
    icon: <LibraryIcon />,
    segment: "question-banks",
    perms: ["question_banks:view", "question_banks:view_any"],
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
    perms: ["gradebook:view", "gradebook:view_any", "gradebook:create"],
  },
  attendance: {
    i18nKey: "org.nav.attendance",
    icon: <CalendarCheckIcon />,
    segment: "attendance",
    perms: ["attendance:view", "attendance:view_any", "attendance:create"],
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
  billing: {
    i18nKey: "billing.title",
    icon: <CreditCardIcon />,
    segment: "billing",
    perms: ["billing:manage"],
  },
  files: {
    i18nKey: "org.nav.files",
    icon: <FileIcon />,
    segment: "files",
    perms: ["media:view_any"],
  },
  notifications: {
    i18nKey: "notifications.title",
    icon: <BellIcon />,
    segment: "notifications",
  },
  conversations: {
    i18nKey: "org.nav.conversations",
    icon: <MessageCircleIcon />,
    segment: "conversations",
    // Any participant (view) or org chat admin (manage) may reach it; the plan
    // gate + the manage-only exemption below decide what they actually see.
    perms: ["conversations:view", "conversations:manage"],
    feature: "chat",
    featureExemptPerms: ["conversations:manage"],
  },
  tickets: {
    i18nKey: "org.nav.tickets",
    icon: <TicketIcon />,
    segment: "tickets",
    perms: ["tickets:view", "tickets:manage"],
  },
}

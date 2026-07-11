import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"
import type { NavGroup } from "@/components/layout/nav-main"

import {
  BellIcon,
  Building2Icon,
  CalendarIcon,
  CheckSquareIcon,
  ClipboardCheckIcon,
  ClipboardListIcon,
  DumbbellIcon,
  FileVideoIcon,
  GraduationCapIcon,
  HelpCircleIcon,
  InboxIcon,
  KeyIcon,
  LayoutDashboardIcon,
  ReceiptTextIcon,
  SchoolIcon,
  ShieldIcon,
  SparklesIcon,
  TagIcon,
  TrophyIcon,
  UsersIcon,
  VideoIcon,
} from "lucide-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { OrgSwitcher } from "@/components/admin/layout/org-switcher"
import { AppSidebar as AppSidebarShared } from "@/components/layout/app-sidebar"
import { Sidebar } from "@/components/ui/sidebar"
import { useAdminStore } from "@/stores/admin"

export function AppSidebar({
  user,
  ...props
}: React.ComponentProps<typeof Sidebar> & {
  user?: GithubCom4H1RZooraInternalDomainUser
}) {
  const { t } = useTranslation()
  const { activeOrganization, setActiveOrganization } = useAdminStore()

  const navGroups: NavGroup[] = [
    {
      label: t("admin.nav.overview"),
      items: [
        { title: t("admin.dashboard.title"), url: "/admin/dashboard", icon: <LayoutDashboardIcon /> },
        { title: t("admin.organizations"), url: "/admin/organizations", icon: <Building2Icon /> },
        { title: t("admin.changelog.title"), url: "/admin/changelog", icon: <SparklesIcon /> },
        { title: t("admin.tutorials.title"), url: "/admin/tutorials", icon: <GraduationCapIcon /> },
        { title: t("notifications.title"), url: "/admin/notifications", icon: <BellIcon /> },
        { title: t("admin.leads.title"), url: "/admin/leads", icon: <InboxIcon /> },
      ],
    },
    {
      label: t("admin.nav.teaching"),
      items: [
        { title: t("admin.nav.classes"), url: "/admin/classes", icon: <SchoolIcon /> },
        { title: t("admin.nav.sessions"), url: "/admin/sessions", icon: <CalendarIcon /> },
        { title: t("admin.attendance.title"), url: "/admin/attendance", icon: <ClipboardCheckIcon /> },
        { title: t("admin.gradebook.title"), url: "/admin/gradebook", icon: <TrophyIcon /> },
      ],
    },
    {
      label: t("admin.nav.rooms"),
      items: [
        { title: t("admin.offlines.title"), url: "/admin/offlines", icon: <FileVideoIcon /> },
        { title: t("admin.liveRooms.title"), url: "/admin/live-rooms", icon: <VideoIcon /> },
        { title: t("admin.practices.title"), url: "/admin/practices", icon: <DumbbellIcon /> },
      ],
    },
    {
      label: t("admin.nav.assessments"),
      items: [
        { title: t("admin.quizzes.title"), url: "/admin/quizzes", icon: <ClipboardListIcon /> },
        { title: t("admin.corrections.title"), url: "/admin/corrections", icon: <CheckSquareIcon /> },
        { title: t("admin.questions.title"), url: "/admin/questions", icon: <HelpCircleIcon /> },
      ],
    },
    {
      label: t("admin.nav.users"),
      items: [
        { title: t("admin.nav.users"), url: "/admin/users", icon: <UsersIcon /> },
        { title: t("admin.nav.roles"), url: "/admin/roles", icon: <ShieldIcon /> },
        { title: t("admin.permissions.title"), url: "/admin/permissions", icon: <KeyIcon /> },
      ],
    },
    {
      label: t("billing.title"),
      items: [
        { title: t("billing.admin.prices"), url: "/admin/billing/prices", icon: <TagIcon /> },
        { title: t("billing.admin.invoicesTitle"), url: "/admin/billing/invoices", icon: <ReceiptTextIcon /> },
      ],
    },
  ]

  return (
    <AppSidebarShared
      user={user}
      navGroups={navGroups}
      accountTo="/admin/account"
      headerExtra={<OrgSwitcher selected={activeOrganization} onSelect={setActiveOrganization} />}
      {...props}
    />
  )
}

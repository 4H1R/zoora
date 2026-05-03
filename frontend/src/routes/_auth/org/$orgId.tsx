import { createFileRoute, Outlet } from "@tanstack/react-router"
import { FileIcon, LayoutDashboardIcon, SchoolIcon, SettingsIcon, ShieldIcon, UsersIcon } from "lucide-react"
import { AccessProvider } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetRoles } from "@/api/roles/roles"
import { useGetUsersMe } from "@/api/users/users"
import { LanguageSwitcher } from "@/components/language-switcher"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { SidebarBreadcrumb } from "@/components/layout/sidebar-breadcrumb"
import { ThemeToggle } from "@/components/theme-toggle"
import { useDirection } from "@/components/ui/direction"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { buildAccess } from "@/lib/access"

export const Route = createFileRoute("/_auth/org/$orgId")({
  component: RouteComponent,
})

const SEGMENT_KEYS: Record<string, string> = {
  dashboard: "org.nav.dashboard",
  classes: "org.nav.classes",
  members: "org.nav.members",
  users: "org.nav.users",
  roles: "org.nav.roles",
  settings: "org.nav.settings",
  files: "org.nav.files",
}

function RouteComponent() {
  const { orgId } = Route.useParams()
  const { t } = useTranslation()
  const { data } = useGetUsersMe()
  const direction = useDirection()
  const sidebarSide = direction === "rtl" ? "right" : "left"

  const { data: rolesData } = useGetRoles({ organization_id: orgId })

  const user = (data?.status === 200 && data.data.data) || undefined
  const allRoles = (rolesData?.status === 200 && rolesData.data.data) || []
  const access = user ? buildAccess(user, allRoles) : null

  const navGroups = [
    {
      label: t("org.panel"),
      items: [
        { title: t("org.nav.dashboard"), url: `/org/${orgId}/dashboard`, icon: <LayoutDashboardIcon /> },
        { title: t("org.nav.classes"), url: `/org/${orgId}/classes`, icon: <SchoolIcon /> },
      ],
    },
    {
      label: t("org.nav.management"),
      items: [
        { title: t("org.nav.users"), url: `/org/${orgId}/users`, icon: <UsersIcon /> },
        { title: t("org.nav.roles"), url: `/org/${orgId}/roles`, icon: <ShieldIcon /> },
        { title: t("org.nav.settings"), url: `/org/${orgId}/settings`, icon: <SettingsIcon /> },
        { title: t("org.nav.files"), url: `/org/${orgId}/files`, icon: <FileIcon /> },
      ],
    },
  ]

  if (!access) return null

  return (
    <AccessProvider config={access.config} user={access.user}>
      <SidebarProvider>
        <AppSidebar user={user} navGroups={navGroups} side={sidebarSide} />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="md:hidden" />
            <SidebarBreadcrumb
              className="hidden md:flex"
              prefixLabel={t("org.panel")}
              pathPrefix={new RegExp(`^/org/${orgId}/?`)}
              segmentKeys={SEGMENT_KEYS}
            />
            <div className="ms-auto flex items-center gap-2">
              <LanguageSwitcher />
              <ThemeToggle />
            </div>
          </header>
          <div className="flex flex-1 flex-col gap-4 px-4 py-4 lg:px-8">
            <Outlet />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </AccessProvider>
  )
}

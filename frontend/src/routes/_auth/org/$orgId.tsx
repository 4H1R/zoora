import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router"
import { FileIcon, LayoutDashboardIcon, SchoolIcon, SettingsIcon, ShieldIcon, UsersIcon } from "lucide-react"
import { useEffect } from "react"
import { AccessProvider } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetOrganizationsId } from "@/api/organizations/organizations"
import { useGetRoles } from "@/api/roles/roles"
import { useGetUsersMe } from "@/api/users/users"
import { LanguageSwitcher } from "@/components/language-switcher"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { SidebarBreadcrumb } from "@/components/layout/sidebar-breadcrumb"
import { SplashScreen } from "@/components/splash-screen"
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
  const { data, isLoading: userLoading } = useGetUsersMe()
  const direction = useDirection()
  const sidebarSide = direction === "rtl" ? "right" : "left"
  const navigate = useNavigate()

  const { data: rolesData } = useGetRoles()
  const { data: orgData, isLoading: orgLoading } = useGetOrganizationsId(orgId)

  const user = (data?.status === 200 && data.data.data) || undefined
  const allRoles = (rolesData?.status === 200 && rolesData.data.data) || []
  const access = user ? buildAccess(user, allRoles) : null

  useEffect(() => {
    if (userLoading || orgLoading) return
    if (!user || user.organization_id !== orgId) {
      navigate({ to: "/" })
    }
  }, [userLoading, orgLoading, user, orgId, orgData, navigate])

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

  if (userLoading || orgLoading || !access) return <SplashScreen />
  if (!user || user.organization_id !== orgId) {
    navigate({ to: "/" })
    return null
  }

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
          <div className="container flex flex-1 flex-col gap-4 py-4">
            <Outlet />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </AccessProvider>
  )
}

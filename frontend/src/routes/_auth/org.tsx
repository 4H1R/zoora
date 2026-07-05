import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"
import { AccessProvider } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetUsersMe } from "@/api/users/users"
import { WhatsNewButton } from "@/components/changelog/whats-new-button"
import { LanguageSwitcher } from "@/components/language-switcher"
// import { LiveClock } from "@/components/live-clock"
import { BreadcrumbProvider } from "@/components/layout/breadcrumb-context"
import { AppSidebar } from "@/components/layout/app-sidebar"
import { MobileBreadcrumb } from "@/components/layout/mobile-breadcrumb"
import { SidebarBreadcrumb } from "@/components/layout/sidebar-breadcrumb"
import { SplashScreen } from "@/components/splash-screen"
import { ThemeToggle } from "@/components/theme-toggle"
import { useDirection } from "@/components/ui/direction"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { buildAccess } from "@/lib/access"
import { buildOrgNavGroups } from "@/lib/org-nav"
import { ORG_ROUTES } from "@/lib/org-routes"

export const Route = createFileRoute("/_auth/org")({
  component: RouteComponent,
})

// Map each top-level path segment to its i18n label for the breadcrumb. Derived
// from ORG_ROUTES (the single source of truth shared with the sidebar nav) so it
// can never drift out of sync — adding a route there auto-labels its breadcrumb.
const SEGMENT_KEYS: Record<string, string> = {
  ...Object.fromEntries(
    Object.values(ORG_ROUTES).map((spec) => [spec.segment, spec.i18nKey])
  ),
  members: "org.nav.members",
}

function RouteComponent() {
  const { t } = useTranslation()
  const { data, isLoading: userLoading } = useGetUsersMe()
  const direction = useDirection()
  const sidebarSide = direction === "rtl" ? "right" : "left"
  const navigate = useNavigate()

  const user = (data?.status === 200 && data.data.data) || undefined
  const access = user ? buildAccess(user) : null

  // The org boundary is the host (the backend asserts caller.org == host.org on
  // every request). A logged-in user with no org membership (e.g. a platform
  // admin landing here) has nothing to show, so bounce to the root resolver.
  useEffect(() => {
    if (userLoading) return
    if (!user || !user.organization_id) {
      navigate({ to: "/" })
    }
  }, [userLoading, user, navigate])

  if (userLoading || !access) return <SplashScreen />
  if (!user || !user.organization_id) return null

  const navGroups = buildOrgNavGroups(t, access.has)

  return (
    <AccessProvider config={access.config} user={access.user}>
      <BreadcrumbProvider>
        <SidebarProvider>
          <AppSidebar user={user} navGroups={navGroups} side={sidebarSide} />
          <SidebarInset>
            <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
              <SidebarTrigger className="md:hidden" />
              <SidebarBreadcrumb
                className="hidden md:flex"
                prefixLabel={t("org.panel")}
                pathPrefix={new RegExp(`^/org/?`)}
                segmentKeys={SEGMENT_KEYS}
              />
              <div className="ms-auto flex items-center gap-2">
                {/* <LiveClock className="me-1 hidden sm:flex" /> */}
                <WhatsNewButton />
                <LanguageSwitcher />
                <ThemeToggle />
              </div>
            </header>
            <MobileBreadcrumb className="px-4 pb-2 pt-4 md:hidden" />
            <div className="container flex flex-1 flex-col gap-4 py-4">
              <Outlet />
            </div>
          </SidebarInset>
        </SidebarProvider>
      </BreadcrumbProvider>
    </AccessProvider>
  )
}

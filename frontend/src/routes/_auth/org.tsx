import { createFileRoute, Outlet, useNavigate, useRouterState } from "@tanstack/react-router"
import { useEffect } from "react"
import { AccessProvider } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetUsersMe } from "@/api/users/users"
import { MajorModal } from "@/components/changelog/major-modal"
import { NavZoora } from "@/components/changelog/nav-zoora"
import { AppSidebar } from "@/components/layout/app-sidebar"
// import { LiveClock } from "@/components/live-clock"
import { BreadcrumbProvider } from "@/components/layout/breadcrumb-context"
import { MobileBreadcrumb } from "@/components/layout/mobile-breadcrumb"
import { SidebarBreadcrumb } from "@/components/layout/sidebar-breadcrumb"
import { LanguageSwitcher } from "@/components/language-switcher"
import { NotificationBell } from "@/components/notifications/notification-bell"
import { ChatProvider } from "@/components/org/conversations/chat-provider"
import { SplashScreen } from "@/components/splash-screen"
import { ThemeToggle } from "@/components/theme-toggle"
import { useDirection } from "@/components/ui/direction"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { buildAccess } from "@/lib/access"
import { useFeatureGate } from "@/lib/entitlements"
import { buildOrgNavGroups } from "@/lib/org-nav"
import { ORG_ROUTES } from "@/lib/org-routes"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org")({
  component: RouteComponent,
})

// Map each top-level path segment to its i18n label for the breadcrumb. Derived
// from ORG_ROUTES (the single source of truth shared with the sidebar nav) so it
// can never drift out of sync — adding a route there auto-labels its breadcrumb.
const SEGMENT_KEYS: Record<string, string> = {
  ...Object.fromEntries(Object.values(ORG_ROUTES).map((spec) => [spec.segment, spec.i18nKey])),
  members: "org.nav.members",
  tutorials: "tutorials.title",
  "whats-new": "whatsNew.title",
  notifications: "notifications.title",
  account: "org.nav.account",
}

function RouteComponent() {
  const { t } = useTranslation()
  const { data, isLoading: userLoading } = useGetUsersMe()
  const hasFeature = useFeatureGate()
  const direction = useDirection()
  const sidebarSide = direction === "rtl" ? "right" : "left"
  const navigate = useNavigate()

  // Conversations is a full-height master-detail chat that should span the whole
  // available width on desktop rather than being capped by the centered
  // `container` max-width used by every other org page.
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const convBase = `/org/${ORG_ROUTES.conversations.segment}`
  const fullBleed = pathname.startsWith(convBase)
  // A specific conversation is open (detail pane, not the list). On mobile this
  // thread should own the whole screen — hide the app header + breadcrumb and
  // drop the surrounding padding so it goes edge to edge. Desktop is unchanged.
  const chatDetail = fullBleed && pathname.slice(convBase.length).replace(/\//g, "").length > 0

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

  const navGroups = buildOrgNavGroups(t, access.has, hasFeature)

  const layout = (
    <AccessProvider config={access.config} user={access.user}>
      <BreadcrumbProvider>
        <SidebarProvider>
          <AppSidebar user={user} navGroups={navGroups} side={sidebarSide} contentExtra={<NavZoora />} />
          <SidebarInset>
            <header
              className={cn("flex h-16 shrink-0 items-center gap-2 border-b px-4", chatDetail && "hidden md:flex")}
            >
              <SidebarTrigger className="md:hidden" />
              <SidebarBreadcrumb
                className="hidden md:flex"
                prefixLabel={t("org.panel")}
                pathPrefix={new RegExp(`^/org/?`)}
                segmentKeys={SEGMENT_KEYS}
              />
              <div className="ms-auto flex items-center gap-2">
                {/* <LiveClock className="me-1 hidden sm:flex" /> */}
                <NotificationBell to="/org/notifications" />
                <LanguageSwitcher />
                <ThemeToggle />
              </div>
            </header>
            <MajorModal />
            {!chatDetail && <MobileBreadcrumb className="px-4 pt-4 pb-2 md:hidden" />}
            <div
              className={cn(
                "flex flex-1 flex-col",
                fullBleed ? "min-h-0" : "container",
                chatDetail ? "gap-0 p-0 md:gap-4 md:px-4 md:py-4" : cn("gap-4 py-4", fullBleed && "px-4")
              )}
            >
              <Outlet />
            </div>
          </SidebarInset>
        </SidebarProvider>
      </BreadcrumbProvider>
    </AccessProvider>
  )

  // Only orgs on the chat plan open a realtime WS connection; others render the
  // layout unchanged with no ChatProvider (and thus no socket).
  return hasFeature("chat") ? <ChatProvider>{layout}</ChatProvider> : layout
}

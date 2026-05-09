import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"

import { useGetUsersMe } from "@/api/users/users"
import { AdminBreadcrumb } from "@/components/admin/layout/admin-breadcrumb"
import { SplashScreen } from "@/components/splash-screen"
import { AppSidebar } from "@/components/admin/layout/app-sidebar"
import { LanguageSwitcher } from "@/components/language-switcher"
import { ThemeToggle } from "@/components/theme-toggle"
import { useDirection } from "@/components/ui/direction"
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"

export const Route = createFileRoute("/_admin")({
  component: RouteComponent,
})

function RouteComponent() {
  const { data, isLoading, isFetching } = useGetUsersMe()
  const navigate = useNavigate()
  const direction = useDirection()
  const sidebarSide = direction === "rtl" ? "right" : "left"

  useEffect(() => {
    if (isLoading || isFetching) return
    if (data?.status === 200 && data.data.data?.is_admin) return
    navigate({ to: "/" })
  }, [navigate, data, isLoading, isFetching])

  const user = (data?.status === 200 && data.data.data) || undefined

  if (isLoading || isFetching) return <SplashScreen />

  return (
    <SidebarProvider>
      <AppSidebar user={user} side={sidebarSide} />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="md:hidden" />
          <AdminBreadcrumb className="hidden md:flex" />
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
  )
}

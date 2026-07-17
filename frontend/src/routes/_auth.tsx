import { createFileRoute, Outlet, useLocation, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"

import { useGetUsersMe } from "@/api/users/users"

export const Route = createFileRoute("/_auth")({
  // Client-only: Nitro serves a shell and the browser renders this subtree, so
  // its client-only deps are never SSR'd. Only '/' is prerendered with content.
  ssr: false,
  component: RouteComponent,
})

function RouteComponent() {
  const { isSuccess, isLoading, isFetching } = useGetUsersMe()
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    if (isSuccess || isLoading || isFetching) return
    // Never capture the login URL itself — that would nest redirect params.
    if (location.pathname === "/login") return
    // Preserve the originally-requested path (+ its own query) so login can send
    // the user back. Depend on pathname/searchStr, not href, so this effect does
    // not feed on the redirect param it writes.
    navigate({
      to: "/login",
      search: { redirect: location.pathname + location.searchStr },
      replace: true,
    })
  }, [isSuccess, navigate, isLoading, isFetching, location.pathname, location.searchStr])

  return <Outlet />
}

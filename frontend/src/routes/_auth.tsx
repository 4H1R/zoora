import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router"
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

  useEffect(() => {
    if (isSuccess || isLoading || isFetching) return
    navigate({ to: "/login" })
  }, [isSuccess, navigate, isLoading, isFetching])

  return <Outlet />
}

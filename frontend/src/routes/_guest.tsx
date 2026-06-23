import { createFileRoute, Outlet, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"

import { useGetUsersMe } from "@/api/users/users"

export const Route = createFileRoute("/_guest")({
  component: RouteComponent,
})

function RouteComponent() {
  const { data, isError, isLoading } = useGetUsersMe()
  const navigate = useNavigate()

  useEffect(() => {
    if (isError || isLoading) return

    const user = (data?.status === 200 && data.data.data) || undefined
    if (user?.is_admin) {
      navigate({ to: "/admin/dashboard" })
    } else if (user?.organization_id) {
      navigate({ to: "/org/dashboard" })
    }
  }, [isError, navigate, isLoading, data])

  return <Outlet />
}

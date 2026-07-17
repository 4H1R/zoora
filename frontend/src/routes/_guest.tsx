import { createFileRoute, Outlet, useNavigate, useSearch } from "@tanstack/react-router"
import { useEffect } from "react"

import { useGetUsersMe } from "@/api/users/users"
import { safeRedirectPath } from "@/lib/redirect"

export const Route = createFileRoute("/_guest")({
  // Client-only: Nitro serves a shell and the browser renders this subtree, so
  // its client-only deps are never SSR'd. Only '/' is prerendered with content.
  ssr: false,
  component: RouteComponent,
})

function RouteComponent() {
  const { data, isError, isLoading } = useGetUsersMe()
  const navigate = useNavigate()
  const search = useSearch({ strict: false }) as { redirect?: string }

  useEffect(() => {
    if (isError || isLoading) return

    const user = (data?.status === 200 && data.data.data) || undefined
    if (!user) return

    const target = safeRedirectPath(search.redirect)
    if (target) {
      navigate({ to: target })
    } else if (user.is_admin) {
      navigate({ to: "/admin/dashboard" })
    } else if (user.organization_id) {
      navigate({ to: "/org/dashboard" })
    }
  }, [isError, navigate, isLoading, data, search.redirect])

  return <Outlet />
}

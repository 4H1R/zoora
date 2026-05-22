import { createFileRoute, useNavigate } from "@tanstack/react-router"

import { useRequirePerm } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/medias/")({
  head: () => orgHead("org.nav.files"),
  component: RouteComponent,
})

function RouteComponent() {
  const { orgId } = Route.useParams()
  const navigate = useNavigate()
  const allowed = useRequirePerm(["media:view", "media:view_any"], () =>
    navigate({ to: "/org/$orgId/dashboard", params: { orgId } })
  )
  if (!allowed) return null
  return <div>Hello "/_auth/org/$orgId/medias/"!</div>
}

import { createFileRoute } from "@tanstack/react-router"

import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/medias/")({
  head: () => orgHead("org.nav.files"),
  component: RouteComponent,
})

function RouteComponent() {
  const allowed = useOrgGuard(["media:view", "media:view_any"])
  if (!allowed) return null
  return <div>Hello "/_auth/org/$orgId/medias/"!</div>
}

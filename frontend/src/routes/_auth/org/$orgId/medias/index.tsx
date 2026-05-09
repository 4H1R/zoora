import { createFileRoute } from "@tanstack/react-router"

import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/medias/")({
  head: () => orgHead("org.nav.files"),
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_auth/org/$orgId/medias/"!</div>
}

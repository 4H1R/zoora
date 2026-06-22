import { createFileRoute, redirect } from "@tanstack/react-router"

export const Route = createFileRoute("/_auth/org/$orgId/")({
  beforeLoad: ({ params }) => {
    throw redirect({ to: "/org/$orgId/dashboard", params: { orgId: params.orgId } })
  },
})

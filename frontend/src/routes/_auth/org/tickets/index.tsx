import { createFileRoute } from "@tanstack/react-router"

import { TicketsPage } from "@/components/org/tickets/tickets-page"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

type TicketsSearch = { ticket?: string }

export const Route = createFileRoute("/_auth/org/tickets/")({
  head: () => orgHead("org.nav.tickets"),
  validateSearch: (search: Record<string, unknown>): TicketsSearch => ({
    ticket: typeof search.ticket === "string" ? search.ticket : undefined,
  }),
  component: RouteComponent,
})

function RouteComponent() {
  const allowed = useOrgGuard(["tickets:view", "tickets:manage"])
  if (!allowed) return null
  return <TicketsPage />
}

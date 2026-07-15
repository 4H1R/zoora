import type { TicketFilters } from "@/components/org/tickets/ticket-list"

import { useNavigate } from "@tanstack/react-router"
import { PlusIcon } from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { useGetTickets } from "@/api/tickets/tickets"
import { CreateTicketDialog } from "@/components/org/tickets/create-ticket-dialog"
import { TicketList } from "@/components/org/tickets/ticket-list"
import { TicketThread } from "@/components/org/tickets/ticket-thread"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/lib/utils"
import { Route } from "@/routes/_auth/org/tickets/index"

export function TicketsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { user } = useAccess()
  const { ticket: selectedId } = Route.useSearch()

  const [filters, setFilters] = useState<TicketFilters>({})
  const [createOpen, setCreateOpen] = useState(false)

  const { data, isLoading } = useGetTickets({
    status: filters.status,
    type: filters.type,
    class_id: filters.classId,
    page_size: 100,
  })
  const tickets = (data?.status === 200 && data.data.data?.items) || []

  const select = (id: string | undefined) => navigate({ to: ".", search: id ? { ticket: id } : {}, replace: false })

  return (
    <div className="flex flex-1 flex-col gap-4">
      <PageHeader
        title={t("tickets.title")}
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4" />
            {t("tickets.new")}
          </Button>
        }
      />

      <Card className="flex min-h-0 flex-1 flex-row gap-0 overflow-hidden p-0 md:h-[calc(100dvh-12rem)]">
        {/* List pane: full width on mobile until a ticket is selected. */}
        <div className={cn("w-full min-w-0 md:block md:w-80 md:border-e lg:w-96", selectedId && "hidden")}>
          <TicketList
            tickets={tickets}
            isLoading={isLoading}
            selectedId={selectedId}
            onSelect={select}
            filters={filters}
            onFiltersChange={setFilters}
            currentUserId={user.id}
          />
        </div>
        {/* Thread pane: hidden on mobile until a ticket is selected. */}
        <div className={cn("min-w-0 flex-1 md:block", !selectedId && "hidden")}>
          <TicketThread ticketId={selectedId} currentUserId={user.id} onBack={() => select(undefined)} />
        </div>
      </Card>

      <CreateTicketDialog open={createOpen} onOpenChange={setCreateOpen} onCreated={select} />
    </div>
  )
}

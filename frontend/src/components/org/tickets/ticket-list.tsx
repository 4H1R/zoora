import type { GithubCom4H1RZooraInternalDomainTicket as Ticket } from "@/api/model"

import { TicketIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClasses } from "@/api/classes/classes"
import { TicketStatusBadge } from "@/components/org/tickets/ticket-badges"
import { EmptyState } from "@/components/ui/empty-state"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { formatRelativeTime } from "@/lib/relative-time"
import { cn } from "@/lib/utils"

export type TicketFilters = {
  status?: string
  type?: string
  classId?: string
}

const STATUS_VALUES = ["all", "open", "answered", "closed"] as const
const TYPE_VALUES = ["all", "question", "grade_objection", "other"] as const

export function TicketList({
  tickets,
  isLoading,
  selectedId,
  onSelect,
  filters,
  onFiltersChange,
  currentUserId,
}: {
  tickets: Ticket[]
  isLoading: boolean
  selectedId?: string
  onSelect: (id: string) => void
  filters: TicketFilters
  onFiltersChange: (next: TicketFilters) => void
  currentUserId: string
}) {
  const { t, i18n } = useTranslation()

  const { data: classesData } = useGetClasses()
  const classes = (classesData?.status === 200 && classesData.data.data?.items) || []

  const statusItems = STATUS_VALUES.map((v) => ({
    value: v,
    label: v === "all" ? t("tickets.filters.all") : t(`tickets.status.${v}`),
  }))
  const typeItems = TYPE_VALUES.map((v) => ({
    value: v,
    label: v === "all" ? t("tickets.filters.all") : t(`tickets.type.${v}`),
  }))
  const classItems = [
    { value: "all", label: t("tickets.filters.all") },
    ...classes.map((c) => ({ value: c.id ?? "", label: c.name ?? "" })),
  ]

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-wrap gap-2 border-b p-3">
        <div className="min-w-0 flex-1">
          <label className="text-muted-foreground mb-1.5 block text-xs font-medium">
            {t("tickets.filters.status")}
          </label>
          <Select
            items={statusItems}
            value={filters.status ?? "all"}
            onValueChange={(v) => onFiltersChange({ ...filters, status: v === "all" ? undefined : (v ?? undefined) })}
          >
            <SelectTrigger size="sm" className="w-full" aria-label={t("tickets.filters.status")}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="min-w-0 flex-1">
          <label className="text-muted-foreground mb-1.5 block text-xs font-medium">
            {t("tickets.filters.type")}
          </label>
          <Select
            items={typeItems}
            value={filters.type ?? "all"}
            onValueChange={(v) => onFiltersChange({ ...filters, type: v === "all" ? undefined : (v ?? undefined) })}
          >
            <SelectTrigger size="sm" className="w-full" aria-label={t("tickets.filters.type")}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {typeItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="w-full">
          <label className="text-muted-foreground mb-1.5 block text-xs font-medium">
            {t("tickets.filters.class")}
          </label>
          <Select
            items={classItems}
            value={filters.classId ?? "all"}
            onValueChange={(v) => onFiltersChange({ ...filters, classId: v === "all" ? undefined : (v ?? undefined) })}
          >
            <SelectTrigger size="sm" className="w-full" aria-label={t("tickets.filters.class")}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {classItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="space-y-2 p-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="space-y-2 rounded-lg border p-3">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/2" />
              </div>
            ))}
          </div>
        ) : tickets.length === 0 ? (
          <EmptyState
            icon={TicketIcon}
            title={t("tickets.empty.title")}
            description={t("tickets.empty.description")}
            className="h-full justify-center"
          />
        ) : (
          <ul className="space-y-1 p-2">
            {tickets.map((ticket) => {
              // The ball is in the viewer's court when the other side spoke
              // last: creators wait on "answered", handlers wait on "open".
              const isCreator = ticket.user_id === currentUserId
              const waitingOnYou =
                (isCreator && ticket.status === "answered") || (!isCreator && ticket.status === "open")
              const selected = ticket.id === selectedId
              return (
                <li key={ticket.id}>
                  <button
                    type="button"
                    onClick={() => ticket.id && onSelect(ticket.id)}
                    className={cn(
                      "w-full rounded-lg border border-transparent p-3 text-start transition-colors",
                      "hover:bg-muted/60",
                      selected && "border-border bg-muted",
                      waitingOnYou && "border-s-2 border-s-primary bg-primary/5"
                    )}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <span className="min-w-0 truncate text-sm font-medium">{ticket.title}</span>
                      <span className="text-muted-foreground shrink-0 text-xs">
                        {formatRelativeTime(ticket.updated_at, i18n.language)}
                      </span>
                    </div>
                    <div className="mt-1 flex items-center gap-2">
                      <TicketStatusBadge status={ticket.status} />
                      <span className="text-muted-foreground min-w-0 truncate text-xs">
                        {ticket.class?.name ?? ""}
                        {!isCreator && ticket.user?.name ? ` · ${ticket.user.name}` : ""}
                      </span>
                    </div>
                    {waitingOnYou && (
                      <div className="text-primary mt-1 text-xs font-medium">{t("tickets.thread.waitingOnYou")}</div>
                    )}
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </div>
    </div>
  )
}

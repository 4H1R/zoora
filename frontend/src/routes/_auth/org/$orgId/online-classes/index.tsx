import { createFileRoute } from "@tanstack/react-router"
import { ChevronLeft, ChevronRight, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetLiveRooms } from "@/api/live-sessions/live-sessions"
import { ColumnsToggle } from "@/components/data-table/columns-toggle"
import { DataTable } from "@/components/data-table/data-table"
import { LiveRoomCard, LiveRoomCardSkeleton } from "@/components/live-room-card"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { ViewModeToggle, type ViewMode } from "@/components/view-mode-toggle"
import { useOrgGuard } from "@/lib/access"
import { useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useLiveRoomColumns } from "./-columns"

// Backend page size for GET /live-rooms (domain.DefaultPageSize).
const PAGE_SIZE = 20

// statusTab maps the segmented control value to the backend `status` filter.
// "all" => no filter; the rest map 1:1 to LiveRoom.status values.
const STATUS_TABS = ["all", "active", "created", "finished"] as const
type StatusTab = (typeof STATUS_TABS)[number]

const onlineClassesSearchSchema = z.object({
  status: z.enum(STATUS_TABS).optional().default("active"),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
})

export const Route = createFileRoute("/_auth/org/$orgId/online-classes/")({
  head: () => orgHead("org.nav.onlineClasses"),
  validateSearch: onlineClassesSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { status, order_by, order_dir, page } = Route.useSearch()
  const navigate = Route.useNavigate()
  const allowed = useOrgGuard(["live_sessions:view", "live_sessions:view_any"])

  const currentPage = page ?? 1
  const activeTab: StatusTab = status ?? "active"

  const { data, isPending } = useGetLiveRooms(
    {
      status: activeTab === "all" ? undefined : activeTab,
      // URL drives sort; default to the soonest-scheduled-first ordering.
      order_by: order_by || "scheduled_start_time",
      order_dir: order_dir || "desc",
      page: currentPage,
    },
    {
      query: {
        enabled: allowed,
        // Poll so a room going active/finished flips without a manual refresh.
        refetchInterval: 20_000,
        refetchOnWindowFocus: true,
      },
    }
  )

  const roomsData = (data?.status === 200 && data.data.data) || undefined
  const rooms = roomsData?.items ?? []
  const total = roomsData?.total ?? 0

  const [viewMode, setViewMode] = useState<ViewMode>("grid")
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useLiveRoomColumns()
  const table = useAdminTable({ data: rooms, columns, rowCount: total, sorting })

  const renderContent = () => {
    if (viewMode === "table") {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={isPending}
              emptyTitle={t("onlineClassesPage.noResults")}
              emptyHint={t("onlineClassesPage.noResultsHint")}
            />
          </div>
        </Card>
      )
    }

    if (isPending) {
      return (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 8 }, (_, i) => (
            <LiveRoomCardSkeleton key={i} />
          ))}
        </div>
      )
    }

    if (rooms.length === 0) {
      return (
        <EmptyState
          icon={VideoIcon}
          title={t("onlineClassesPage.noResults")}
          description={t("onlineClassesPage.noResultsHint")}
        />
      )
    }

    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {rooms.map((room) => (
          <LiveRoomCard key={room.id} room={room} />
        ))}
      </div>
    )
  }

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("onlineClassesPage.title")} />

      <div className="flex flex-row items-center justify-between gap-3">
        <ToggleGroup
          value={[activeTab]}
          onValueChange={(values) => {
            const next = values.find((v) => v !== activeTab) as StatusTab | undefined
            if (next) navigate({ search: (prev) => ({ ...prev, status: next, page: 1 }) })
          }}
          className="border-border rounded-lg border"
        >
          <ToggleGroupItem value="all" className="px-3 text-xs">
            {t("onlineClassesPage.tabs.all")}
          </ToggleGroupItem>
          <ToggleGroupItem value="active" className="px-3 text-xs">
            {t("onlineClassesPage.tabs.live")}
          </ToggleGroupItem>
          <ToggleGroupItem value="created" className="px-3 text-xs">
            {t("onlineClassesPage.tabs.notStarted")}
          </ToggleGroupItem>
          <ToggleGroupItem value="finished" className="px-3 text-xs">
            {t("onlineClassesPage.tabs.finished")}
          </ToggleGroupItem>
        </ToggleGroup>

        <div className="flex items-center gap-2">
          {viewMode === "table" && (
            <ColumnsToggle
              table={table}
              columnsLabel={t("onlineClassesPage.toolbar.columns")}
              toggleColumnsLabel={t("onlineClassesPage.toolbar.toggleColumns")}
            />
          )}
          <ViewModeToggle value={viewMode} onChange={setViewMode} />
        </div>
      </div>

      {renderContent()}

      {total > PAGE_SIZE && (
        <Pagination
          page={currentPage}
          total={total}
          onPageChange={(p) => navigate({ search: (prev) => ({ ...prev, page: p }) })}
        />
      )}
    </div>
  )
}

function Pagination({
  page,
  total,
  onPageChange,
}: {
  page: number
  total: number
  onPageChange: (page: number) => void
}) {
  const { t } = useTranslation()
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className="text-muted-foreground flex items-center justify-end gap-2 text-xs">
      <span>{t("pagination.pageOf", { page, total: totalPages })}</span>
      <Button
        size="xs"
        variant="outline"
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
        aria-label={t("pagination.previous")}
      >
        <ChevronLeft className="size-3.5 rtl:rotate-180" />
      </Button>
      <Button
        size="xs"
        variant="outline"
        disabled={page >= totalPages}
        onClick={() => onPageChange(page + 1)}
        aria-label={t("pagination.next")}
      >
        <ChevronRight className="size-3.5 rtl:rotate-180" />
      </Button>
    </div>
  )
}

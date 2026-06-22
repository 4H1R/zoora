import type { NavFn } from "@/lib/data-table"

import { createFileRoute } from "@tanstack/react-router"
import { ChevronLeft, ChevronRight, SearchIcon, VideoIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"
import { z } from "zod"

import { useGetLiveRooms } from "@/api/live-sessions/live-sessions"
import { LiveRoomCard, LiveRoomCardSkeleton } from "@/components/live-room-card"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

// Backend page size for GET /live-rooms (domain.DefaultPageSize).
const PAGE_SIZE = 20

// statusTab maps the segmented control value to the backend `status` filter.
// "all" => no filter; the rest map 1:1 to LiveRoom.status values.
const STATUS_TABS = ["all", "active", "created", "finished"] as const
type StatusTab = (typeof STATUS_TABS)[number]

const onlineClassesSearchSchema = z.object({
  search: z.string().optional(),
  status: z.enum(STATUS_TABS).optional().default("active"),
  page: z.number().int().positive().optional().default(1),
})

export const Route = createFileRoute("/_auth/org/$orgId/online-classes/")({
  head: () => orgHead("org.nav.onlineClasses"),
  validateSearch: onlineClassesSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { search, status, page } = Route.useSearch()
  const navigate = Route.useNavigate() as unknown as NavFn
  const allowed = useOrgGuard(["live_sessions:view", "live_sessions:view_any"])

  const currentPage = page ?? 1
  const activeTab: StatusTab = status ?? "active"

  const { data, isPending } = useGetLiveRooms(
    {
      search: search || undefined,
      status: activeTab === "all" ? undefined : activeTab,
      order_by: "scheduled_start_time",
      order_dir: "desc",
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

  const [localSearch, setLocalSearch] = useState(search ?? "")
  const [debouncedSearch] = useDebounce(localSearch, 300)

  useEffect(() => {
    navigate({ search: (prev) => ({ ...prev, search: debouncedSearch || undefined, page: 1 }) })
  }, [debouncedSearch])

  const renderContent = () => {
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
        <div className="text-muted-foreground flex flex-col items-center gap-2 py-16 text-center">
          <VideoIcon className="size-8 opacity-40" />
          <p className="text-sm font-medium">{t("onlineClassesPage.noResults")}</p>
          <p className="text-xs">{t("onlineClassesPage.noResultsHint")}</p>
        </div>
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

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative max-w-xs flex-1">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2" />
          <Input
            placeholder={t("onlineClassesPage.searchPlaceholder")}
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            className="ps-9"
          />
        </div>

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

import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { PlusIcon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminLiveRooms } from "@/api/admin-livesessions/admin-livesessions"
import { LiveRoomCreateModal } from "@/components/admin/live-rooms/LiveRoomCreateModal"
import { LiveRoomTable } from "@/components/admin/live-rooms/LiveRoomTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/live-rooms/")({
  head: () => adminHead("admin.liveRooms.title"),
  validateSearch: adminSearchSchema,
  component: LiveRoomsPage,
})

function LiveRoomsPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()

  const currentPage = page ?? 1
  const [createOpen, setCreateOpen] = useState(false)

  const { data, isLoading } = useGetAdminLiveRooms({
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const liveRoomsData = (data?.status === 200 && data.data.data) || undefined
  const rooms: LiveRoom[] = liveRoomsData?.items ?? []
  const total = liveRoomsData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <VideoIcon />,
      label: t("admin.liveRooms.stats.total"),
      value: total,
      loading: isLoading,
    },
    {
      icon: <VideoIcon />,
      label: t("admin.liveRooms.stats.active"),
      value: rooms.filter((r) => r.status === "active").length,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.liveRooms.title")}
        actions={
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.liveRooms.newLiveRoom")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <LiveRoomTable rooms={rooms} total={total} isLoading={isLoading} sorting={sorting} />

      <LiveRoomCreateModal open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  )
}

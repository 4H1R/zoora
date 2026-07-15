import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, PlusIcon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminLiveRooms } from "@/api/admin-livesessions/admin-livesessions"
import { useGetClassesId } from "@/api/classes/classes"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { LiveRoomCreateModal } from "@/components/admin/live-rooms/LiveRoomCreateModal"
import { LiveRoomTable } from "@/components/admin/live-rooms/LiveRoomTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/classes/$classId/live-rooms")({
  head: () => adminHead("admin.liveRooms.title"),
  validateSearch: adminSearchSchema,
  component: ClassLiveRoomsPage,
})

function ClassLiveRoomsPage() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1
  const [createOpen, setCreateOpen] = useState(false)

  const sessionId: string | undefined = undefined

  const { data: classData } = useGetClassesId(classId)
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  const { data, isLoading } = useGetAdminLiveRooms(
    {
      class_id: classId,
      class_session_id: sessionId || undefined,
      search: search || undefined,
      page: currentPage,
      order_by: order_by || undefined,
      order_dir: order_dir || undefined,
    },
    { query: { enabled: !!sessionId } }
  )

  const liveRoomsData = (data?.status === 200 && data.data.data) || undefined
  const rooms: LiveRoom[] = sessionId ? (liveRoomsData?.items ?? []) : []
  const total = sessionId ? (liveRoomsData?.total ?? 0) : 0

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
        title={cls?.name ? `${cls.name} · ${t("admin.liveRooms.title")}` : t("admin.liveRooms.title")}
        actions={
          <div className="flex items-center gap-2">
            <Link to="/admin/classes/$classId/sessions" params={{ classId }}>
              <Button variant="outline" size="sm">
                <ArrowLeftIcon data-icon="inline-start" />
                {t("admin.classManagement.backToSessions")}
              </Button>
            </Link>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              <PlusIcon data-icon="inline-start" />
              {t("admin.liveRooms.newLiveRoom")}
            </Button>
          </div>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.liveRooms.filter.class")}</label>
          <ClassPicker value={classId} onChange={() => {}} disabled />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.liveRooms.filter.session")}</label>
          <SessionPicker classId={classId} value={sessionId} onChange={() => {}} disabled />
        </div>
      </Card>
      {sessionId ? (
        <LiveRoomTable rooms={rooms} total={total} isLoading={isLoading} sorting={sorting} />
      ) : (
        <Card className="text-muted-foreground p-8 text-center text-sm">
          {t("admin.liveRooms.filter.selectSessionFirst")}
        </Card>
      )}

      <LiveRoomCreateModal open={createOpen} onOpenChange={setCreateOpen} defaultClassId={classId} />
    </div>
  )
}

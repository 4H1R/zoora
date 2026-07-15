import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { FileVideoIcon, PlusIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminOfflines } from "@/api/admin-offlines/admin-offlines"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { OfflineCreateModal } from "@/components/admin/offlines/OfflineCreateModal"
import { OfflineTable } from "@/components/admin/offlines/OfflineTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/offlines/")({
  head: () => adminHead("admin.offlines.title"),
  validateSearch: adminSearchSchema,
  component: OfflinesPage,
})

function OfflinesPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const [classId, setClassId] = useState<string | undefined>(undefined)
  const [sessionId, setSessionId] = useState<string | undefined>(undefined)

  const [formOpen, setFormOpen] = useState(false)
  const [editingRoom, setEditingRoom] = useState<OfflineRoom | null>(null)

  const handleEdit = (room: OfflineRoom) => {
    setEditingRoom(room)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditingRoom(null)
    setFormOpen(true)
  }

  const handleFormOpenChange = (open: boolean) => {
    setFormOpen(open)
    if (!open) setEditingRoom(null)
  }

  const handleClassChange = (id: string) => {
    setClassId(id || undefined)
    setSessionId(undefined)
  }

  const handleClearFilters = () => {
    setClassId(undefined)
    setSessionId(undefined)
  }

  const { data, isLoading } = useGetAdminOfflines({
    class_id: classId || undefined,
    class_session_id: sessionId || undefined,
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const roomsData = (data?.status === 200 && data.data.data) || undefined
  const rooms = roomsData?.items ?? []
  const total = roomsData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <FileVideoIcon />,
      label: t("admin.offlines.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.offlines.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.offlines.newRoom")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.offlines.filter.class")}</label>
          <ClassPicker value={classId} onChange={handleClassChange} />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.offlines.filter.session")}</label>
          <SessionPicker classId={classId} value={sessionId} onChange={(id) => setSessionId(id || undefined)} />
        </div>
        {(classId || sessionId) && (
          <Button variant="outline" size="sm" onClick={handleClearFilters}>
            <XIcon data-icon="inline-start" />
            {t("admin.offlines.filter.clear")}
          </Button>
        )}
      </Card>
      <OfflineTable rooms={rooms} total={total} isLoading={isLoading} sorting={sorting} onEdit={handleEdit} />

      <OfflineCreateModal open={formOpen} onOpenChange={handleFormOpenChange} room={editingRoom} />
    </div>
  )
}

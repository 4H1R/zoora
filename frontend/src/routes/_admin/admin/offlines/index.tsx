import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { FileVideoIcon, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminOfflines } from "@/api/admin-offlines/admin-offlines"
import { OfflineCreateModal } from "@/components/admin/offlines/OfflineCreateModal"
import { OfflineTable } from "@/components/admin/offlines/OfflineTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
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

  const { data, isLoading } = useGetAdminOfflines({
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
      <OfflineTable
        rooms={rooms}
        total={total}
        isLoading={isLoading}
        sorting={sorting}
        onEdit={handleEdit}
      />

      <OfflineCreateModal open={formOpen} onOpenChange={handleFormOpenChange} room={editingRoom} />
    </div>
  )
}

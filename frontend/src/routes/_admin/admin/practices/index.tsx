import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { DumbbellIcon, PlusIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminPractices } from "@/api/admin-practices/admin-practices"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { PracticeCreateModal } from "@/components/admin/practices/PracticeCreateModal"
import { PracticeTable } from "@/components/admin/practices/PracticeTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/practices/")({
  head: () => adminHead("admin.practices.title"),
  validateSearch: adminSearchSchema,
  component: PracticesPage,
})

function PracticesPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()

  const currentPage = page ?? 1

  const [classId, setClassId] = useState<string | undefined>(undefined)
  const [sessionId, setSessionId] = useState<string | undefined>(undefined)

  const [formOpen, setFormOpen] = useState(false)
  const [editingPractice, setEditingPractice] = useState<PracticeRoom | null>(null)

  const handleEdit = (p: PracticeRoom) => {
    setEditingPractice(p)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditingPractice(null)
    setFormOpen(true)
  }

  const handleFormOpenChange = (open: boolean) => {
    setFormOpen(open)
    if (!open) setEditingPractice(null)
  }

  const handleClassChange = (id: string) => {
    setClassId(id || undefined)
    setSessionId(undefined)
  }

  const handleClearFilters = () => {
    setClassId(undefined)
    setSessionId(undefined)
  }

  const { data, isLoading } = useGetAdminPractices({
    class_id: classId || undefined,
    class_session_id: sessionId || undefined,
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const practicesData = (data?.status === 200 && data.data.data) || undefined
  const practices = practicesData?.items ?? []
  const total = practicesData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <DumbbellIcon />,
      label: t("admin.practices.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.practices.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.practices.newPractice")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.practices.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={handleClassChange} />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.practices.filter.session")}
          </label>
          <SessionPicker
            classId={classId}
            value={sessionId}
            onChange={(id) => setSessionId(id || undefined)}
          />
        </div>
        {(classId || sessionId) && (
          <Button variant="outline" size="sm" onClick={handleClearFilters}>
            <XIcon data-icon="inline-start" />
            {t("admin.practices.filter.clear")}
          </Button>
        )}
      </Card>
      <PracticeTable
        practices={practices}
        total={total}
        isLoading={isLoading}
        sorting={sorting}
        onEdit={handleEdit}
      />
      <PracticeCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        practice={editingPractice}
      />
    </div>
  )
}

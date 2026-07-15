import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, DumbbellIcon, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetClassesId } from "@/api/classes/classes"
import { useGetPractices } from "@/api/practices/practices"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { PracticeCreateModal } from "@/components/admin/practices/PracticeCreateModal"
import { PracticeTable } from "@/components/admin/practices/PracticeTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/classes/$classId/practices")({
  head: () => adminHead("admin.practices.title"),
  validateSearch: adminSearchSchema,
  component: ClassPracticesPage,
})

function ClassPracticesPage() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const sessionId: string | undefined = undefined

  const [formOpen, setFormOpen] = useState(false)
  const [editingPractice, setEditingPractice] = useState<PracticeRoom | null>(null)

  const { data: classData } = useGetClassesId(classId)
  const cls = (classData?.status === 200 && classData.data.data) || undefined

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

  const { data, isLoading } = useGetPractices(
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

  const practicesData = (data?.status === 200 && data.data.data) || undefined
  const practices = sessionId ? (practicesData?.items ?? []) : []
  const total = sessionId ? (practicesData?.total ?? 0) : 0

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
        title={cls?.name ? `${cls.name} · ${t("admin.practices.title")}` : t("admin.practices.title")}
        actions={
          <div className="flex items-center gap-2">
            <Link to="/admin/classes/$classId/sessions" params={{ classId }}>
              <Button variant="outline" size="sm">
                <ArrowLeftIcon data-icon="inline-start" />
                {t("admin.classManagement.backToSessions")}
              </Button>
            </Link>
            <Button size="sm" onClick={handleCreate}>
              <PlusIcon data-icon="inline-start" />
              {t("admin.practices.newPractice")}
            </Button>
          </div>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.practices.filter.class")}</label>
          <ClassPicker value={classId} onChange={() => {}} disabled />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">{t("admin.practices.filter.session")}</label>
          <SessionPicker classId={classId} value={sessionId} onChange={() => {}} disabled />
        </div>
      </Card>
      {sessionId ? (
        <PracticeTable
          practices={practices}
          total={total}
          isLoading={isLoading}
          sorting={sorting}
          onEdit={handleEdit}
        />
      ) : (
        <Card className="text-muted-foreground p-8 text-center text-sm">
          {t("admin.practices.filter.selectSessionFirst")}
        </Card>
      )}
      <PracticeCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        practice={editingPractice}
        defaultClassId={classId}
      />
    </div>
  )
}

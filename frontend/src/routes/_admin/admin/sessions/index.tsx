import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { CalendarClockIcon, PlusIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminSessions } from "@/api/admin-classes/admin-classes"
import { ClassPicker } from "@/components/admin/forms/ClassSessionPicker"
import { SessionCreateModal } from "@/components/admin/sessions/SessionCreateModal"
import { SessionTable } from "@/components/admin/sessions/SessionTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/sessions/")({
  head: () => adminHead("admin.sessions.title"),
  validateSearch: adminSearchSchema,
  component: SessionsPage,
})

function SessionsPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const [classId, setClassId] = useState<string | undefined>(undefined)

  const [formOpen, setFormOpen] = useState(false)
  const [editingSession, setEditingSession] = useState<Session | null>(null)

  const handleEdit = (session: Session) => {
    setEditingSession(session)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditingSession(null)
    setFormOpen(true)
  }

  const handleFormOpenChange = (open: boolean) => {
    setFormOpen(open)
    if (!open) setEditingSession(null)
  }

  const handleClearFilters = () => {
    setClassId(undefined)
  }

  const { data, isLoading } = useGetAdminSessions({
    class_id: classId || undefined,
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const sessionsData = (data?.status === 200 && data.data.data) || undefined
  const sessions = sessionsData?.items ?? []
  const total = sessionsData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <CalendarClockIcon />,
      label: t("admin.sessions.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  const editClassId = editingSession?.class_id ?? classId ?? ""

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.sessions.title")}
        actions={
          <Button size="sm" onClick={handleCreate} disabled={!classId}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.sessions.newSession")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.sessions.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={(id) => setClassId(id || undefined)} />
        </div>
        {classId && (
          <Button variant="outline" size="sm" onClick={handleClearFilters}>
            <XIcon data-icon="inline-start" />
            {t("admin.sessions.filter.clear")}
          </Button>
        )}
      </Card>
      <SessionTable
        sessions={sessions}
        total={total}
        isLoading={isLoading}
        sorting={sorting}
        showClass
        onEdit={handleEdit}
      />
      <SessionCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        classId={editClassId}
        session={editingSession}
      />
    </div>
  )
}

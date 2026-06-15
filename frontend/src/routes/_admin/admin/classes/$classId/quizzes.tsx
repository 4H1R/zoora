import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, ClipboardListIcon, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetClassesId } from "@/api/classes/classes"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { QuizCreateModal } from "@/components/admin/quizzes/QuizCreateModal"
import { QuizQuestionsDialog } from "@/components/admin/quizzes/QuizQuestionsDialog"
import { QuizTable } from "@/components/admin/quizzes/QuizTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/classes/$classId/quizzes")({
  head: () => adminHead("admin.quizzes.title"),
  validateSearch: adminSearchSchema,
  component: ClassQuizzesPage,
})

function ClassQuizzesPage() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const sessionId: string | undefined = undefined

  const [formOpen, setFormOpen] = useState(false)
  const [editingQuiz, setEditingQuiz] = useState<Quiz | null>(null)

  const [questionsOpen, setQuestionsOpen] = useState(false)
  const [activeQuiz, setActiveQuiz] = useState<Quiz | null>(null)

  const { data: classData } = useGetClassesId(classId)
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  const handleEdit = (q: Quiz) => {
    setEditingQuiz(q)
    setFormOpen(true)
  }

  const handleCreate = () => {
    setEditingQuiz(null)
    setFormOpen(true)
  }

  const handleFormOpenChange = (open: boolean) => {
    setFormOpen(open)
    if (!open) setEditingQuiz(null)
  }

  const handleManageQuestions = (q: Quiz) => {
    setActiveQuiz(q)
    setQuestionsOpen(true)
  }

  const handleQuestionsOpenChange = (open: boolean) => {
    setQuestionsOpen(open)
    if (!open) setActiveQuiz(null)
  }

  const { data, isLoading } = useGetQuizzes(
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

  const quizzesData = (data?.status === 200 && data.data.data) || undefined
  const quizzes = sessionId ? (quizzesData?.items ?? []) : []
  const total = sessionId ? (quizzesData?.total ?? 0) : 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <ClipboardListIcon />,
      label: t("admin.quizzes.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={
          cls?.name ? `${cls.name} · ${t("admin.quizzes.title")}` : t("admin.quizzes.title")
        }
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
              {t("admin.quizzes.newQuiz")}
            </Button>
          </div>
        }
      />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.quizzes.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={() => {}} disabled />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.quizzes.filter.session")}
          </label>
          <SessionPicker
            classId={classId}
            value={sessionId}
            onChange={() => {}}
            disabled
          />
        </div>
      </Card>
      {sessionId ? (
        <QuizTable
          quizzes={quizzes}
          total={total}
          isLoading={isLoading}
          sorting={sorting}
          onEdit={handleEdit}
          onManageQuestions={handleManageQuestions}
        />
      ) : (
        <Card className="text-muted-foreground p-8 text-center text-sm">
          {t("admin.quizzes.filter.selectSessionFirst")}
        </Card>
      )}
      <QuizCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        quiz={editingQuiz}
        defaultClassId={classId}
      />
      <QuizQuestionsDialog
        open={questionsOpen}
        onOpenChange={handleQuestionsOpenChange}
        quiz={activeQuiz}
      />
    </div>
  )
}

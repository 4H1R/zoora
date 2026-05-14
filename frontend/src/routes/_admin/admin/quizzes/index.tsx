import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { ClipboardListIcon, PlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminQuizzes } from "@/api/admin-quizzes/admin-quizzes"
import { QuizCreateModal } from "@/components/admin/quizzes/QuizCreateModal"
import { QuizQuestionsDialog } from "@/components/admin/quizzes/QuizQuestionsDialog"
import { QuizTable } from "@/components/admin/quizzes/QuizTable"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/quizzes/")({
  head: () => adminHead("admin.quizzes.title"),
  validateSearch: adminSearchSchema,
  component: QuizzesPage,
})

function QuizzesPage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const [formOpen, setFormOpen] = useState(false)
  const [editingQuiz, setEditingQuiz] = useState<Quiz | null>(null)

  const [questionsOpen, setQuestionsOpen] = useState(false)
  const [activeQuiz, setActiveQuiz] = useState<Quiz | null>(null)

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

  const { data, isLoading } = useGetAdminQuizzes({
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const quizzesData = (data?.status === 200 && data.data.data) || undefined
  const quizzes = quizzesData?.items ?? []
  const total = quizzesData?.total ?? 0

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
        title={t("admin.quizzes.title")}
        actions={
          <Button size="sm" onClick={handleCreate}>
            <PlusIcon data-icon="inline-start" />
            {t("admin.quizzes.newQuiz")}
          </Button>
        }
      />
      <StatCards stats={statCards} />
      <QuizTable
        quizzes={quizzes}
        total={total}
        isLoading={isLoading}
        sorting={sorting}
        onEdit={handleEdit}
        onManageQuestions={handleManageQuestions}
      />
      <QuizCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        quiz={editingQuiz}
      />
      <QuizQuestionsDialog
        open={questionsOpen}
        onOpenChange={handleQuestionsOpenChange}
        quiz={activeQuiz}
      />
    </div>
  )
}

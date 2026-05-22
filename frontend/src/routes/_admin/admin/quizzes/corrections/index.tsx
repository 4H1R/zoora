import type {
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { CheckCircle2Icon, CheckSquareIcon, ClockIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminQuizzes } from "@/api/admin-quizzes/admin-quizzes"
import { useGetQuizzesIdSubmissions } from "@/api/quizzes/quizzes"
import { ClassPicker } from "@/components/admin/forms/ClassSessionPicker"
import { CorrectionsTable } from "@/components/admin/quizzes/corrections/CorrectionsTable"
import { GradeSubmissionDialog } from "@/components/admin/quizzes/corrections/GradeSubmissionDialog"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

export const Route = createFileRoute("/_admin/admin/quizzes/corrections/")({
  head: () => adminHead("admin.corrections.title"),
  validateSearch: adminSearchSchema,
  component: CorrectionsPage,
})

const STATUS_OPTIONS = ["submitted", "graded", "in_progress"] as const

function CorrectionsPage() {
  const { t } = useTranslation()
  const { status, order_by, order_dir, page } = Route.useSearch()
  const navigate = Route.useNavigate()
  const currentPage = page ?? 1

  const [classId, setClassId] = useState<string | undefined>(undefined)
  const [quizId, setQuizId] = useState<string | undefined>(undefined)
  const [gradeOpen, setGradeOpen] = useState(false)
  const [activeSubmission, setActiveSubmission] = useState<QuizSubmission | null>(null)

  const { data: quizzesResp, isLoading: quizzesLoading } = useGetAdminQuizzes(
    { class_id: classId },
    { query: { enabled: !!classId } }
  )
  const quizzes: Quiz[] = (quizzesResp?.status === 200 && quizzesResp.data.data?.items) || []

  const { data: subsResp, isLoading: subsLoading } = useGetQuizzesIdSubmissions(
    quizId ?? "",
    {
      status: status || undefined,
      page: currentPage,
      order_by: order_by || undefined,
      order_dir: order_dir || undefined,
    },
    { query: { enabled: !!quizId } }
  )
  const subsData = (subsResp?.status === 200 && subsResp.data.data) || undefined
  const submissions = subsData?.items ?? []
  const total = subsData?.total ?? 0

  const pendingCount = submissions.filter((s) => s.status === "submitted").length
  const gradedCount = submissions.filter((s) => s.status === "graded").length

  const handleClassChange = (id: string) => {
    setClassId(id || undefined)
    setQuizId(undefined)
  }

  const handleClear = () => {
    setClassId(undefined)
    setQuizId(undefined)
    navigate({
      search: (prev) => ({ ...prev, status: undefined, page: 1 }),
    })
  }

  const handleGrade = (s: QuizSubmission) => {
    setActiveSubmission(s)
    setGradeOpen(true)
  }

  const handleStatusChange = (value: string | null) => {
    navigate({
      search: (prev) => ({
        ...prev,
        status: !value || value === "all" ? undefined : value,
        page: 1,
      }),
    })
  }

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <CheckSquareIcon />,
      label: t("admin.corrections.stats.total"),
      value: total,
      loading: subsLoading,
    },
    {
      icon: <ClockIcon />,
      label: t("admin.corrections.stats.pending"),
      value: pendingCount,
      loading: subsLoading,
    },
    {
      icon: <CheckCircle2Icon />,
      label: t("admin.corrections.stats.graded"),
      value: gradedCount,
      loading: subsLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.corrections.title")} />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.corrections.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={handleClassChange} />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.corrections.filter.quiz")}
          </label>
          <Select
            value={quizId ?? ""}
            onValueChange={(v) => setQuizId(v || undefined)}
            disabled={!classId || quizzesLoading}
          >
            <SelectTrigger className="w-full">
              <SelectValue
                placeholder={
                  classId
                    ? t("admin.corrections.filter.quizPlaceholder")
                    : t("admin.corrections.filter.selectClassFirst")
                }
              />
            </SelectTrigger>
            <SelectContent>
              {quizzes.map((q) => (
                <SelectItem key={q.id} value={q.id ?? ""}>
                  {q.title}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.corrections.filter.status")}
          </label>
          <Select value={status ?? "all"} onValueChange={handleStatusChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.corrections.filter.allStatuses")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("admin.corrections.filter.allStatuses")}</SelectItem>
              {STATUS_OPTIONS.map((s) => (
                <SelectItem key={s} value={s}>
                  {t(`admin.corrections.statuses.${s}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {(classId || quizId || status) && (
          <Button variant="outline" size="sm" onClick={handleClear}>
            <XIcon data-icon="inline-start" />
            {t("admin.corrections.filter.clear")}
          </Button>
        )}
      </Card>

      {quizId ? (
        <CorrectionsTable
          submissions={submissions}
          total={total}
          isLoading={subsLoading}
          sorting={sorting}
          onGrade={handleGrade}
        />
      ) : (
        <Card className="text-muted-foreground flex flex-col items-center gap-3 p-8 text-center text-sm">
          <CheckSquareIcon className="size-8 opacity-40" />
          {classId
            ? t("admin.corrections.filter.selectQuizFirst")
            : t("admin.corrections.filter.selectClassFirst")}
        </Card>
      )}

      <GradeSubmissionDialog
        open={gradeOpen}
        onOpenChange={(open) => {
          setGradeOpen(open)
          if (!open) setActiveSubmission(null)
        }}
        submission={activeSubmission}
        quizId={quizId}
      />
    </div>
  )
}

import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { Link, useParams } from "@tanstack/react-router"
import {
  CalendarClockIcon,
  ClipboardListIcon,
  ClockIcon,
  ListChecksIcon,
  LockKeyholeIcon,
  PencilIcon,
  PlayIcon,
  PlusIcon,
  ShuffleIcon,
  Trash2Icon,
  TrophyIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuizzesQueryKey,
  useDeleteQuizzesId,
  useGetQuizzes,
} from "@/api/quizzes/quizzes"
import { QuizQuestionsDialog } from "@/components/admin/quizzes/QuizQuestionsDialog"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useCanSelfOr } from "@/lib/access"
import { formatSessionDate } from "@/lib/session-status"

import { QuizFormDialog } from "./QuizFormDialog"
import { useQuizPermissions } from "./use-quiz-permissions"

interface QuizCardProps {
  quiz: Quiz
  index: number
  orgId: string
  classSessionId: string
  onEdit: (q: Quiz) => void
  onManageQuestions: (q: Quiz) => void
  onDelete: (q: Quiz) => void
}

function QuizCard({ quiz, index, orgId, classSessionId, onEdit, onManageQuestions, onDelete }: QuizCardProps) {
  const { t, i18n } = useTranslation()
  const canEdit = useCanSelfOr("quizzes:update", "quizzes:update_any", quiz.user_id)
  const canDelete = useCanSelfOr("quizzes:delete", "quizzes:delete_any", quiz.user_id)
  const tileNumber = String(index + 1).padStart(2, "0")
  const createdStr = formatSessionDate(quiz.created_at, i18n.language, "short")

  return (
    <div className="group/quiz bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/quiz:opacity-100"
      />
      <div className="flex items-start justify-between gap-3">
        <div className="bg-muted text-foreground flex size-10 items-center justify-center rounded-xl">
          <ClipboardListIcon className="size-5" />
        </div>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-2">
        <Eyebrow>{t("org.session.quizzes.cardEyebrow")}</Eyebrow>
        <h3 className="line-clamp-2 text-xl leading-snug font-semibold tracking-tight text-balance">
          {quiz.title ?? "—"}
        </h3>
        {quiz.description ? (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">{quiz.description}</p>
        ) : null}
        {(quiz.no_back_navigation || quiz.shuffle_questions) ? (
          <div className="mt-1 flex flex-wrap items-center gap-1.5">
            {quiz.no_back_navigation ? (
              <span
                className="border-foreground/15 text-muted-foreground inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider"
                title={t("org.session.quizzes.flags.noBackNavigation")}
              >
                <LockKeyholeIcon className="size-3" />
                {t("org.session.quizzes.flags.noBackShort")}
              </span>
            ) : null}
            {quiz.shuffle_questions ? (
              <span
                className="border-foreground/15 text-muted-foreground inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider"
                title={t("org.session.quizzes.flags.shuffleQuestions")}
              >
                <ShuffleIcon className="size-3" />
                {t("org.session.quizzes.flags.shuffleShort")}
              </span>
            ) : null}
          </div>
        ) : null}
      </div>

      <div className="border-foreground/10 grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.quizzes.duration")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-sm tabular-nums">
            <ClockIcon className="size-3.5" />
            {quiz.duration_minutes ?? 0} {t("org.session.quizzes.minutesShort")}
          </span>
        </div>
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.quizzes.totalScore")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-sm tabular-nums">
            <TrophyIcon className="size-3.5" />
            {(quiz.total_score ?? 0).toFixed(2)}
          </span>
        </div>
      </div>

      <div className="border-foreground/10 mt-auto flex items-center justify-between gap-2 border-t border-dashed pt-3">
        <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs">
          <CalendarClockIcon className="size-3.5" />
          {createdStr}
        </span>
        <div className="flex items-center gap-1.5">
          {(canEdit || canDelete) ? (
            <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/quiz:opacity-100">
              {canEdit ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  title={t("org.session.quizzes.actions.manageQuestions")}
                  onClick={() => onManageQuestions(quiz)}
                >
                  <ListChecksIcon />
                </Button>
              ) : null}
              {canEdit ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  title={t("org.session.quizzes.actions.edit")}
                  onClick={() => onEdit(quiz)}
                >
                  <PencilIcon />
                </Button>
              ) : null}
              {canDelete ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  title={t("org.session.quizzes.actions.delete")}
                  onClick={() => onDelete(quiz)}
                >
                  <Trash2Icon />
                </Button>
              ) : null}
            </div>
          ) : null}
          {quiz.id && orgId ? (
            <Button
              size="sm"
              render={
                <Link
                  to="/org/$orgId/classes/classsessions/$classSessionId/quizzes/$quizId/take"
                  params={{ orgId, classSessionId, quizId: quiz.id }}
                />
              }
            >
              <PlayIcon className="size-3.5" />
              {t("org.session.quizzes.actions.take")}
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function QuizCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex flex-col gap-5 rounded-2xl p-5 ring-1">
      <div className="flex items-center justify-between">
        <Skeleton className="size-10 rounded-xl" />
        <Skeleton className="h-3 w-8" />
      </div>
      <div className="flex flex-col gap-2">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-6 w-4/5" />
        <Skeleton className="h-3 w-3/5" />
      </div>
      <div className="grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <Skeleton className="h-8 w-20" />
        <Skeleton className="h-8 w-20" />
      </div>
    </div>
  )
}

interface EmptyStateProps {
  canCreate: boolean
  onCreate: () => void
}

function EmptyState({ canCreate, onCreate }: EmptyStateProps) {
  const { t } = useTranslation()
  return (
    <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
      <ClipboardListIcon className="text-muted-foreground size-8" />
      <h3 className="text-foreground text-lg font-semibold tracking-tight">
        {t("org.session.quizzes.emptyTitle")}
      </h3>
      <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
        {t("org.session.quizzes.emptyHint")}
      </p>
      {canCreate ? (
        <Button className="mt-2" onClick={onCreate}>
          <PlusIcon className="size-4" />
          {t("org.session.quizzes.newQuiz")}
        </Button>
      ) : null}
    </div>
  )
}

interface QuizzesSectionProps {
  classId: string
  classSessionId: string
}

export function QuizzesSection({ classId, classSessionId }: QuizzesSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canView, canCreate } = useQuizPermissions()
  const { orgId } = useParams({ strict: false }) as { orgId?: string }

  const quizzesQuery = useGetQuizzes(
    { class_session_id: classSessionId },
    { query: { enabled: canView } }
  )
  const quizzes =
    (quizzesQuery.data?.status === 200 && quizzesQuery.data.data.data?.items) || []

  const [formOpen, setFormOpen] = useState(false)
  const [editingQuiz, setEditingQuiz] = useState<Quiz | null>(null)
  const [questionsOpen, setQuestionsOpen] = useState(false)
  const [questionsQuiz, setQuestionsQuiz] = useState<Quiz | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingQuiz, setDeletingQuiz] = useState<Quiz | null>(null)

  const deleteMutation = useDeleteQuizzesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.quizzes.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
        setDeleteOpen(false)
        setDeletingQuiz(null)
      },
    },
  })

  const openCreate = () => {
    setEditingQuiz(null)
    setFormOpen(true)
  }

  if (!canView) return null

  return (
    <section id="quizzes" className="flex flex-col gap-5 scroll-mt-20">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{t("org.session.quizzes.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.quizzes.title")}</h2>
        </div>
        {canCreate ? (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.session.quizzes.newQuiz")}
          </Button>
        ) : null}
      </div>

      {quizzesQuery.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <QuizCardSkeleton />
          <QuizCardSkeleton />
          <QuizCardSkeleton />
        </div>
      ) : quizzes.length === 0 ? (
        <EmptyState canCreate={canCreate} onCreate={openCreate} />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {quizzes.map((q, i) => (
            <QuizCard
              key={q.id}
              quiz={q}
              index={i}
              orgId={orgId ?? ""}
              classSessionId={classSessionId}
              onEdit={(quiz) => {
                setEditingQuiz(quiz)
                setFormOpen(true)
              }}
              onManageQuestions={(quiz) => {
                setQuestionsQuiz(quiz)
                setQuestionsOpen(true)
              }}
              onDelete={(quiz) => {
                setDeletingQuiz(quiz)
                setDeleteOpen(true)
              }}
            />
          ))}
        </div>
      )}

      <QuizFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open)
          if (!open) setEditingQuiz(null)
        }}
        quiz={editingQuiz}
        classId={classId}
        classSessionId={classSessionId}
      />

      <QuizQuestionsDialog
        open={questionsOpen}
        onOpenChange={(open) => {
          setQuestionsOpen(open)
          if (!open) setQuestionsQuiz(null)
        }}
        quiz={questionsQuiz}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          setDeleteOpen(open)
          if (!open) setDeletingQuiz(null)
        }}
        resourceName={deletingQuiz?.title ?? ""}
        onConfirm={() => {
          if (deletingQuiz?.id) deleteMutation.mutate({ id: deletingQuiz.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}

import type {
  GithubCom4H1RZooraInternalDomainGradeAnswerDTO as GradeAnswerDTO,
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
  GithubCom4H1RZooraInternalDomainSubmissionAnswer as SubmissionAnswer,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useGetQuestionBanksQuestionsQuestionId } from "@/api/question-banks/question-banks"
import {
  getGetQuizzesIdSubmissionsQueryKey,
  getGetQuizzesSubmissionsSubmissionIdQueryKey,
  usePostQuizzesSubmissionsSubmissionIdGrade,
} from "@/api/quizzes/quizzes"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

interface GradeSubmissionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  submission: QuizSubmission | null
  quizId?: string
  quizMaxScore?: number
}

export function GradeSubmissionDialog({
  open,
  onOpenChange,
  submission,
  quizId,
  quizMaxScore,
}: GradeSubmissionDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [scores, setScores] = useState<Record<string, number>>({})

  useEffect(() => {
    if (open && submission?.answers) {
      const initial: Record<string, number> = {}
      for (const a of submission.answers) {
        if (a.question_id) initial[a.question_id] = a.earned_score ?? 0
      }
      setScores(initial)
    }
  }, [open, submission])

  const gradeMutation = usePostQuizzesSubmissionsSubmissionIdGrade({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.corrections.dialog.saveSuccess"))
        if (submission?.id) {
          queryClient.invalidateQueries({
            queryKey: getGetQuizzesSubmissionsSubmissionIdQueryKey(submission.id),
          })
        }
        if (quizId) {
          queryClient.invalidateQueries({
            queryKey: getGetQuizzesIdSubmissionsQueryKey(quizId),
          })
        }
        onOpenChange(false)
      },
      onError: () => {
        toast.error(t("admin.corrections.dialog.saveFailed"))
      },
    },
  })

  const handleSave = () => {
    if (!submission?.id) return
    const grades: GradeAnswerDTO[] = Object.entries(scores).map(([question_id, earned_score]) => ({
      question_id,
      earned_score,
    }))
    gradeMutation.mutate({ submissionId: submission.id, data: { grades } })
  }

  const totalScore = Object.values(scores).reduce((sum, v) => sum + (Number.isFinite(v) ? v : 0), 0)

  const answers = submission?.answers ?? []

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100%-2rem)] !max-w-6xl">
        <DialogHeader>
          <DialogTitle className="flex flex-wrap items-center gap-2">
            <span>{t("admin.corrections.dialog.title")}</span>
            {submission?.user?.name && (
              <Badge variant="secondary" className="text-xs font-normal">
                {submission.user.name}
              </Badge>
            )}
            {answers.length > 0 && (
              <Badge variant="outline" className="text-xs font-normal">
                {answers.length} {t("admin.corrections.dialog.answer")}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>
            {t("admin.corrections.dialog.description")}
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-[70vh] space-y-4 overflow-y-auto pe-1">
          {answers.map((answer, idx) => (
            <AnswerRow
              key={answer.question_id ?? idx}
              index={idx}
              answer={answer}
              score={scores[answer.question_id ?? ""] ?? 0}
              onScoreChange={(v) => {
                if (!answer.question_id) return
                setScores((prev) => ({ ...prev, [answer.question_id!]: v }))
              }}
            />
          ))}
          {answers.length === 0 && (
            <div className="text-muted-foreground rounded-md border px-3 py-6 text-center text-sm">
              {t("admin.corrections.dialog.noAnswer")}
            </div>
          )}
        </div>

        <DialogFooter className="items-center justify-between sm:justify-between">
          <div className="text-sm">
            <span className="text-muted-foreground">{t("admin.corrections.dialog.totalScore")}:</span>{" "}
            <span className="font-semibold tabular-nums">{totalScore.toFixed(2)}</span>
            {quizMaxScore != null && quizMaxScore > 0 && (
              <span className="text-muted-foreground ms-1 tabular-nums">
                {t("admin.corrections.scoreOf", { max: quizMaxScore.toFixed(2) })}
              </span>
            )}
          </div>
          <Button onClick={handleSave} disabled={gradeMutation.isPending || !submission?.id}>
            {t("admin.corrections.dialog.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface AnswerRowProps {
  index: number
  answer: SubmissionAnswer
  score: number
  onScoreChange: (v: number) => void
}

function AnswerRow({ index, answer, score, onScoreChange }: AnswerRowProps) {
  const { t } = useTranslation()
  const { data, isLoading } = useGetQuestionBanksQuestionsQuestionId(
    answer.question_id ?? "",
    { query: { enabled: !!answer.question_id } }
  )
  const question: Question | undefined = data?.status === 200 ? data.data.data : undefined

  const maxScore = (question?.options ?? []).reduce(
    (sum, o) => sum + Math.max(0, o.score ?? 0),
    0
  )

  return (
    <div className="bg-card rounded-lg border shadow-sm">
      <div className="bg-muted/40 flex flex-wrap items-center justify-between gap-3 rounded-t-lg border-b px-4 py-2">
        <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
          <span className="bg-background rounded px-1.5 py-0.5 font-semibold tabular-nums">
            #{index + 1}
          </span>
          {question?.type && (
            <Badge variant="secondary" className="text-xs">
              {t(`admin.questions.types.${question.type}`)}
            </Badge>
          )}
          {answer.spent_seconds != null && (
            <span className="tabular-nums">
              {t("admin.corrections.dialog.spent")}: {answer.spent_seconds}
              {t("admin.corrections.dialog.secondsShort")}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <label className="text-muted-foreground text-xs">
            {t("admin.corrections.dialog.score")}
          </label>
          <Input
            type="number"
            step="0.5"
            min={0}
            value={score}
            onChange={(e) => onScoreChange(Number(e.target.value))}
            className="h-8 w-24 text-end tabular-nums"
          />
          {maxScore > 0 && (
            <span className="text-muted-foreground text-xs tabular-nums">/ {maxScore}</span>
          )}
        </div>
      </div>

      <div className="grid gap-4 p-4 lg:grid-cols-2">
        <section>
          <div className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
            {t("admin.questions.title")}
          </div>
          <div className="text-sm whitespace-pre-wrap break-words">
            {isLoading ? (
              <span className="text-muted-foreground">{t("admin.corrections.dialog.loading")}</span>
            ) : question?.text ? (
              question.text
            ) : (
              <span className="text-muted-foreground">{t("admin.corrections.dialog.questionMissing")}</span>
            )}
          </div>
          {question?.options && question.options.length > 0 && (
            <ul className="mt-3 space-y-1">
              {question.options.map((o) => {
                const isCorrect = (o.score ?? 0) > 0
                return (
                  <li
                    key={o.id ?? o.value}
                    className={cn(
                      "rounded-md border px-2 py-1 text-xs",
                      isCorrect
                        ? "border-emerald-500/40 bg-emerald-500/5 text-emerald-700 dark:text-emerald-300"
                        : "text-muted-foreground"
                    )}
                  >
                    <span className="me-2">{o.value}</span>
                    {(o.score ?? 0) !== 0 && (
                      <span className="tabular-nums opacity-70">({o.score})</span>
                    )}
                  </li>
                )
              })}
            </ul>
          )}
        </section>

        <section>
          <div className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
            {t("admin.corrections.dialog.answer")}
          </div>
          <div className="bg-muted/30 min-h-[3rem] rounded-md border px-3 py-2 text-sm">
            {renderAnswer(answer, question, t)}
          </div>
        </section>
      </div>
    </div>
  )
}

function renderAnswer(
  answer: SubmissionAnswer,
  question: Question | undefined,
  t: (k: string) => string
) {
  const selected = answer.selected_option_ids ?? []
  if (selected.length > 0 && question?.options) {
    const labels = selected.map((id) => {
      const opt = question.options?.find((o) => o.id === id)
      return opt?.value ?? id
    })
    return (
      <div>
        <span className="me-1">{t("admin.corrections.dialog.selectedOptions")}:</span>
        <span className="text-foreground">{labels.join(", ")}</span>
      </div>
    )
  }
  if (answer.value && answer.value.length > 0) {
    return <div className="text-foreground whitespace-pre-wrap break-words">{answer.value}</div>
  }
  return <div className="italic">{t("admin.corrections.dialog.noAnswer")}</div>
}

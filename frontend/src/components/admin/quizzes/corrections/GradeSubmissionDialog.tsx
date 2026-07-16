import type {
  GithubCom4H1RZooraInternalDomainSubmissionAntiCheatReport as AntiCheatReport,
  GithubCom4H1RZooraInternalDomainGradeAnswerDTO as GradeAnswerDTO,
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuestionOption as QuestionOption,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
  GithubCom4H1RZooraInternalDomainSubmissionAnswer as SubmissionAnswer,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { CheckCircle2, Circle, Lightbulb, Minus, PenLine, Quote, XCircle, Zap } from "lucide-react"
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
import { formatScore } from "@/lib/score"
import { cn } from "@/lib/utils"

import { ExamIntegrityPanel, fastQuestionIds } from "./exam-integrity"

interface GradeSubmissionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  submission: QuizSubmission | null
  quizId?: string
  quizMaxScore?: number
  report?: AntiCheatReport
}

export function GradeSubmissionDialog({
  open,
  onOpenChange,
  submission,
  quizId,
  quizMaxScore,
  report,
}: GradeSubmissionDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [scores, setScores] = useState<Record<string, number>>({})
  // UI-only: question_ids the teacher has touched this session. Does not affect save payload.
  const [touched, setTouched] = useState<Set<string>>(new Set())

  useEffect(() => {
    if (open && submission?.answers) {
      const initial: Record<string, number> = {}
      for (const a of submission.answers) {
        if (a.question_id) initial[a.question_id] = a.earned_score ?? 0
      }
      setScores(initial)
      setTouched(new Set())
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
  const fastIds = fastQuestionIds(report)
  const gradedCount = answers.filter(
    (a) => a.earned_score != null || (a.question_id ? touched.has(a.question_id) : false)
  ).length
  const progressPct = answers.length > 0 ? (gradedCount / answers.length) * 100 : 0
  const allGraded = answers.length > 0 && gradedCount === answers.length

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100%-2rem)] !max-w-6xl gap-0 p-0">
        <DialogHeader className="px-6 pt-6 pb-4">
          <DialogTitle className="flex flex-wrap items-center gap-2">
            <span>{t("admin.corrections.dialog.title")}</span>
            {submission?.user?.name && (
              <Badge variant="secondary" className="text-xs font-normal">
                {submission.user.name}
              </Badge>
            )}
          </DialogTitle>
          <DialogDescription>{t("admin.corrections.dialog.description")}</DialogDescription>
        </DialogHeader>

        {/* Always-visible metric bar: running total + grading progress */}
        <div className="bg-muted/40 flex flex-wrap items-end justify-between gap-4 border-y px-6 py-3">
          <div>
            <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
              {t("admin.corrections.dialog.totalScore")}
            </div>
            <div className="flex items-baseline gap-1">
              <span className="text-2xl font-semibold tabular-nums">{formatScore(totalScore)}</span>
              {quizMaxScore != null && quizMaxScore > 0 && (
                <span className="text-muted-foreground text-sm tabular-nums">
                  {t("admin.corrections.scoreOf", { max: formatScore(quizMaxScore) })}
                </span>
              )}
            </div>
          </div>
          <div className="min-w-[12rem] flex-1">
            <div className="mb-1 flex items-center justify-between gap-2 text-xs">
              <span className="text-muted-foreground font-medium">{t("admin.corrections.dialog.gradingProgress")}</span>
              <span
                className={cn(
                  "font-medium tabular-nums",
                  allGraded ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground"
                )}
              >
                {t("admin.corrections.dialog.gradedOf", {
                  done: gradedCount,
                  total: answers.length,
                })}
              </span>
            </div>
            <div className="bg-border h-2 overflow-hidden rounded-full">
              <div
                className={cn("h-full rounded-full transition-all", allGraded ? "bg-emerald-500" : "bg-primary")}
                style={{ width: `${progressPct}%` }}
              />
            </div>
          </div>
        </div>

        <ExamIntegrityPanel report={report} submission={submission ?? undefined} />

        <div className="max-h-[65vh] space-y-4 overflow-y-auto px-6 py-4">
          {answers.map((answer, idx) => (
            <AnswerRow
              key={answer.question_id ?? idx}
              index={idx}
              answer={answer}
              fast={!!answer.question_id && fastIds.has(answer.question_id)}
              score={scores[answer.question_id ?? ""] ?? 0}
              onScoreChange={(v) => {
                if (!answer.question_id) return
                setScores((prev) => ({ ...prev, [answer.question_id!]: v }))
                setTouched((prev) => {
                  if (prev.has(answer.question_id!)) return prev
                  const next = new Set(prev)
                  next.add(answer.question_id!)
                  return next
                })
              }}
            />
          ))}
          {answers.length === 0 && (
            <div className="text-muted-foreground rounded-md border px-3 py-6 text-center text-sm">
              {t("admin.corrections.dialog.noAnswer")}
            </div>
          )}
        </div>

        <DialogFooter className="border-t px-6 py-4">
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
  fast?: boolean
}

function AnswerRow({ index, answer, score, onScoreChange, fast }: AnswerRowProps) {
  const { t } = useTranslation()
  const { data, isLoading } = useGetQuestionBanksQuestionsQuestionId(answer.question_id ?? "", {
    query: { enabled: !!answer.question_id },
  })
  const question: Question | undefined = data?.status === 200 ? data.data.data : undefined

  const options = question?.options ?? []
  const maxScore = options.reduce((sum, o) => sum + Math.max(0, o.score ?? 0), 0)
  const selectedIds = new Set(answer.selected_option_ids ?? [])

  const questionType = question?.type
  // Free-text types (short_answer, descriptive) are graded from the student's typed value,
  // not from picking an option. Fall back to option-presence when type is unknown.
  const isChoice = questionType === "choice" || (!questionType && options.length > 0)
  const expectedAnswers = options.filter((o) => (o.score ?? 0) > 0)

  // Verdict tint for the per-question score input.
  const scoreState = maxScore <= 0 ? "neutral" : score >= maxScore ? "full" : score > 0 ? "partial" : "empty"

  return (
    <div className="bg-card rounded-lg border shadow-sm">
      <div className="bg-muted/40 flex flex-wrap items-center justify-between gap-3 rounded-t-lg border-b px-4 py-2.5">
        <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
          <span className="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-full text-xs font-semibold tabular-nums">
            {index + 1}
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
          {fast && (
            <Badge variant="outline" className="border-amber-500/50 bg-amber-500/10 text-amber-700 dark:text-amber-300">
              <Zap className="me-1 size-3" />
              {t("admin.corrections.integrity.fastAnswer")}
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          <label className="text-muted-foreground text-xs font-medium">{t("admin.corrections.dialog.score")}</label>
          <div
            className={cn(
              "flex items-center gap-1 rounded-md border px-1 ps-2 transition-colors",
              scoreState === "full" && "border-emerald-500/50 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300",
              scoreState === "partial" && "border-amber-500/50 bg-amber-500/10 text-amber-700 dark:text-amber-300"
            )}
          >
            <Input
              type="number"
              step="0.5"
              min={0}
              value={score}
              onChange={(e) => onScoreChange(Number(e.target.value))}
              className="h-8 w-16 border-0 bg-transparent px-0 text-end text-sm font-semibold tabular-nums shadow-none focus-visible:ring-0"
            />
            {maxScore > 0 && (
              <span className="text-muted-foreground text-xs tabular-nums">/ {formatScore(maxScore)}</span>
            )}
          </div>
        </div>
      </div>

      <div className="space-y-4 p-4">
        <section>
          <div className="text-muted-foreground mb-1.5 text-xs font-medium tracking-wide uppercase">
            {t("admin.questions.title")}
          </div>
          <div className="text-sm break-words whitespace-pre-wrap">
            {isLoading ? (
              <span className="text-muted-foreground">{t("admin.corrections.dialog.loading")}</span>
            ) : question?.text ? (
              question.text
            ) : (
              <span className="text-muted-foreground">{t("admin.corrections.dialog.questionMissing")}</span>
            )}
          </div>
        </section>

        {isChoice ? (
          <section>
            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
              <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
                {t("admin.corrections.dialog.options")}
              </div>
              <OptionLegend />
            </div>
            <ul className="space-y-1.5">
              {options.map((o) => (
                <OptionRow key={o.id ?? o.value} option={o} chosen={!!o.id && selectedIds.has(o.id)} />
              ))}
            </ul>
          </section>
        ) : (
          <>
            <StudentAnswer value={answer.value} type={questionType} expected={expectedAnswers} />
            {questionType === "descriptive" && (
              <DescriptiveInsights answer={answer} modelAnswer={question?.model_answer} />
            )}
          </>
        )}
      </div>
    </div>
  )
}

// Advisory grading hints for a descriptive answer: the teacher's model answer
// and the char-trigram similarity of the student's text to it. Both are hints
// only — the teacher assigns the score manually.
function DescriptiveInsights({ answer, modelAnswer }: { answer: SubmissionAnswer; modelAnswer?: string }) {
  const { t } = useTranslation()
  const hasSimilarity = answer.similarity_pct != null

  if (!modelAnswer && !hasSimilarity) return null

  return (
    <section className="space-y-3">
      {!!modelAnswer && (
        <div>
          <div className="text-muted-foreground mb-1.5 text-xs font-medium tracking-wide uppercase">
            {t("admin.corrections.dialog.modelAnswer")}
          </div>
          <div className="text-muted-foreground bg-muted/30 max-h-40 overflow-y-auto rounded-md border px-3.5 py-2.5 text-sm leading-relaxed break-words whitespace-pre-wrap">
            {modelAnswer}
          </div>
        </div>
      )}

      {hasSimilarity && (
        <div className="rounded-md border border-amber-500/40 bg-amber-500/5 px-3.5 py-2.5">
          <span className="inline-flex items-center gap-1.5 text-sm font-medium text-amber-700 dark:text-amber-300">
            <Lightbulb className="size-4" />
            {t("admin.corrections.dialog.similarity")}:{" "}
            <span className="tabular-nums">{answer.similarity_pct}%</span>
          </span>
          <p className="text-muted-foreground mt-1 text-xs">{t("admin.corrections.dialog.suggestionHint")}</p>
        </div>
      )}
    </section>
  )
}

function StudentAnswer({ value, type, expected }: { value?: string; type?: string; expected: QuestionOption[] }) {
  const { t } = useTranslation()
  const hasAnswer = !!value && value.trim().length > 0
  const isDescriptive = type === "descriptive"
  const Icon = isDescriptive ? Quote : PenLine

  return (
    <section className="space-y-3">
      <div>
        <div className="text-muted-foreground mb-1.5 flex items-center gap-1.5 text-xs font-medium tracking-wide uppercase">
          <Icon className="size-3.5" />
          {t("admin.corrections.dialog.answer")}
        </div>
        {hasAnswer ? (
          <div
            className={cn(
              "border-s-primary/50 bg-muted/40 rounded-md border border-s-4 px-3.5 py-2.5 break-words whitespace-pre-wrap",
              isDescriptive ? "max-h-72 overflow-y-auto text-sm leading-relaxed" : "text-base font-medium"
            )}
          >
            {value}
          </div>
        ) : (
          <div className="text-muted-foreground bg-muted/20 rounded-md border border-dashed px-3.5 py-3 text-sm italic">
            {t("admin.corrections.dialog.noAnswer")}
          </div>
        )}
      </div>

      {/* Reference for the grader — accepted answers set on the question (short answer only). */}
      {!isDescriptive && expected.length > 0 && (
        <div>
          <div className="text-muted-foreground mb-1.5 text-xs font-medium tracking-wide uppercase">
            {t("admin.corrections.dialog.expectedAnswer")}
          </div>
          <div className="flex flex-wrap gap-1.5">
            {expected.map((o) => (
              <Badge
                key={o.id ?? o.value}
                variant="outline"
                className="border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
              >
                {o.value}
              </Badge>
            ))}
          </div>
        </div>
      )}
    </section>
  )
}

function OptionRow({ option, chosen }: { option: QuestionOption; chosen: boolean }) {
  const { t } = useTranslation()
  const isCorrect = (option.score ?? 0) > 0

  // Two orthogonal signals encoded in one leading icon:
  //   correct + chosen  → filled green check (right answer, picked)
  //   correct + missed  → hollow green circle (right answer, left blank)
  //   wrong   + chosen  → red cross (student's mistake)
  //   wrong   + missed  → faint dash (neutral)
  const Icon = isCorrect ? (chosen ? CheckCircle2 : Circle) : chosen ? XCircle : Minus

  return (
    <li
      className={cn(
        "flex items-start gap-2.5 rounded-md border px-2.5 py-2 text-sm transition-colors",
        // Signal A — correctness (answer key): green start-accent + tint on every correct option.
        isCorrect
          ? "border-s-4 border-emerald-500/30 border-s-emerald-500 bg-emerald-500/5"
          : chosen
            ? // wrong + chosen: flag the mistake in red.
              "border-s-destructive border-destructive/30 bg-destructive/5 border-s-4"
            : "border-border"
      )}
    >
      <Icon
        className={cn(
          "mt-0.5 size-4 shrink-0",
          isCorrect
            ? "text-emerald-600 dark:text-emerald-400"
            : chosen
              ? "text-destructive"
              : "text-muted-foreground/50"
        )}
      />
      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1">
        <span
          className={cn(
            "break-words",
            isCorrect ? "text-emerald-800 dark:text-emerald-200" : chosen ? "text-destructive" : "text-foreground"
          )}
        >
          {option.value}
        </span>

        {/* Signal A label — this option is (part of) the correct answer. */}
        {isCorrect && (
          <Badge
            variant="outline"
            className="border-emerald-500/40 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
          >
            {t("admin.corrections.dialog.correctAnswer")}
            {(option.score ?? 0) !== 0 && <span className="ms-1 tabular-nums opacity-80">+{option.score}</span>}
          </Badge>
        )}

        {/* Signal B — student's choice: separate pill on every selected option, colored by outcome. */}
        {chosen && (
          <Badge
            variant="outline"
            className={cn(
              isCorrect
                ? "border-emerald-500/50 bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                : "border-destructive/50 bg-destructive/15 text-destructive"
            )}
          >
            {t("admin.corrections.dialog.studentChoice")}
          </Badge>
        )}
      </div>
    </li>
  )
}

function OptionLegend() {
  const { t } = useTranslation()
  return (
    <div className="text-muted-foreground flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
      <span className="inline-flex items-center gap-1">
        <CheckCircle2 className="size-3.5 text-emerald-600 dark:text-emerald-400" />
        {t("admin.corrections.dialog.correctAnswer")}
      </span>
      <span className="inline-flex items-center gap-1">
        <span className="border-primary/60 bg-primary/10 rounded-full border px-1.5 py-px text-[10px] font-medium">
          {t("admin.corrections.dialog.studentChoice")}
        </span>
      </span>
      <span className="inline-flex items-center gap-1">
        <XCircle className="text-destructive size-3.5" />
        {t("admin.corrections.dialog.incorrectChoice")}
      </span>
    </div>
  )
}

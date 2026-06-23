import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
  GithubCom4H1RZooraInternalDomainSubmitAnswerDTO as SubmitAnswer,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { ArrowLeftIcon, ArrowRightIcon, FlagIcon, LockKeyholeIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuizzesIdSubmissionsQueryKey,
  usePostQuizzesSubmissionsSubmissionIdSubmit,
} from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Spinner } from "@/components/ui/spinner"
import { useNow } from "@/lib/session-status"

import { DecorativeBackground } from "./decorations"
import { CenterMessage } from "./messages"
import { ProgressDots } from "./progress-dots"
import { QuestionInput } from "./question-input"
import { clearPersistedState, loadPersistedState, savePersistedState } from "./storage"
import { TimerPill } from "./timer-pill"
import type { AnswerState } from "./types"
import { emptyAnswer } from "./types"
import {
  computeDeadline,
  countAnswered,
  questionTypeKey,
  shuffleSeeded,
} from "./utils"

interface PlayAreaProps {
  quiz: Quiz
  room: QuizRoom
  submission: QuizSubmission
  questions: Question[]
  backHref: string
}

export function PlayArea({ quiz, room, submission, questions, backHref }: PlayAreaProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const nowMs = useNow(1000)

  const submissionId = submission.id!
  const deadline = computeDeadline(submission.started_at, quiz.duration_minutes ?? 0, room)
  const remainingSeconds = Math.max(0, Math.floor((deadline - nowMs) / 1000))
  const isExpired = remainingSeconds <= 0

  // Question order — computed once per mount; deterministic shuffle keyed by submission id
  const [orderedQuestionIds] = useState<string[]>(() => {
    const ids = questions.map((q) => q.id!).filter(Boolean)
    return quiz.shuffle_questions ? shuffleSeeded(ids, submissionId) : ids
  })

  const [answers, setAnswers] = useState<Record<string, AnswerState>>(() => {
    const saved = loadPersistedState(submissionId)
    return saved?.answers ?? {}
  })
  const [index, setIndex] = useState<number>(() => {
    const saved = loadPersistedState(submissionId)
    if (saved && saved.index >= 0 && saved.index < orderedQuestionIds.length) return saved.index
    return 0
  })
  const [confirmOpen, setConfirmOpen] = useState(false)

  const submittedRef = useRef(false)

  useEffect(() => {
    savePersistedState(submissionId, { answers, order: orderedQuestionIds, index })
  }, [answers, index, orderedQuestionIds, submissionId])

  const currentQuestionId = orderedQuestionIds[index]

  usePerQuestionTimer(currentQuestionId, setAnswers)

  const submitMutation = usePostQuizzesSubmissionsSubmissionIdSubmit({
    mutation: {
      onSuccess: () => {
        clearPersistedState(submissionId)
        queryClient.invalidateQueries({ queryKey: getGetQuizzesIdSubmissionsQueryKey(quiz.id!) })
        toast.success(t("org.session.quizzes.take.submitted"))
      },
      onError: () => {
        submittedRef.current = false
        toast.error(t("org.session.quizzes.take.submitFailed"))
      },
    },
  })

  function finalize(reason: "manual" | "auto") {
    if (submittedRef.current) return
    submittedRef.current = true
    const payload: SubmitAnswer[] = orderedQuestionIds.map((qid) => {
      const a = answers[qid] ?? emptyAnswer()
      return {
        question_id: qid,
        selected_option_ids: a.selected_option_ids,
        value: a.value,
        spent_seconds: a.spent_seconds,
      }
    })
    submitMutation.mutate({ submissionId, data: { answers: payload } })
    if (reason === "auto") toast.message(t("org.session.quizzes.take.autoSubmitNote"))
  }

  useEffect(() => {
    if (isExpired && submission.status === "in_progress" && !submittedRef.current) {
      finalize("auto")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpired, submission.status])

  useEffect(() => {
    if (submission.status !== "in_progress") return
    const handler = (e: BeforeUnloadEvent) => {
      if (submittedRef.current) return
      e.preventDefault()
      e.returnValue = ""
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [submission.status])

  if (submitMutation.isSuccess || submission.status !== "in_progress") {
    return (
      <div className="relative isolate flex min-h-[60vh] flex-col items-center justify-center gap-4 py-16 text-center">
        <DecorativeBackground />
        <Spinner className="size-6" />
        <Eyebrow>{t("org.session.quizzes.take.submitting")}</Eyebrow>
      </div>
    )
  }

  const currentQuestion = questions.find((q) => q.id === currentQuestionId)
  if (!currentQuestion) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.empty.title")}
        description={t("org.session.quizzes.take.empty.description")}
        backHref={backHref}
      />
    )
  }

  const currentAnswer = answers[currentQuestionId] ?? emptyAnswer()
  const total = orderedQuestionIds.length
  const isLast = index === total - 1
  const isFirst = index === 0

  function setAnswer(updater: (prev: AnswerState) => AnswerState) {
    setAnswers((prev) => ({ ...prev, [currentQuestionId]: updater(prev[currentQuestionId] ?? emptyAnswer()) }))
  }

  return (
    <div className="relative isolate flex flex-col gap-8 pb-32 pt-6">
      <DecorativeBackground />

      <div className="sticky top-0 z-10 -mx-4 backdrop-blur md:-mx-6">
        <div className="bg-background/70 border-foreground/10 mx-4 flex items-center justify-between gap-4 rounded-2xl border px-4 py-3 md:mx-6 md:px-6">
          <div className="flex items-center gap-3">
            <span className="text-muted-foreground font-mono text-xs tracking-[0.25em] uppercase">
              {t("org.session.quizzes.take.progress", { current: index + 1, total })}
            </span>
            <ProgressDots order={orderedQuestionIds} index={index} answers={answers} />
          </div>
          <TimerPill remainingSeconds={remainingSeconds} />
        </div>
      </div>

      <article className="bg-card ring-foreground/10 relative flex flex-col gap-6 overflow-hidden rounded-3xl p-6 ring-1 md:p-10">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/6%,transparent_60%)]"
        />
        <div className="flex items-center justify-between gap-3">
          <Eyebrow>
            {t("org.session.quizzes.take.questionN", { n: index + 1 })} ·{" "}
            {t(`org.session.quizzes.take.types.${questionTypeKey(currentQuestion.type)}`)}
          </Eyebrow>
          <span className="text-muted-foreground font-mono text-[10px] tracking-[0.25em]">
            /{String(index + 1).padStart(2, "0")}
          </span>
        </div>

        <h2 className="max-w-3xl text-2xl leading-snug font-semibold tracking-tight text-balance md:text-3xl">
          {currentQuestion.text}
        </h2>

        <QuestionInput question={currentQuestion} answer={currentAnswer} onChange={setAnswer} />
      </article>

      <div className="flex flex-wrap items-center justify-between gap-3">
        {quiz.no_back_navigation ? (
          <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase">
            <LockKeyholeIcon className="size-3.5" />
            {t("org.session.quizzes.flags.noBackNavigation")}
          </span>
        ) : (
          <Button
            variant="outline"
            onClick={() => setIndex((i) => Math.max(0, i - 1))}
            disabled={isFirst}
          >
            <ArrowLeftIcon className="size-4" />
            {t("org.session.quizzes.take.previous")}
          </Button>
        )}

        <div className="ms-auto flex items-center gap-2">
          {isLast ? (
            <Button onClick={() => setConfirmOpen(true)}>
              <FlagIcon className="size-4" />
              {t("org.session.quizzes.take.finish")}
            </Button>
          ) : (
            <Button onClick={() => setIndex((i) => Math.min(total - 1, i + 1))}>
              {t("org.session.quizzes.take.next")}
              <ArrowRightIcon className="size-4" />
            </Button>
          )}
        </div>
      </div>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("org.session.quizzes.take.confirm.title")}</DialogTitle>
            <DialogDescription>
              {t("org.session.quizzes.take.confirm.description", {
                answered: countAnswered(answers, orderedQuestionIds, questions),
                total,
              })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              {t("common.cancel")}
            </Button>
            <Button
              onClick={() => {
                setConfirmOpen(false)
                finalize("manual")
              }}
              disabled={submitMutation.isPending}
            >
              {submitMutation.isPending ? <Spinner className="size-4" /> : <FlagIcon className="size-4" />}
              {t("org.session.quizzes.take.finishConfirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// Per-question timer: tracks dwell time on the active question and commits to
// the answer's spent_seconds. Commits the remaining delta on unmount/switch.
function usePerQuestionTimer(
  currentQuestionId: string | undefined,
  setAnswers: (updater: (prev: Record<string, AnswerState>) => Record<string, AnswerState>) => void,
) {
  const lastTickRef = useRef<number>(Date.now())

  useEffect(() => {
    if (!currentQuestionId) return
    lastTickRef.current = Date.now()

    function commit() {
      const now = Date.now()
      const delta = Math.max(0, Math.floor((now - lastTickRef.current) / 1000))
      if (delta === 0) return
      lastTickRef.current = now
      setAnswers((prev) => {
        const existing = prev[currentQuestionId!] ?? emptyAnswer()
        return {
          ...prev,
          [currentQuestionId!]: { ...existing, spent_seconds: existing.spent_seconds + delta },
        }
      })
    }

    const interval = setInterval(commit, 1000)
    return () => {
      commit()
      clearInterval(interval)
    }
  }, [currentQuestionId, setAnswers])
}

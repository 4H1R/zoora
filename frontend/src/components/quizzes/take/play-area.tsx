import type { AnswerState } from "./types"
import type { DeviceInfo } from "@/lib/device"
import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
  GithubCom4H1RZooraInternalDomainSubmitAnswerDTO as SubmitAnswer,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { AlertTriangleIcon, ArrowLeftIcon, ArrowRightIcon, FlagIcon, LockKeyholeIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuizzesIdSubmissionsQueryKey,
  usePostQuizzesSubmissionsSubmissionIdAnswers,
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
import { detectDevice } from "@/lib/device"
import { useNow } from "@/lib/session-status"

import { DecorativeBackground } from "./decorations"
import { CenterMessage } from "./messages"
import { ProgressRail } from "./progress-dots"
import { QuestionInput } from "./question-input"
import { clearPersistedState, loadPersistedState, savePersistedState } from "./storage"
import { SystemImage } from "./system-image"
import { TimerPill } from "./timer-pill"
import { emptyAnswer } from "./types"
import { useExamLockdown } from "./use-exam-lockdown"
import { useTabVisibility } from "./use-tab-visibility"
import { computeDeadline, countAnswered, penaltyText, questionTypeKey, shuffleSeeded } from "./utils"

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
  const startedMs = submission.started_at ? Date.parse(submission.started_at) : deadline
  const totalSeconds = Math.max(1, Math.round((deadline - startedMs) / 1000))
  const remainingSeconds = Math.max(0, Math.floor((deadline - nowMs) / 1000))
  const isExpired = remainingSeconds <= 0

  // Question order — computed once per mount; deterministic shuffle keyed by submission id
  const [orderedQuestionIds] = useState<string[]>(() => {
    const ids = questions.map((q) => q.id!).filter(Boolean)
    return quiz.shuffle_questions ? shuffleSeeded(ids, submissionId) : ids
  })

  const [persistedTabHidden] = useState(() => loadPersistedState(submissionId)?.tabHidden)
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
  const lastBlockToastRef = useRef(0)

  const currentQuestionId = orderedQuestionIds[index]

  usePerQuestionTimer(currentQuestionId, setAnswers)

  // track_tab_switches: count hidden-tab events + accumulated hidden seconds,
  // warn (non-blocking) on return, and send the totals on final submit. Seeded
  // from and persisted to localStorage so a mid-quiz refresh resumes the
  // counters instead of resetting them to zero.
  const tabVisibility = useTabVisibility(Boolean(quiz.track_tab_switches), {
    initial: persistedTabHidden,
    onReturn: (count) => toast.warning(t("org.session.quizzes.take.antiCheat.tabWarning", { count })),
    onChange: (stats) =>
      savePersistedState(submissionId, { answers, order: orderedQuestionIds, index, tabHidden: stats }),
  })

  useEffect(() => {
    savePersistedState(submissionId, {
      answers,
      order: orderedQuestionIds,
      index,
      tabHidden: quiz.track_tab_switches ? tabVisibility.read() : undefined,
    })
  }, [answers, index, orderedQuestionIds, submissionId, quiz.track_tab_switches, tabVisibility])

  // Incremental server-side save: commit each answer as the student advances so
  // progress survives a crash/late submit, no_back_navigation is enforced
  // server-side (an answered question locks), and tab-hidden totals accumulate
  // via the backend's monotonic max-merge instead of arriving only once at
  // final submit. Best-effort — the final submit is the authoritative backstop,
  // so failures (network, no-back 409) are swallowed and not retried.
  const answersRef = useRef(answers)
  answersRef.current = answers
  const saveMutation = usePostQuizzesSubmissionsSubmissionIdAnswers()
  const savedSigRef = useRef<Record<string, string>>({})

  function flushAnswer(questionId: string | undefined) {
    if (!questionId || submittedRef.current) return
    const a = answersRef.current[questionId]
    if (!a || (a.selected_option_ids.length === 0 && a.value.trim().length === 0)) return
    const tab = quiz.track_tab_switches ? tabVisibility.read() : null
    const sig = JSON.stringify([a.selected_option_ids, a.value, a.spent_seconds, tab])
    if (savedSigRef.current[questionId] === sig) return
    savedSigRef.current[questionId] = sig
    saveMutation.mutate({
      submissionId,
      data: {
        question_id: questionId,
        selected_option_ids: a.selected_option_ids,
        value: a.value,
        spent_seconds: a.spent_seconds,
        ...(tab ? { tab_hidden_count: tab.count, tab_hidden_seconds: tab.seconds } : {}),
      },
    })
  }

  // Flush the question the student just left. Reading orderedQuestionIds via ref
  // isn't needed — order is fixed for the mount.
  const prevIndexRef = useRef(index)
  useEffect(() => {
    if (prevIndexRef.current !== index) {
      flushAnswer(orderedQuestionIds[prevIndexRef.current])
      prevIndexRef.current = index
    }
    // flushAnswer reads latest state via refs; deps intentionally limited to index.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [index])

  // disable_copy_paste / disable_right_click_shortcuts: silent preventDefault
  // with a throttled toast so rapid attempts don't spam.
  useExamLockdown({
    disableCopyPaste: Boolean(quiz.disable_copy_paste),
    disableShortcuts: Boolean(quiz.disable_right_click_shortcuts),
    onBlocked: () => {
      const now = Date.now()
      if (now - lastBlockToastRef.current < 2000) return
      lastBlockToastRef.current = now
      toast.message(t("org.session.quizzes.take.antiCheat.blocked"))
    },
  })

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
    const tabStats = quiz.track_tab_switches ? tabVisibility.read() : null
    // Advisory device snapshot — the server pairs this with the raw user-agent
    // header. Best-effort; never block submit if detection throws.
    let device: DeviceInfo | undefined
    try {
      device = detectDevice()
    } catch {
      device = undefined
    }
    submitMutation.mutate({
      submissionId,
      data: {
        answers: payload,
        ...(tabStats ? { tab_hidden_count: tabStats.count, tab_hidden_seconds: tabStats.seconds } : {}),
        ...(device ? { device } : {}),
      },
    })
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
    <div className="relative isolate flex flex-col gap-8 pt-6 pb-32">
      <DecorativeBackground />

      <div className="sticky top-0 z-10 -mx-4 backdrop-blur md:-mx-6">
        <div className="bg-background/70 supports-[backdrop-filter]:bg-background/55 border-foreground/10 mx-4 flex items-center justify-between gap-4 rounded-2xl border px-3 py-2.5 shadow-sm md:mx-6 md:ps-5 md:pe-4">
          <div className="flex items-center gap-3 md:gap-4">
            <div className="flex flex-col gap-1">
              <Eyebrow className="text-[0.6rem] tracking-[0.28em]">
                {t("org.session.quizzes.take.questionLabel")}
              </Eyebrow>
              <div className="flex items-baseline gap-1 font-mono leading-none">
                <span className="text-foreground text-xl font-semibold tabular-nums md:text-2xl">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <span className="text-muted-foreground/50 text-sm tabular-nums">
                  / {String(total).padStart(2, "0")}
                </span>
              </div>
            </div>
            <div className="bg-foreground/10 hidden h-9 w-px sm:block" />
            <ProgressRail order={orderedQuestionIds} index={index} answers={answers} />
            <span className="sr-only">{t("org.session.quizzes.take.progress", { current: index + 1, total })}</span>
          </div>
          <TimerPill remainingSeconds={remainingSeconds} totalSeconds={totalSeconds} />
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

        {quiz.render_as_image && currentQuestion.system_image_media_id ? (
          <SystemImage mediaID={currentQuestion.system_image_media_id} className="max-h-40 w-auto" />
        ) : (
          <h2 className="max-w-3xl text-2xl leading-snug font-semibold tracking-tight text-balance md:text-3xl">
            {currentQuestion.text}
          </h2>
        )}

        <PenaltyBadge question={currentQuestion} />

        <QuestionInput question={currentQuestion} answer={currentAnswer} onChange={setAnswer} />
      </article>

      <div className="flex flex-wrap items-center justify-between gap-3">
        {quiz.no_back_navigation ? (
          <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase">
            <LockKeyholeIcon className="size-3.5" />
            {t("org.session.quizzes.flags.noBackNavigation")}
          </span>
        ) : (
          <Button variant="outline" onClick={() => setIndex((i) => Math.max(0, i - 1))} disabled={isFirst}>
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

// PenaltyBadge surfaces the per-question negative-marking penalty (display-only).
function PenaltyBadge({ question }: { question: Question }) {
  const { t } = useTranslation()
  const text = penaltyText(question.negative_config, t)
  if (!text) return null
  return (
    <span className="border-destructive/30 text-destructive bg-destructive/5 inline-flex w-fit items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium">
      <AlertTriangleIcon className="size-3.5" />
      {text}
    </span>
  )
}

// Per-question timer: tracks dwell time on the active question and commits to
// the answer's spent_seconds. Commits the remaining delta on unmount/switch.
function usePerQuestionTimer(
  currentQuestionId: string | undefined,
  setAnswers: (updater: (prev: Record<string, AnswerState>) => Record<string, AnswerState>) => void
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

import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuestionOption as QOption,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizRule as QuizRule,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
  GithubCom4H1RZooraInternalDomainSubmitAnswerDTO as SubmitAnswer,
} from "@/api/model"

import { useQueries, useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import {
  AlertTriangleIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  CheckCircle2Icon,
  ClipboardListIcon,
  ClockIcon,
  FlagIcon,
  LockKeyholeIcon,
  ShuffleIcon,
  TrophyIcon,
} from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetQuestionBanksIdQuestionsQueryOptions } from "@/api/question-banks/question-banks"
import {
  getGetQuizzesIdSubmissionsQueryKey,
  useGetQuizzesId,
  useGetQuizzesIdRooms,
  useGetQuizzesIdRules,
  useGetQuizzesIdSubmissions,
  usePostQuizzesIdSubmissions,
  usePostQuizzesSubmissionsSubmissionIdSubmit,
} from "@/api/quizzes/quizzes"
import { Eyebrow } from "@/components/eyebrow"
import { useQuizPermissions } from "@/components/org/quizzes/use-quiz-permissions"
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
import { Skeleton } from "@/components/ui/skeleton"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

export const Route = createFileRoute(
  "/_auth/org/$orgId/classes/classsessions/$classSessionId/quizzes/$quizId/take",
)({
  head: () => orgHead("org.session.quizzes.take.headTitle"),
  component: RouteComponent,
})

type AnswerState = {
  selected_option_ids: string[]
  value: string
  spent_seconds: number
}

const STORAGE_PREFIX = "zoora.quiz.take."

function storageKey(submissionId: string) {
  return STORAGE_PREFIX + submissionId
}

function loadLocal(submissionId: string): {
  answers: Record<string, AnswerState>
  order: string[]
  index: number
} | null {
  try {
    const raw = localStorage.getItem(storageKey(submissionId))
    if (!raw) return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function saveLocal(
  submissionId: string,
  state: { answers: Record<string, AnswerState>; order: string[]; index: number },
) {
  try {
    localStorage.setItem(storageKey(submissionId), JSON.stringify(state))
  } catch {
    // ignore quota
  }
}

function clearLocal(submissionId: string) {
  try {
    localStorage.removeItem(storageKey(submissionId))
  } catch {
    // ignore
  }
}

function shuffleSeeded<T>(arr: T[], seed: string): T[] {
  // deterministic per (seed) shuffle so reload preserves order
  let h = 2166136261
  for (let i = 0; i < seed.length; i++) {
    h = (h ^ seed.charCodeAt(i)) * 16777619
  }
  const out = arr.slice()
  for (let i = out.length - 1; i > 0; i--) {
    h = (h * 1664525 + 1013904223) | 0
    const j = Math.abs(h) % (i + 1)
    ;[out[i], out[j]] = [out[j], out[i]]
  }
  return out
}

function formatClock(totalSeconds: number) {
  const t = Math.max(0, Math.floor(totalSeconds))
  const h = Math.floor(t / 3600)
  const m = Math.floor((t % 3600) / 60)
  const s = t % 60
  const pad = (n: number) => n.toString().padStart(2, "0")
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
}

function DecorativeBackground() {
  return (
    <>
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/10%,transparent_55%)]"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 opacity-[0.05] [background-image:linear-gradient(var(--color-foreground)_1px,transparent_1px),linear-gradient(90deg,var(--color-foreground)_1px,transparent_1px)] [background-size:48px_48px] [mask-image:radial-gradient(ellipse_at_top,black,transparent_70%)]"
      />
    </>
  )
}

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId, classSessionId, quizId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView } = useQuizPermissions()

  const sessionHref = `/org/${orgId}/classes/classsessions/${classSessionId}`

  const quizQ = useGetQuizzesId(quizId)
  const quiz = (quizQ.data?.status === 200 && quizQ.data.data.data) || undefined

  const roomsQ = useGetQuizzesIdRooms(quizId, undefined, {
    query: { enabled: !!quiz },
  })
  const rooms = (roomsQ.data?.status === 200 && roomsQ.data.data.data?.items) || []

  const rulesQ = useGetQuizzesIdRules(quizId, undefined, {
    query: { enabled: !!quiz },
  })
  const rules = (rulesQ.data?.status === 200 && rulesQ.data.data.data?.items) || []

  const submissionsQ = useGetQuizzesIdSubmissions(quizId, undefined, {
    query: { enabled: !!quiz },
  })
  const submissions = (submissionsQ.data?.status === 200 && submissionsQ.data.data.data?.items) || []
  const myInProgress = submissions.find((s) => s.status === "in_progress")
  const mySubmitted = submissions.find((s) => s.status !== "in_progress")

  if (!allowed) return null
  if (!canView) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={sessionHref}
      />
    )
  }

  if (quizQ.isPending || roomsQ.isPending || rulesQ.isPending || submissionsQ.isPending) {
    return <LoadingScreen />
  }

  if (quizQ.isError || !quiz) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.notFound.title")}
        description={t("org.session.quizzes.take.notFound.description")}
        backHref={sessionHref}
      />
    )
  }

  const room = pickRoomForSession(rooms, classSessionId)

  if (mySubmitted) {
    return <ResultScreen quiz={quiz} submission={mySubmitted} backHref={sessionHref} />
  }

  return (
    <QuizRunner
      quiz={quiz}
      room={room}
      rules={rules}
      existingSubmission={myInProgress}
      backHref={sessionHref}
    />
  )
}

function pickRoomForSession(rooms: QuizRoom[], classSessionId: string): QuizRoom | undefined {
  const matched = rooms.find((r) => r.class_session_id === classSessionId)
  if (matched) return matched
  return rooms[0]
}

function LoadingScreen() {
  return (
    <div className="relative isolate flex flex-col gap-8 py-16">
      <DecorativeBackground />
      <Skeleton className="h-4 w-40" />
      <Skeleton className="h-12 w-3/4" />
      <Skeleton className="h-72 w-full" />
    </div>
  )
}

function CenterMessage({
  title,
  description,
  backHref,
}: {
  title: string
  description: string
  backHref: string
}) {
  const { t } = useTranslation()
  return (
    <div className="relative isolate flex min-h-[60vh] flex-col items-start justify-center gap-4">
      <DecorativeBackground />
      <Eyebrow>{t("org.session.quizzes.take.eyebrow")}</Eyebrow>
      <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{title}</h1>
      <p className="text-muted-foreground max-w-md text-base leading-relaxed">{description}</p>
      <Button variant="outline" render={<Link to={backHref} />}>
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}

interface QuizRunnerProps {
  quiz: Quiz
  room: QuizRoom | undefined
  rules: QuizRule[]
  existingSubmission: QuizSubmission | undefined
  backHref: string
}

function QuizRunner({ quiz, room, rules, existingSubmission, backHref }: QuizRunnerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  // Resolve unique banks then filter by rule ids
  const uniqueBankIds = Array.from(
    new Set(rules.map((r) => r.bank_id).filter((id): id is string => !!id)),
  )

  const bankQueries = useQueries({
    queries: uniqueBankIds.map((bankId) =>
      getGetQuestionBanksIdQuestionsQueryOptions(bankId, undefined, {
        query: { enabled: !!bankId, staleTime: 60_000 },
      }),
    ),
  })

  const banksPending = bankQueries.some((q) => q.isPending)
  const banksError = bankQueries.some((q) => q.isError)

  const bankMap = new Map<string, Question[]>()
  bankQueries.forEach((q, idx) => {
    const id = uniqueBankIds[idx]
    if (q.data?.status === 200) {
      bankMap.set(id, q.data.data.data?.items ?? [])
    }
  })

  const allQuestions = buildQuestionList(rules, bankMap)

  // start submission mutation
  const startMutation = usePostQuizzesIdSubmissions({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201) {
          queryClient.invalidateQueries({ queryKey: getGetQuizzesIdSubmissionsQueryKey(quiz.id!) })
        }
      },
      onError: () => {
        toast.error(t("org.session.quizzes.take.startFailed"))
      },
    },
  })

  if (banksPending) {
    return <LoadingScreen />
  }

  if (banksError) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.bankError.title")}
        description={t("org.session.quizzes.take.bankError.description")}
        backHref={backHref}
      />
    )
  }

  if (allQuestions.length === 0) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.empty.title")}
        description={t("org.session.quizzes.take.empty.description")}
        backHref={backHref}
      />
    )
  }

  if (!room) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noRoom.title")}
        description={t("org.session.quizzes.take.noRoom.description")}
        backHref={backHref}
      />
    )
  }

  const now = Date.now()
  const roomStart = room.started_at ? new Date(room.started_at).getTime() : 0
  const roomEnd = room.ended_at ? new Date(room.ended_at).getTime() : 0
  const roomOpen = roomStart > 0 && now >= roomStart && (!roomEnd || now < roomEnd)

  if (!existingSubmission && !roomOpen) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.closed.title")}
        description={t("org.session.quizzes.take.closed.description")}
        backHref={backHref}
      />
    )
  }

  if (!existingSubmission) {
    return (
      <StartScreen
        quiz={quiz}
        room={room}
        totalQuestions={allQuestions.length}
        backHref={backHref}
        starting={startMutation.isPending}
        onBegin={() => {
          if (!room.id) return
          startMutation.mutate({ id: quiz.id!, data: { quiz_room_id: room.id } })
        }}
      />
    )
  }

  return (
    <PlayArea
      quiz={quiz}
      room={room}
      submission={existingSubmission}
      questions={allQuestions}
      backHref={backHref}
    />
  )
}

function buildQuestionList(rules: QuizRule[], bankMap: Map<string, Question[]>): Question[] {
  const seen = new Set<string>()
  const out: Question[] = []
  for (const rule of rules) {
    if (!rule.bank_id) continue
    const bankQs = bankMap.get(rule.bank_id) ?? []
    if (rule.type === "manual") {
      const ids = rule.question_ids ?? []
      for (const id of ids) {
        const q = bankQs.find((bq) => bq.id === id)
        if (q && q.id && !seen.has(q.id)) {
          seen.add(q.id)
          out.push(q)
        }
      }
    } else {
      // random rule: prefer materialized question_ids, else first N
      const ids = rule.question_ids ?? []
      const picked: Question[] = []
      if (ids.length > 0) {
        for (const id of ids) {
          const q = bankQs.find((bq) => bq.id === id)
          if (q) picked.push(q)
        }
      }
      if (picked.length === 0) {
        const count = rule.count ?? bankQs.length
        picked.push(...bankQs.slice(0, count))
      }
      for (const q of picked) {
        if (q.id && !seen.has(q.id)) {
          seen.add(q.id)
          out.push(q)
        }
      }
    }
  }
  return out
}

interface StartScreenProps {
  quiz: Quiz
  room: QuizRoom
  totalQuestions: number
  backHref: string
  starting: boolean
  onBegin: () => void
}

function StartScreen({ quiz, room, totalQuestions, backHref, starting, onBegin }: StartScreenProps) {
  const { t } = useTranslation()
  return (
    <div className="relative isolate flex flex-col gap-10 pb-24 pt-8">
      <DecorativeBackground />

      <div className="flex items-center justify-between">
        <Link
          to={backHref}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.session.quizzes.take.backToSession")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">
          № {(quiz.id ?? "").slice(0, 8).toUpperCase()}
        </span>
      </div>

      <header className="flex flex-col gap-5">
        <Eyebrow>{t("org.session.quizzes.take.eyebrow")}</Eyebrow>
        <h1 className="max-w-4xl text-4xl leading-tight font-semibold tracking-tight text-balance md:text-5xl lg:text-6xl">
          {quiz.title}
        </h1>
        {quiz.description ? (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed md:text-lg">
            {quiz.description}
          </p>
        ) : null}
      </header>

      <section className="bg-card ring-foreground/10 grid grid-cols-2 gap-0 overflow-hidden rounded-2xl px-4 py-6 ring-1 md:grid-cols-4 md:px-6">
        <MetaCell
          icon={<ClipboardListIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.questions")}
          value={totalQuestions.toString()}
          mono
        />
        <MetaCell
          icon={<ClockIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.duration")}
          value={`${quiz.duration_minutes ?? 0} ${t("org.session.quizzes.minutesShort")}`}
          mono
        />
        <MetaCell
          icon={<TrophyIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.totalScore")}
          value={(quiz.total_score ?? 0).toFixed(2)}
          mono
        />
        <MetaCell
          icon={<FlagIcon className="size-4" />}
          label={t("org.session.quizzes.take.meta.closesAt")}
          value={room.ended_at ? new Date(room.ended_at).toLocaleString() : "—"}
        />
      </section>

      <section className="flex flex-col gap-3">
        <Eyebrow>{t("org.session.quizzes.take.rules.eyebrow")}</Eyebrow>
        <ul className="text-foreground/80 max-w-2xl space-y-2 text-sm leading-relaxed">
          <li className="flex items-start gap-2">
            <ClockIcon className="text-muted-foreground mt-0.5 size-4" />
            {t("org.session.quizzes.take.rules.timer", { minutes: quiz.duration_minutes ?? 0 })}
          </li>
          {quiz.no_back_navigation ? (
            <li className="flex items-start gap-2">
              <LockKeyholeIcon className="text-muted-foreground mt-0.5 size-4" />
              {t("org.session.quizzes.take.rules.noBack")}
            </li>
          ) : (
            <li className="flex items-start gap-2">
              <ArrowLeftIcon className="text-muted-foreground mt-0.5 size-4" />
              {t("org.session.quizzes.take.rules.backAllowed")}
            </li>
          )}
          {quiz.shuffle_questions ? (
            <li className="flex items-start gap-2">
              <ShuffleIcon className="text-muted-foreground mt-0.5 size-4" />
              {t("org.session.quizzes.take.rules.shuffle")}
            </li>
          ) : null}
          <li className="flex items-start gap-2">
            <AlertTriangleIcon className="text-muted-foreground mt-0.5 size-4" />
            {t("org.session.quizzes.take.rules.autoSubmit")}
          </li>
        </ul>
      </section>

      <div className="flex items-center gap-3">
        <Button size="lg" onClick={onBegin} disabled={starting}>
          {starting ? <Spinner className="size-4" /> : <CheckCircle2Icon className="size-4" />}
          {t("org.session.quizzes.take.begin")}
        </Button>
        <Button variant="outline" render={<Link to={backHref} />}>
          {t("common.cancel")}
        </Button>
      </div>
    </div>
  )
}

function MetaCell({
  icon,
  label,
  value,
  mono = false,
}: {
  icon: React.ReactNode
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className="flex flex-col gap-2 border-b border-dashed py-5 pe-4 ps-4 md:border-b-0 md:border-s md:py-0 md:first:border-s-0 md:first:ps-0">
      <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase">
        {icon}
        {label}
      </span>
      <span
        className={cn(
          "text-foreground text-base leading-tight font-medium md:text-lg",
          mono && "font-mono tabular-nums",
        )}
      >
        {value}
      </span>
    </div>
  )
}

interface PlayAreaProps {
  quiz: Quiz
  room: QuizRoom
  submission: QuizSubmission
  questions: Question[]
  backHref: string
}

function PlayArea({ quiz, room, submission, questions, backHref }: PlayAreaProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const now = useNow(1000)

  const submissionId = submission.id!
  const startedAtMs = new Date(submission.started_at ?? Date.now()).getTime()
  const durationMs = (quiz.duration_minutes ?? 0) * 60_000
  const roomEndMs = room.ended_at ? new Date(room.ended_at).getTime() : Infinity
  const deadline = Math.min(startedAtMs + durationMs, roomEndMs)
  const remainingSeconds = Math.max(0, Math.floor((deadline - now) / 1000))
  const isExpired = remainingSeconds <= 0

  // Order: optionally shuffle per-user deterministically using submissionId
  const orderedQuestionIds = quiz.shuffle_questions
    ? shuffleSeeded(
        questions.map((q) => q.id!),
        submissionId,
      )
    : questions.map((q) => q.id!)

  // Local state hydrated from storage
  const [answers, setAnswers] = useState<Record<string, AnswerState>>(() => {
    const saved = loadLocal(submissionId)
    return saved?.answers ?? {}
  })
  const [index, setIndex] = useState<number>(() => {
    const saved = loadLocal(submissionId)
    if (saved && saved.index >= 0 && saved.index < orderedQuestionIds.length) {
      return saved.index
    }
    return 0
  })
  const [confirmOpen, setConfirmOpen] = useState(false)

  const lastTickRef = useRef<number>(Date.now())
  const submittedOnceRef = useRef(false)

  // persist local state on changes
  useEffect(() => {
    saveLocal(submissionId, { answers, order: orderedQuestionIds, index })
  }, [answers, index, orderedQuestionIds, submissionId])

  // tick per-question timer every second on the active question
  useEffect(() => {
    const currentId = orderedQuestionIds[index]
    if (!currentId) return
    const interval = setInterval(() => {
      const now = Date.now()
      const delta = Math.max(0, Math.floor((now - lastTickRef.current) / 1000))
      if (delta > 0) {
        lastTickRef.current = now
        setAnswers((prev) => {
          const existing = prev[currentId] ?? emptyAnswer()
          return {
            ...prev,
            [currentId]: { ...existing, spent_seconds: existing.spent_seconds + delta },
          }
        })
      }
    }, 1000)
    return () => {
      const now = Date.now()
      const delta = Math.max(0, Math.floor((now - lastTickRef.current) / 1000))
      lastTickRef.current = now
      if (delta > 0) {
        setAnswers((prev) => {
          const existing = prev[currentId] ?? emptyAnswer()
          return {
            ...prev,
            [currentId]: { ...existing, spent_seconds: existing.spent_seconds + delta },
          }
        })
      }
      clearInterval(interval)
    }
  }, [index, orderedQuestionIds])

  const submitMutation = usePostQuizzesSubmissionsSubmissionIdSubmit({
    mutation: {
      onSuccess: () => {
        clearLocal(submissionId)
        queryClient.invalidateQueries({ queryKey: getGetQuizzesIdSubmissionsQueryKey(quiz.id!) })
        toast.success(t("org.session.quizzes.take.submitted"))
      },
      onError: () => {
        toast.error(t("org.session.quizzes.take.submitFailed"))
      },
    },
  })

  function finalize(reason: "manual" | "auto") {
    if (submittedOnceRef.current) return
    submittedOnceRef.current = true
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
    if (reason === "auto") {
      toast.message(t("org.session.quizzes.take.autoSubmitNote"))
    }
  }

  // Auto-submit when time is up
  useEffect(() => {
    if (isExpired && !submittedOnceRef.current && submission.status === "in_progress") {
      finalize("auto")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isExpired])

  // Best-effort save on unload
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (!submittedOnceRef.current && submission.status === "in_progress") {
        e.preventDefault()
        e.returnValue = ""
      }
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [submission.status])

  const currentQId = orderedQuestionIds[index]
  const currentQuestion = questions.find((q) => q.id === currentQId)
  const currentAnswer = answers[currentQId ?? ""] ?? emptyAnswer()

  const total = orderedQuestionIds.length
  const isLast = index === total - 1
  const isFirst = index === 0

  function setAnswer(updater: (prev: AnswerState) => AnswerState) {
    setAnswers((prev) => ({ ...prev, [currentQId!]: updater(prev[currentQId!] ?? emptyAnswer()) }))
  }

  if (submitMutation.isSuccess || submission.status !== "in_progress") {
    // show interstitial; results page will be rendered after refetch
    return (
      <div className="relative isolate flex min-h-[60vh] flex-col items-center justify-center gap-4 py-16 text-center">
        <DecorativeBackground />
        <Spinner className="size-6" />
        <Eyebrow>{t("org.session.quizzes.take.submitting")}</Eyebrow>
      </div>
    )
  }

  if (!currentQuestion) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.empty.title")}
        description={t("org.session.quizzes.take.empty.description")}
        backHref={backHref}
      />
    )
  }

  const danger = remainingSeconds <= 30
  const warn = remainingSeconds <= 60 && !danger

  return (
    <div className="relative isolate flex flex-col gap-8 pb-32 pt-6">
      <DecorativeBackground />

      <div className="sticky top-0 z-10 -mx-4 backdrop-blur md:-mx-6">
        <div className="bg-background/70 border-foreground/10 mx-4 flex items-center justify-between gap-4 rounded-2xl border px-4 py-3 md:mx-6 md:px-6">
          <div className="flex items-center gap-3">
            <span className="text-muted-foreground font-mono text-xs tracking-[0.25em] uppercase">
              {t("org.session.quizzes.take.progress", { current: index + 1, total })}
            </span>
            <ProgressDots total={total} index={index} answered={answers} order={orderedQuestionIds} />
          </div>
          <div
            className={cn(
              "ring-foreground/15 inline-flex items-center gap-2 rounded-xl px-3 py-1.5 font-mono text-base tabular-nums ring-1 transition-colors md:text-lg",
              warn && "ring-amber-500/50 text-amber-600 dark:text-amber-400",
              danger && "ring-destructive/60 text-destructive animate-pulse",
            )}
          >
            <ClockIcon className="size-4" />
            {formatClock(remainingSeconds)}
          </div>
        </div>
      </div>

      <article className="bg-card ring-foreground/10 relative flex flex-col gap-6 overflow-hidden rounded-3xl p-6 ring-1 md:p-10">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/6%,transparent_60%)]"
        />
        <div className="flex items-center justify-between gap-3">
          <Eyebrow>
            {t("org.session.quizzes.take.questionN", { n: index + 1 })} · {questionTypeLabel(currentQuestion.type, t)}
          </Eyebrow>
          <span className="text-muted-foreground font-mono text-[10px] tracking-[0.25em]">
            /{String(index + 1).padStart(2, "0")}
          </span>
        </div>

        <h2 className="max-w-3xl text-2xl leading-snug font-semibold tracking-tight text-balance md:text-3xl">
          {currentQuestion.text}
        </h2>

        <QuestionInput
          question={currentQuestion}
          answer={currentAnswer}
          onChange={setAnswer}
        />
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

function ProgressDots({
  total,
  index,
  answered,
  order,
}: {
  total: number
  index: number
  answered: Record<string, AnswerState>
  order: string[]
}) {
  return (
    <div className="flex items-center gap-1">
      {order.map((qid, i) => {
        const a = answered[qid]
        const hasAnswer = !!a && (a.selected_option_ids.length > 0 || a.value.trim().length > 0)
        return (
          <span
            key={qid}
            aria-hidden
            className={cn(
              "size-1.5 rounded-full transition-all",
              i === index
                ? "bg-foreground w-4"
                : hasAnswer
                ? "bg-foreground/70"
                : "bg-foreground/20",
            )}
          />
        )
      })}
    </div>
  )
}

function emptyAnswer(): AnswerState {
  return { selected_option_ids: [], value: "", spent_seconds: 0 }
}

function countAnswered(
  answers: Record<string, AnswerState>,
  order: string[],
  questions: Question[],
): number {
  let n = 0
  for (const qid of order) {
    const q = questions.find((qq) => qq.id === qid)
    const a = answers[qid]
    if (!a || !q) continue
    if (q.type === "choice" && a.selected_option_ids.length > 0) n++
    else if ((q.type === "short_answer" || q.type === "descriptive") && a.value.trim().length > 0) n++
  }
  return n
}

function questionTypeLabel(type: Question["type"], t: (k: string) => string) {
  if (type === "choice") return t("org.session.quizzes.take.types.choice")
  if (type === "short_answer") return t("org.session.quizzes.take.types.short")
  return t("org.session.quizzes.take.types.descriptive")
}

interface QuestionInputProps {
  question: Question
  answer: AnswerState
  onChange: (updater: (prev: AnswerState) => AnswerState) => void
}

function QuestionInput({ question, answer, onChange }: QuestionInputProps) {
  const { t } = useTranslation()
  const opts = question.options ?? []

  if (question.type === "choice") {
    const isMulti = countPositive(opts) > 1
    return (
      <div className="flex flex-col gap-2">
        {opts.map((opt, i) => {
          const id = opt.id ?? String(i)
          const checked = answer.selected_option_ids.includes(id)
          return (
            <OptionTile
              key={id}
              index={i}
              label={opt.value ?? ""}
              checked={checked}
              onClick={() => {
                onChange((prev) => {
                  if (isMulti) {
                    return {
                      ...prev,
                      selected_option_ids: checked
                        ? prev.selected_option_ids.filter((x) => x !== id)
                        : [...prev.selected_option_ids, id],
                    }
                  }
                  return { ...prev, selected_option_ids: checked ? [] : [id] }
                })
              }}
            />
          )
        })}
        {isMulti ? (
          <span className="text-muted-foreground mt-1 font-mono text-[10px] tracking-[0.25em] uppercase">
            {t("org.session.quizzes.take.multiSelectHint")}
          </span>
        ) : null}
      </div>
    )
  }

  if (question.type === "short_answer") {
    return (
      <Input
        value={answer.value}
        onChange={(e) => onChange((prev) => ({ ...prev, value: e.target.value }))}
        placeholder={t("org.session.quizzes.take.shortPlaceholder")}
        className="max-w-2xl text-base"
      />
    )
  }

  return (
    <Textarea
      value={answer.value}
      onChange={(e) => onChange((prev) => ({ ...prev, value: e.target.value }))}
      placeholder={t("org.session.quizzes.take.descriptivePlaceholder")}
      rows={8}
      className="max-w-3xl text-base"
    />
  )
}

function countPositive(opts: QOption[]): number {
  let n = 0
  for (const o of opts) if ((o.score ?? 0) > 0) n++
  return n
}

function OptionTile({
  index,
  label,
  checked,
  onClick,
}: {
  index: number
  label: string
  checked: boolean
  onClick: () => void
}) {
  const letter = String.fromCharCode(65 + index)
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "group/option ring-foreground/15 hover:ring-foreground/40 relative flex items-center gap-4 rounded-xl bg-card px-4 py-3 text-start ring-1 transition-all hover:-translate-y-0.5",
        checked && "ring-foreground bg-foreground/[0.04] shadow-sm",
      )}
    >
      <span
        className={cn(
          "ring-foreground/20 flex size-8 shrink-0 items-center justify-center rounded-lg font-mono text-sm font-semibold ring-1 transition-colors",
          checked && "bg-foreground text-background ring-foreground",
        )}
      >
        {letter}
      </span>
      <span className="text-foreground text-base leading-snug">{label}</span>
      {checked ? <CheckCircle2Icon className="text-foreground ms-auto size-5" /> : null}
    </button>
  )
}

interface ResultScreenProps {
  quiz: Quiz
  submission: QuizSubmission
  backHref: string
}

function ResultScreen({ quiz, submission, backHref }: ResultScreenProps) {
  const { t } = useTranslation()
  const total = quiz.total_score ?? 0
  const earned = submission.total_score ?? 0
  const pct = total > 0 ? Math.round((earned / total) * 100) : 0
  const answers = submission.answers ?? []
  const totalSpent = answers.reduce((acc, a) => acc + (a.spent_seconds ?? 0), 0)

  return (
    <div className="relative isolate flex flex-col gap-10 pb-24 pt-8">
      <DecorativeBackground />

      <div className="flex items-center justify-between">
        <Link
          to={backHref}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.session.quizzes.take.backToSession")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">
          {submission.status === "graded"
            ? t("org.session.quizzes.take.result.graded")
            : t("org.session.quizzes.take.result.submitted")}
        </span>
      </div>

      <header className="flex flex-col gap-4">
        <Eyebrow>{t("org.session.quizzes.take.result.eyebrow")}</Eyebrow>
        <h1 className="text-4xl leading-tight font-semibold tracking-tight md:text-5xl">
          {quiz.title}
        </h1>
      </header>

      <section className="bg-card ring-foreground/10 grid grid-cols-2 gap-0 overflow-hidden rounded-3xl px-4 py-6 ring-1 md:grid-cols-4 md:px-6">
        <MetaCell
          icon={<TrophyIcon className="size-4" />}
          label={t("org.session.quizzes.take.result.earned")}
          value={earned.toFixed(2)}
          mono
        />
        <MetaCell
          icon={<FlagIcon className="size-4" />}
          label={t("org.session.quizzes.take.result.outOf")}
          value={total.toFixed(2)}
          mono
        />
        <MetaCell
          icon={<CheckCircle2Icon className="size-4" />}
          label={t("org.session.quizzes.take.result.percent")}
          value={`${pct}%`}
          mono
        />
        <MetaCell
          icon={<ClockIcon className="size-4" />}
          label={t("org.session.quizzes.take.result.totalTime")}
          value={formatClock(totalSpent)}
          mono
        />
      </section>

      <section className="flex flex-col gap-3">
        <Eyebrow>{t("org.session.quizzes.take.result.perQuestionEyebrow")}</Eyebrow>
        <div className="flex flex-col divide-y divide-dashed">
          {answers.map((a, i) => (
            <div key={a.question_id} className="flex items-center justify-between gap-4 py-3">
              <span className="text-muted-foreground font-mono text-xs tracking-[0.2em]">
                Q{String(i + 1).padStart(2, "0")}
              </span>
              <span className="text-foreground/80 ms-2 grow truncate font-mono text-xs">
                {a.question_id?.slice(0, 8)}
              </span>
              <span className="text-muted-foreground font-mono text-xs tabular-nums">
                {formatClock(a.spent_seconds ?? 0)}
              </span>
              <span className="text-foreground font-mono text-sm font-semibold tabular-nums">
                {(a.earned_score ?? 0).toFixed(2)}
              </span>
            </div>
          ))}
        </div>
      </section>

      <Button variant="outline" render={<Link to={backHref} />}>
        <ArrowLeftIcon className="size-4" />
        {t("org.session.quizzes.take.backToSession")}
      </Button>
    </div>
  )
}

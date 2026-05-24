import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizRule as QuizRule,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { useQueries, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetQuestionBanksIdQuestionsQueryOptions } from "@/api/question-banks/question-banks"
import {
  getGetQuizzesIdSubmissionsQueryKey,
  usePostQuizzesIdSubmissions,
} from "@/api/quizzes/quizzes"

import { CenterMessage, LoadingScreen } from "./messages"
import { PlayArea } from "./play-area"
import { StartScreen } from "./start-screen"
import { buildQuestionList, isRoomOpen } from "./utils"

interface QuizRunnerProps {
  quiz: Quiz
  room: QuizRoom | undefined
  rules: QuizRule[]
  existingSubmission: QuizSubmission | undefined
  backHref: string
}

export function QuizRunner({ quiz, room, rules, existingSubmission, backHref }: QuizRunnerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

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

  const startMutation = usePostQuizzesIdSubmissions({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201 && quiz.id) {
          queryClient.invalidateQueries({ queryKey: getGetQuizzesIdSubmissionsQueryKey(quiz.id) })
        }
      },
      onError: () => toast.error(t("org.session.quizzes.take.startFailed")),
    },
  })

  if (bankQueries.some((q) => q.isPending)) return <LoadingScreen />
  if (bankQueries.some((q) => q.isError)) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.bankError.title")}
        description={t("org.session.quizzes.take.bankError.description")}
        backHref={backHref}
      />
    )
  }

  const bankMap = new Map<string, Question[]>()
  bankQueries.forEach((q, i) => {
    const id = uniqueBankIds[i]
    if (q.data?.status === 200) bankMap.set(id, q.data.data.data?.items ?? [])
  })

  const questions = buildQuestionList(rules, bankMap)

  if (questions.length === 0) {
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

  if (!existingSubmission && !isRoomOpen(room, Date.now())) {
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
        totalQuestions={questions.length}
        backHref={backHref}
        starting={startMutation.isPending}
        onBegin={() => {
          if (!room.id || !quiz.id) return
          startMutation.mutate({ id: quiz.id, data: { quiz_room_id: room.id } })
        }}
      />
    )
  }

  return (
    <PlayArea
      quiz={quiz}
      room={room}
      submission={existingSubmission}
      questions={questions}
      backHref={backHref}
    />
  )
}

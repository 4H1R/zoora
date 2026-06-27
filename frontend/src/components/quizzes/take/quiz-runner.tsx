import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuizzesIdSubmissionsQueryKey,
  usePostQuizzesIdSubmissions,
} from "@/api/quizzes/quizzes"

import { CenterMessage } from "./messages"
import { PlayArea } from "./play-area"
import { StartScreen } from "./start-screen"
import { isRoomOpen } from "./utils"

interface QuizRunnerProps {
  quiz: Quiz
  room: QuizRoom | undefined
  questions: Question[]
  existingSubmission: QuizSubmission | undefined
  backHref: string
}

export function QuizRunner({
  quiz,
  room,
  questions,
  existingSubmission,
  backHref,
}: QuizRunnerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

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
        questions={questions}
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

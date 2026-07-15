import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
  GithubCom4H1RZooraInternalDomainQuizRoom as QuizRoom,
  GithubCom4H1RZooraInternalDomainQuizSubmission as QuizSubmission,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { DoorClosedIcon, LockKeyholeIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetQuizzesIdSubmissionsQueryKey, usePostQuizzesIdSubmissions } from "@/api/quizzes/quizzes"

import { requestGeolocation } from "./geolocation"
import { CenterMessage } from "./messages"
import { PlayArea } from "./play-area"
import { StartScreen } from "./start-screen"
import { isRoomOpen } from "./utils"

interface QuizRunnerProps {
  quiz: Quiz
  room: QuizRoom | undefined
  questions: Question[]
  // Pre-start metadata from /preview: the real questions are only fetched once a
  // submission exists (the backend freezes them on start), so the StartScreen is
  // driven by these instead of the (empty, pre-start) questions array.
  previewCount: number
  previewHasNegative: boolean
  existingSubmission: QuizSubmission | undefined
  backHref: string
}

export function QuizRunner({
  quiz,
  room,
  questions,
  previewCount,
  previewHasNegative,
  existingSubmission,
  backHref,
}: QuizRunnerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [locating, setLocating] = useState(false)

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

  // Pre-start the questions array is empty (fetched only after starting), so the
  // "no questions" guard uses the preview count instead; once playing, it uses
  // the frozen set.
  const questionCount = existingSubmission ? questions.length : previewCount
  if (questionCount === 0) {
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
        icon={<DoorClosedIcon />}
      />
    )
  }

  if (!existingSubmission && !isRoomOpen(room, Date.now())) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.closed.title")}
        description={t("org.session.quizzes.take.closed.description")}
        backHref={backHref}
        icon={<LockKeyholeIcon />}
      />
    )
  }

  if (!existingSubmission) {
    async function handleBegin() {
      if (!room?.id || !quiz.id) return
      // require_gps: must send coords OR gps_denied, or the backend rejects the
      // start with a validation error. Denial is non-blocking — proceed anyway.
      let geo = {}
      if (quiz.require_gps) {
        setLocating(true)
        const result = await requestGeolocation()
        setLocating(false)
        if (result.gps_denied) toast.warning(t("org.session.quizzes.take.antiCheat.gpsDenied"))
        geo = result
      }
      startMutation.mutate({ id: quiz.id, data: { quiz_room_id: room.id, ...geo } })
    }

    return (
      <StartScreen
        quiz={quiz}
        room={room}
        totalQuestions={previewCount}
        hasNegativeMarking={previewHasNegative}
        backHref={backHref}
        starting={startMutation.isPending}
        locating={locating}
        onBegin={() => void handleBegin()}
      />
    )
  }

  return <PlayArea quiz={quiz} room={room} submission={existingSubmission} questions={questions} backHref={backHref} />
}

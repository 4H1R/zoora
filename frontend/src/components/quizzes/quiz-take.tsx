import { useTranslation } from "react-i18next"

import {
  useGetQuizzesId,
  useGetQuizzesIdQuestions,
  useGetQuizzesIdRooms,
  useGetQuizzesIdSubmissions,
} from "@/api/quizzes/quizzes"
import { useQuizPermissions } from "@/components/org/quizzes/use-quiz-permissions"
import { CenterMessage, LoadingScreen } from "@/components/quizzes/take/messages"
import { QuizRunner } from "@/components/quizzes/take/quiz-runner"
import { ResultScreen } from "@/components/quizzes/take/result-screen"
import { pickRoomForSession } from "@/components/quizzes/take/utils"

// QuizTake renders the full exam-taking flow for a quiz. The logic only needs
// quizId; classSessionId is optional and only affects room selection + the
// back link. Both the class-scoped take route and the standalone exam take
// route render this component.
export function QuizTake({
  quizId,
  classSessionId,
}: {
  quizId: string
  classSessionId?: string
}) {
  const { t } = useTranslation()
  const { canView } = useQuizPermissions()

  const backHref = classSessionId
    ? `/org/classes/classsessions/${classSessionId}`
    : `/org/exams`

  const quizQ = useGetQuizzesId(quizId, { query: { enabled: canView } })
  const quiz = (quizQ.data?.status === 200 && quizQ.data.data.data) || undefined

  const roomsQ = useGetQuizzesIdRooms(quizId, undefined, {
    query: { enabled: canView && !!quiz },
  })
  const rooms = (roomsQ.data?.status === 200 && roomsQ.data.data.data?.items) || []

  const questionsQ = useGetQuizzesIdQuestions(quizId, {
    query: { enabled: canView && !!quiz },
  })
  const questions = (questionsQ.data?.status === 200 && questionsQ.data.data.data?.items) || []

  const submissionsQ = useGetQuizzesIdSubmissions(quizId, undefined, {
    query: { enabled: canView && !!quiz },
  })
  const submissions = (submissionsQ.data?.status === 200 && submissionsQ.data.data.data?.items) || []
  const inProgress = submissions.find((s) => s.status === "in_progress")
  const finalSubmission = submissions.find((s) => s.status !== "in_progress")

  if (!canView) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={backHref}
      />
    )
  }

  if (quizQ.isPending || roomsQ.isPending || questionsQ.isPending || submissionsQ.isPending) {
    return <LoadingScreen />
  }

  if (quizQ.data?.status === 403 || questionsQ.data?.status === 403) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={backHref}
      />
    )
  }

  if (quizQ.isError || !quiz) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.notFound.title")}
        description={t("org.session.quizzes.take.notFound.description")}
        backHref={backHref}
      />
    )
  }

  if (questionsQ.isError) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.bankError.title")}
        description={t("org.session.quizzes.take.bankError.description")}
        backHref={backHref}
      />
    )
  }

  if (finalSubmission) {
    return <ResultScreen quiz={quiz} submission={finalSubmission} backHref={backHref} />
  }

  return (
    <QuizRunner
      quiz={quiz}
      room={pickRoomForSession(rooms, classSessionId ?? "")}
      questions={questions}
      existingSubmission={inProgress}
      backHref={backHref}
    />
  )
}

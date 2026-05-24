import { createFileRoute } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import {
  useGetQuizzesId,
  useGetQuizzesIdRooms,
  useGetQuizzesIdRules,
  useGetQuizzesIdSubmissions,
} from "@/api/quizzes/quizzes"
import { useQuizPermissions } from "@/components/org/quizzes/use-quiz-permissions"
import { LoadingScreen, CenterMessage } from "@/components/quizzes/take/messages"
import { QuizRunner } from "@/components/quizzes/take/quiz-runner"
import { ResultScreen } from "@/components/quizzes/take/result-screen"
import { pickRoomForSession } from "@/components/quizzes/take/utils"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute(
  "/_auth/org/$orgId/classes/classsessions/$classSessionId/quizzes/$quizId/take",
)({
  head: () => orgHead("org.session.quizzes.take.headTitle"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { orgId, classSessionId, quizId } = Route.useParams()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { canView } = useQuizPermissions()

  const backHref = `/org/${orgId}/classes/classsessions/${classSessionId}`

  const quizQ = useGetQuizzesId(quizId)
  const quiz = (quizQ.data?.status === 200 && quizQ.data.data.data) || undefined

  const roomsQ = useGetQuizzesIdRooms(quizId, undefined, { query: { enabled: !!quiz } })
  const rooms = (roomsQ.data?.status === 200 && roomsQ.data.data.data?.items) || []

  const rulesQ = useGetQuizzesIdRules(quizId, undefined, { query: { enabled: !!quiz } })
  const rules = (rulesQ.data?.status === 200 && rulesQ.data.data.data?.items) || []

  const submissionsQ = useGetQuizzesIdSubmissions(quizId, undefined, {
    query: { enabled: !!quiz },
  })
  const submissions =
    (submissionsQ.data?.status === 200 && submissionsQ.data.data.data?.items) || []
  const inProgress = submissions.find((s) => s.status === "in_progress")
  const finalSubmission = submissions.find((s) => s.status !== "in_progress")

  if (!allowed) return null

  if (!canView) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={backHref}
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
      room={pickRoomForSession(rooms, classSessionId)}
      rules={rules}
      existingSubmission={inProgress}
      backHref={backHref}
    />
  )
}

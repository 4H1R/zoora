import { SearchXIcon, ShieldXIcon, TriangleAlertIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  useGetQuizzesId,
  useGetQuizzesIdPreview,
  useGetQuizzesIdQuestions,
  useGetQuizzesIdRooms,
  useGetQuizzesIdSubmissions,
} from "@/api/quizzes/quizzes"
import { useGetUsersMe } from "@/api/users/users"
import { userHasAny } from "@/lib/access"
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
  // This component renders on the standalone /quiz route, which is OUTSIDE the
  // org <AccessProvider>, so the useAccess* hooks would throw. Read perms from
  // the fetched /users/me object with the hook-free helper instead.
  const meQ = useGetUsersMe()
  const me = (meQ.data?.status === 200 && meQ.data.data.data) || undefined
  const canView = userHasAny(me, ["quizzes:view", "quizzes:view_any"])
  // Taking a quiz requires enrollment in its class; the backend enforces this
  // and only the Student preset carries quizzes:take. A viewer (e.g. Manager)
  // can open the quiz but must not see the start flow — starting would 403.
  const canTake = userHasAny(me, ["quizzes:take"])

  const backHref = classSessionId
    ? `/org/classes/class-sessions/${classSessionId}`
    : `/org/exams`

  const quizQ = useGetQuizzesId(quizId, { query: { enabled: canView } })
  const quiz = (quizQ.data?.status === 200 && quizQ.data.data.data) || undefined

  const roomsQ = useGetQuizzesIdRooms(quizId, undefined, {
    query: { enabled: canView && !!quiz },
  })
  const rooms = (roomsQ.data?.status === 200 && roomsQ.data.data.data?.items) || []

  const submissionsQ = useGetQuizzesIdSubmissions(quizId, undefined, {
    query: { enabled: canView && !!quiz },
  })
  const submissions = (submissionsQ.data?.status === 200 && submissionsQ.data.data.data?.items) || []
  const inProgress = submissions.find((s) => s.status === "in_progress")
  const finalSubmission = submissions.find((s) => s.status !== "in_progress")
  const hasSubmission = Boolean(inProgress || finalSubmission)

  // The backend freezes the question set on submission start and 404s
  // /questions until then, so only fetch the real questions once a submission
  // exists (in-progress → PlayArea, final → ResultScreen). Before starting, the
  // StartScreen is driven by /preview (count + negative-marking flag) instead —
  // no question bodies leak and there's no pre-start 404.
  const questionsEnabled = canView && !!quiz && submissionsQ.data?.status === 200 && hasSubmission
  const questionsQ = useGetQuizzesIdQuestions(quizId, {
    query: { enabled: questionsEnabled },
  })
  const questions = (questionsQ.data?.status === 200 && questionsQ.data.data.data?.items) || []

  const previewEnabled = canView && !!quiz && submissionsQ.data?.status === 200 && !hasSubmission
  const previewQ = useGetQuizzesIdPreview(quizId, {
    query: { enabled: previewEnabled },
  })
  const preview = (previewQ.data?.status === 200 && previewQ.data.data.data) || undefined

  if (meQ.isPending) {
    return <LoadingScreen />
  }

  if (!canView) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={backHref}
        icon={<ShieldXIcon />}
        tone="destructive"
      />
    )
  }

  if (
    quizQ.isPending ||
    roomsQ.isPending ||
    submissionsQ.isPending ||
    (questionsEnabled && questionsQ.isLoading) ||
    (previewEnabled && previewQ.isLoading)
  ) {
    return <LoadingScreen />
  }

  if (
    quizQ.data?.status === 403 ||
    questionsQ.data?.status === 403 ||
    previewQ.data?.status === 403
  ) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.noAccess.title")}
        description={t("org.session.quizzes.take.noAccess.description")}
        backHref={backHref}
        icon={<ShieldXIcon />}
        tone="destructive"
      />
    )
  }

  if (quizQ.isError || !quiz) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.notFound.title")}
        description={t("org.session.quizzes.take.notFound.description")}
        backHref={backHref}
        icon={<SearchXIcon />}
      />
    )
  }

  if (questionsQ.isError || previewQ.isError) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.bankError.title")}
        description={t("org.session.quizzes.take.bankError.description")}
        backHref={backHref}
        icon={<TriangleAlertIcon />}
        tone="destructive"
      />
    )
  }

  if (finalSubmission) {
    return (
      <ResultScreen
        quiz={quiz}
        submission={finalSubmission}
        questions={questions}
        backHref={backHref}
      />
    )
  }

  // Viewers without take permission (e.g. Manager) can reach this screen but
  // cannot start the quiz — show a view-only notice instead of the start flow.
  if (!canTake) {
    return (
      <CenterMessage
        title={t("org.session.quizzes.take.viewOnly.title")}
        description={t("org.session.quizzes.take.viewOnly.description")}
        backHref={backHref}
        icon={<ShieldXIcon />}
      />
    )
  }

  return (
    <QuizRunner
      quiz={quiz}
      room={pickRoomForSession(rooms, classSessionId ?? "")}
      questions={questions}
      previewCount={preview?.question_count ?? 0}
      previewHasNegative={preview?.has_negative_marking ?? false}
      existingSubmission={inProgress}
      backHref={backHref}
    />
  )
}

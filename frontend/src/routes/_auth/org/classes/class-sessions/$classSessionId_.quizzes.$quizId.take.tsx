import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute(
  "/_auth/org/classes/class-sessions/$classSessionId_/quizzes/$quizId/take",
)({
  head: () => orgHead("org.session.quizzes.take.headTitle"),
  component: RouteComponent,
})

function RouteComponent() {
  const { classSessionId, quizId } = Route.useParams()
  return <QuizTake quizId={quizId} classSessionId={classSessionId} />
}

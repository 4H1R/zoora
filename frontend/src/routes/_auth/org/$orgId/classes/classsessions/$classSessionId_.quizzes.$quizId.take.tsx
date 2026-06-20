import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute(
  "/_auth/org/$orgId/classes/classsessions/$classSessionId_/quizzes/$quizId/take",
)({
  head: () => orgHead("org.session.quizzes.take.headTitle"),
  component: RouteComponent,
})

function RouteComponent() {
  const { orgId, classSessionId, quizId } = Route.useParams()
  return <QuizTake orgId={orgId} quizId={quizId} classSessionId={classSessionId} />
}

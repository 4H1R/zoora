import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/exams/$quizId/take")({
  head: () => orgHead("org.nav.exams"),
  component: RouteComponent,
})

function RouteComponent() {
  const { orgId, quizId } = Route.useParams()
  return <QuizTake orgId={orgId} quizId={quizId} />
}

import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/exams/$quizId/take")({
  head: () => orgHead("org.nav.exams"),
  component: RouteComponent,
})

function RouteComponent() {
  const { quizId } = Route.useParams()
  return <QuizTake quizId={quizId} />
}

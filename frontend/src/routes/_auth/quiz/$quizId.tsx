import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/quiz/$quizId")({
  head: () => orgHead("org.nav.exams"),
  // classSessionId is optional: present when the quiz is taken inside a live
  // class session, absent for a standalone exam. Same page either way.
  validateSearch: (search: Record<string, unknown>): { classSessionId?: string } => ({
    classSessionId:
      typeof search.classSessionId === "string" ? search.classSessionId : undefined,
  }),
  component: RouteComponent,
})

function RouteComponent() {
  const { quizId } = Route.useParams()
  const { classSessionId } = Route.useSearch()
  return <QuizTake quizId={quizId} classSessionId={classSessionId} />
}

import { createFileRoute } from "@tanstack/react-router"

import { QuizTake } from "@/components/quizzes/quiz-take"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/quiz/$quizId")({
  head: () => orgHead("org.nav.exams"),
  // classSessionId is optional: present when the quiz is taken inside a live
  // class session, absent for a standalone exam. Same page either way.
  validateSearch: (search: Record<string, unknown>): { classSessionId?: string } => ({
    classSessionId: typeof search.classSessionId === "string" ? search.classSessionId : undefined,
  }),
  component: RouteComponent,
})

function RouteComponent() {
  const { quizId } = Route.useParams()
  const { classSessionId } = Route.useSearch()
  // The standalone /quiz route renders outside the org layout (bare <Outlet />),
  // so it has no container. Wrap here to center + constrain like the org content
  // area — otherwise the take screens bleed edge-to-edge (off-screen in RTL).
  return (
    <div className="container py-4">
      <QuizTake quizId={quizId} classSessionId={classSessionId} />
    </div>
  )
}

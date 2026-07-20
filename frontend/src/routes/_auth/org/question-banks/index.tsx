import { createFileRoute } from "@tanstack/react-router"

import { QuestionBanksSection } from "@/components/org/question-banks/QuestionBanksSection"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/question-banks/")({
  head: () => orgHead("org.nav.questionBanks"),
  component: RouteComponent,
})

function RouteComponent() {
  const allowed = useOrgGuard(["question_banks:view", "question_banks:view_any"])
  if (!allowed) return null
  return <QuestionBanksSection />
}

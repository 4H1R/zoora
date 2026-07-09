import { createFileRoute } from "@tanstack/react-router"
import { MessagesSquareIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { EmptyState } from "@/components/ui/empty-state"

export const Route = createFileRoute("/_auth/org/conversations/")({
  component: ConversationsIndex,
})

// Empty detail pane, shown when no conversation is selected. Perm + plan gating
// live in the parent layout route (`route.tsx`).
function ConversationsIndex() {
  const { t } = useTranslation()
  return (
    <EmptyState
      icon={MessagesSquareIcon}
      title={t("conversations.detail.empty.title")}
      description={t("conversations.detail.empty.description")}
      className="flex-1 justify-center border-0 bg-transparent ring-0"
    />
  )
}

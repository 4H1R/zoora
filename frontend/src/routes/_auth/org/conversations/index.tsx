import { createFileRoute } from "@tanstack/react-router"
import { MessagesSquareIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ChatBackground } from "@/components/org/conversations/chat-background"
import { EmptyState } from "@/components/ui/empty-state"

export const Route = createFileRoute("/_auth/org/conversations/")({
  component: ConversationsIndex,
})

// Empty detail pane, shown when no conversation is selected. Perm + plan gating
// live in the parent layout route (`route.tsx`). Carries the same learning-science
// doodle wallpaper as the thread so the pane never reads as blank.
function ConversationsIndex() {
  const { t } = useTranslation()
  return (
    <div className="relative isolate flex flex-1 flex-col">
      <ChatBackground />
      <EmptyState
        icon={MessagesSquareIcon}
        title={t("conversations.detail.empty.title")}
        description={t("conversations.detail.empty.description")}
        className="flex-1 justify-center border-0 bg-transparent ring-0"
      />
    </div>
  )
}

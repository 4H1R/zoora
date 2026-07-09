import { createFileRoute } from "@tanstack/react-router"
import { MessagesSquareIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ChatUpgrade } from "@/components/org/conversations/chat-upgrade"
import { useConversationPermissions } from "@/components/org/conversations/use-conversation-permissions"
import { EmptyState } from "@/components/ui/empty-state"
import { Spinner } from "@/components/ui/spinner"
import { useOrgGuard } from "@/lib/access"
import { FEATURE, useHasFeature } from "@/lib/entitlements"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/conversations/")({
  head: () => orgHead("org.nav.conversations"),
  component: ConversationsPage,
})

function ConversationsPage() {
  const { t } = useTranslation()
  // Perm gate: participants (view) and org chat admins (manage) both pass;
  // anyone else is bounced to the dashboard.
  const allowed = useOrgGuard(["conversations:view", "conversations:manage"])
  const { canUpgrade } = useConversationPermissions()
  // Plan gate — authoritative feature snapshot from /users/me/entitlements.
  const { enabled: chatEnabled, isLoading } = useHasFeature(FEATURE.chat)

  if (!allowed) return null

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center py-16">
        <Spinner className="text-muted-foreground size-6" />
      </div>
    )
  }

  // Plan doesn't include chat → paywall. The manage-only nav exemption means an
  // org admin without the plan lands here; view-only users reach it only by
  // deep-linking (their nav item stays hidden until the plan is active).
  if (!chatEnabled) return <ChatUpgrade canUpgrade={canUpgrade} />

  // TODO(conversations): replace this scaffold with the real messaging UI
  // (conversation list + thread). The API bindings already exist in
  // src/api/conversations. Perms + plan gating are done.
  return (
    <div className="flex flex-1 flex-col">
      <EmptyState
        icon={MessagesSquareIcon}
        title={t("conversations.placeholder.title")}
        description={t("conversations.placeholder.description")}
        className="flex-1 justify-center"
      />
    </div>
  )
}

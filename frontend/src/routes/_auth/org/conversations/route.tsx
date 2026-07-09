import { createFileRoute, Outlet, useParams } from "@tanstack/react-router"

import { ChatUpgrade } from "@/components/org/conversations/chat-upgrade"
import { ConversationSidebar } from "@/components/org/conversations/conversation-list"
import { useConversationPermissions } from "@/components/org/conversations/use-conversation-permissions"
import { Spinner } from "@/components/ui/spinner"
import { useOrgGuard } from "@/lib/access"
import { FEATURE, useHasFeature } from "@/lib/entitlements"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/conversations")({
  head: () => orgHead("org.nav.conversations"),
  component: ConversationsLayout,
})

/**
 * Layout for the conversations screen. Owns the perm gate (view|manage) and plan
 * gate (chat feature), then renders a two-pane master-detail: the conversation
 * list sidebar plus the routed detail (`<Outlet/>`). On small screens only one
 * pane shows at a time — the list, or the detail once a conversation is selected.
 */
function ConversationsLayout() {
  const allowed = useOrgGuard(["conversations:view", "conversations:manage"])
  const { canUpgrade } = useConversationPermissions()
  const { enabled: chatEnabled, isLoading } = useHasFeature(FEATURE.chat)

  const params = useParams({ strict: false }) as { conversationId?: string }
  const hasDetail = Boolean(params.conversationId)

  if (!allowed) return null

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center py-16">
        <Spinner className="text-muted-foreground size-6" />
      </div>
    )
  }

  if (!chatEnabled) return <ChatUpgrade canUpgrade={canUpgrade} />

  return (
    <div className="bg-card flex flex-1 overflow-hidden rounded-xl border">
      <aside
        className={cn(
          "bg-card flex w-full flex-col border-e md:w-80 md:shrink-0 xl:w-96",
          hasDetail && "hidden md:flex"
        )}
      >
        <ConversationSidebar />
      </aside>
      <main className={cn("min-w-0 flex-1 flex-col", hasDetail ? "flex" : "hidden md:flex")}>
        <Outlet />
      </main>
    </div>
  )
}

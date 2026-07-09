import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useChatWs } from "@/components/org/conversations/chat-provider"
import { useConversations } from "@/components/org/conversations/use-conversations"
import { Button } from "@/components/ui/button"

// `?msg=<uuid>` is used in a later phase to deep-link/jump to a specific message.
// Validate it here (drop anything malformed) so the search param is trustworthy.
const searchSchema = z.object({
  msg: z.string().uuid().optional().catch(undefined),
})

export const Route = createFileRoute("/_auth/org/conversations/$conversationId")({
  validateSearch: searchSchema,
  component: ConversationDetail,
})

function ConversationDetail() {
  const { t } = useTranslation()
  const { conversationId } = Route.useParams()
  const { setFocusedConvId, join, leave } = useChatWs()
  const { data: conversations } = useConversations()

  const conversation = conversations?.find((c) => c.id === conversationId)

  // Wire realtime focus + room membership: focusing the thread lets the WS reducer
  // suppress its unread bumps, and joining the room subscribes to its full-payload
  // message stream. Torn down (and focus cleared) on unmount / conversation change.
  useEffect(() => {
    setFocusedConvId(conversationId)
    join(conversationId)
    return () => {
      leave(conversationId)
      setFocusedConvId(null)
    }
    // Context fns are stable for the ChatProvider's lifetime; re-run only when the
    // selected conversation changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [conversationId])

  return (
    <div className="flex flex-1 flex-col">
      <div className="flex items-center gap-2 border-b px-4 py-3">
        <Button
          variant="ghost"
          size="icon-sm"
          className="md:hidden"
          aria-label={t("common.back")}
          render={<Link to="/org/conversations" />}
        >
          <ArrowLeftIcon className="rtl:rotate-180" />
        </Button>
        <p className="min-w-0 truncate text-sm font-semibold">
          {conversation?.name ?? conversationId}
        </p>
      </div>

      {/* Phase 5 replaces this with the real message thread. */}
      <div className="text-muted-foreground flex flex-1 items-center justify-center p-6 text-sm">
        {t("conversations.detail.comingSoon")}
      </div>
    </div>
  )
}

import { createFileRoute } from "@tanstack/react-router"
import { useEffect } from "react"
import { z } from "zod"

import { ChatThread } from "@/components/org/conversations/chat-thread"
import { useChatWs } from "@/components/org/conversations/chat-provider"

// `?msg=<uuid>` deep-links/jumps to a specific message. Validate it here (drop
// anything malformed) so the search param is trustworthy for the thread.
const searchSchema = z.object({
  msg: z.string().uuid().optional().catch(undefined),
})

export const Route = createFileRoute("/_auth/org/conversations/$conversationId")({
  validateSearch: searchSchema,
  component: ConversationDetail,
})

function ConversationDetail() {
  const { conversationId } = Route.useParams()
  const { msg } = Route.useSearch()
  const { setFocusedConvId, join, leave } = useChatWs()

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

  return <ChatThread convId={conversationId} aroundMessageId={msg} />
}

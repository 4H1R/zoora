import { createFileRoute } from "@tanstack/react-router"
import { useEffect } from "react"
import { z } from "zod"

import { ChatThread } from "@/components/org/conversations/chat-thread"
import { useChatWs } from "@/components/org/conversations/chat-provider"

// `?msg=<uuid>` deep-links/jumps to a specific message. Validate it here (drop
// anything malformed) so the search param is trustworthy for the thread.
const searchSchema = z.object({
  msg: z.uuid().optional().catch(undefined),
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

  // Key on conversation + `?msg` so a jump to an UNLOADED message (which sets
  // `?msg`) remounts the thread, resetting its scroll/jump state. The infinite
  // query itself keys on the around-id (`chatKeys.messagesAround`), so each jump
  // target gets its own cache seeded with `{around}`; clearing `?msg` remounts
  // back onto the live base cache (`chatKeys.messages`) that the WS reducer feeds.
  return <ChatThread key={`${conversationId}:${msg ?? ""}`} convId={conversationId} aroundMessageId={msg} />
}

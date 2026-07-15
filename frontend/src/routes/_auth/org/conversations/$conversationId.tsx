import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { useGetUsersMe } from "@/api/users/users"
import { useChatWs } from "@/components/org/conversations/chat-provider"
import { ChatThread } from "@/components/org/conversations/chat-thread"

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
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: meData } = useGetUsersMe()
  const selfId = (meData?.status === 200 && meData.data.data?.id) || null
  const { setFocusedConvId, join, leave, subscribe } = useChatWs()

  // If THIS conversation is deleted (by us on another device, or by another
  // member), the reducer drops it from the list cache; here we also bounce the
  // viewer back to the list so they aren't stranded on a dead thread — with a
  // toast explaining the redirect. The actor is suppressed (`deleted_by`): they
  // already saw a delete-success toast from the settings sheet.
  useEffect(() => {
    return subscribe((e) => {
      if (e.type !== "conversation_deleted") return
      const data = e.data as { conversation_id?: string; deleted_by?: string }
      if (data.conversation_id !== conversationId) return
      if (data.deleted_by !== selfId) toast.info(t("conversations.deleted.toast"))
      navigate({ to: "/org/conversations" })
    })
    // subscribe/navigate are stable; re-wire only when the open conversation
    // or signed-in user changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [conversationId, selfId])

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

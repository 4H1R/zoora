import type { ChatMessage } from "./lib/messages"
import type { postConversationsIdMessagesResponse } from "@/api/conversations/conversations"
import type { GithubCom4H1RZooraInternalDomainSendConversationMessageDTO as SendMessageDTO } from "@/api/model"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { useAccess } from "react-access-engine"
import { uuidv7 } from "uuidv7"

import { usePostConversationsIdMessages } from "@/api/conversations/conversations"
import { useGetUsersMe } from "@/api/users/users"

import { insertOptimistic, markStatus, replaceMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"

type MessagesCache = InfiniteData<ChatMessage[]>

export interface SendMessageInput {
  content: string
  replyToMessageId?: string
  mentions?: string[]
  mediaIds?: string[]
}

// Module-level stash of the original send input, keyed by the optimistic
// message id. The optimistic `ChatMessage` bubble intentionally doesn't carry
// `mentions` (or the reply target beyond `reply_to_message_id`), so `retry`
// reads them back from here to re-POST faithfully — otherwise a retry would
// silently drop mentions and the server would never fire mention
// notifications. Mirrors the attachment upload/pending-registry pattern;
// cleared once the send confirms.
const pendingSends = new Map<string, SendMessageInput>()

/**
 * Build the input for a retry POST: prefer the stashed original send input
 * (which carries `mentions` + the reply target), falling back to the failed
 * bubble's own cached fields if the stash is gone (e.g. after a reload dropped
 * module state). Mentions can only be recovered from the stash.
 */
export function resolveRetryInput(
  stashed: SendMessageInput | undefined,
  failed: Pick<ChatMessage, "content" | "reply_to_message_id">
): SendMessageInput {
  return {
    content: stashed?.content ?? failed.content ?? "",
    replyToMessageId: stashed?.replyToMessageId ?? failed.reply_to_message_id,
    mentions: stashed?.mentions,
  }
}

/**
 * Build the send DTO, omitting empty/undefined optional fields so the server
 * never sees a stray `reply_to_message_id: undefined` or empty arrays.
 */
function buildDto(id: string, input: SendMessageInput): SendMessageDTO {
  const dto: SendMessageDTO = { id, content: input.content }
  if (input.replyToMessageId) dto.reply_to_message_id = input.replyToMessageId
  if (input.mentions && input.mentions.length > 0) dto.mentions = input.mentions
  if (input.mediaIds && input.mediaIds.length > 0) dto.media_ids = input.mediaIds
  return dto
}

/**
 * Pull the server message out of the 201 response. `customInstance` throws on
 * every >= 400 status, so `onSuccess` only ever sees the success variant, but
 * we narrow on `status` to keep the union types honest.
 */
function serverMessage(res: postConversationsIdMessagesResponse): ChatMessage | undefined {
  return res.status === 201 ? (res.data.data as ChatMessage | undefined) : undefined
}

/**
 * Optimistic message-send hook. Returns `send` (compose + POST a brand-new
 * message) and `retry` (re-POST a previously failed bubble with the SAME id).
 *
 * The optimistic bubble is inserted into the `chatKeys.messages(convId)`
 * infinite cache immediately (keyed by a client-generated uuidv7). Because the
 * id is client-supplied and the server treats it idempotently, the WS
 * `new_message` echo, the mutation response, and any retry all converge on that
 * single id — reconciled in place, never duplicated. The composer owns clearing
 * its own input; this hook only touches the cache.
 */
export function useSendMessage(convId: string) {
  const queryClient = useQueryClient()
  const { user } = useAccess()
  const { data: meData } = useGetUsersMe()

  const selfId = user.id
  const selfName = (meData?.status === 200 && meData.data.data?.name) || ""

  const mutation = usePostConversationsIdMessages()

  const key = chatKeys.messages(convId)

  function post(id: string, input: SendMessageInput) {
    mutation.mutate(
      { id: convId, data: buildDto(id, input) },
      {
        onSuccess: (res) => {
          const server = serverMessage(res)
          if (!server) return
          pendingSends.delete(id)
          queryClient.setQueryData<MessagesCache>(key, (old) => replaceMessage(old, server))
        },
        onError: () => {
          queryClient.setQueryData<MessagesCache>(key, (old) => markStatus(old, id, "failed"))
        },
      }
    )
  }

  function send(input: SendMessageInput) {
    const id = uuidv7()
    const optimistic: ChatMessage = {
      id,
      conversation_id: convId,
      sender_id: selfId,
      sender: { id: selfId, name: selfName },
      content: input.content,
      reply_to_message_id: input.replyToMessageId,
      // Optimistic media rendering is Phase 8; keep empty for now.
      media_ids: [],
      created_at: new Date().toISOString(),
      _status: "sending",
    }
    // Stash the full input so a later retry can re-POST with the mentions the
    // optimistic bubble doesn't carry.
    pendingSends.set(id, input)
    queryClient.setQueryData<MessagesCache>(key, (old) => insertOptimistic(old, optimistic))
    post(id, input)
  }

  function retry(id: string) {
    const cache = queryClient.getQueryData<MessagesCache>(key)
    const failed = cache?.pages.flat().find((m) => m.id === id)
    if (!failed) return
    // Flip the bubble back to "sending" and re-POST with the SAME id — the
    // server dedups on the client-supplied id, so a retry never duplicates.
    // Recover mentions/reply from the stash so they survive the retry.
    queryClient.setQueryData<MessagesCache>(key, (old) => markStatus(old, id, "sending"))
    post(id, resolveRetryInput(pendingSends.get(id), failed))
  }

  return { send, retry }
}

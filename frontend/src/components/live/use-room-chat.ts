import { useQueryClient } from "@tanstack/react-query"

import {
  getGetChatsChatIdMessagesQueryKey,
  useDeleteChatsChatIdMessagesMessageId,
  useGetChatsChatIdMessages,
  usePostChatsChatIdMessages,
} from "@/api/chat/chat"
import type { GithubCom4H1RZooraInternalDomainMessage } from "@/api/model"

// Backend-persisted room chat. Polls for new messages (simple + works for all
// roles regardless of data-channel grants). Returns a shape the panel renders.
export interface RoomChatMessage {
  id: string
  content: string
  senderName: string
  senderId?: string
  createdAt: string
  isDeleted: boolean
}

export function useRoomChat(chatId: string | undefined) {
  const queryClient = useQueryClient()

  const { data } = useGetChatsChatIdMessages(
    chatId ?? "",
    undefined,
    {
      query: {
        enabled: !!chatId,
        refetchInterval: 4000,
      },
    },
  )

  const rawMessages =
    data?.status === 200
      ? ((data.data.data?.items ?? []) as GithubCom4H1RZooraInternalDomainMessage[])
      : []

  const messages: RoomChatMessage[] = rawMessages
    .map((msg) => ({
      id: msg.id ?? "",
      content: msg.content ?? "",
      senderName: msg.sender?.name ?? msg.sender_id ?? "—",
      senderId: msg.sender_id,
      createdAt: msg.created_at ?? "",
      isDeleted: false,
    }))
    .sort((a, b) => {
      if (!a.createdAt) return -1
      if (!b.createdAt) return 1
      return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    })

  const invalidate = () => {
    if (!chatId) return
    void queryClient.invalidateQueries({
      queryKey: getGetChatsChatIdMessagesQueryKey(chatId),
    })
  }

  const sendMutation = usePostChatsChatIdMessages({
    mutation: {
      onSuccess: invalidate,
    },
  })

  const deleteMutation = useDeleteChatsChatIdMessagesMessageId({
    mutation: {
      onSuccess: invalidate,
    },
  })

  const send = (content: string) => {
    if (!chatId) return
    sendMutation.mutate({
      chatId,
      data: { content, message_type: "text" },
    })
  }

  const deleteMessage = (messageId: string) => {
    if (!chatId) return
    deleteMutation.mutate({ chatId, messageId })
  }

  return {
    messages,
    send,
    isSending: sendMutation.isPending,
    deleteMessage,
    count: messages.length,
  }
}

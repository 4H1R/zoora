import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"

import {
  getGetChatsChatIdMessagesQueryKey,
  useDeleteChatsChatIdMessagesMessageId,
  useGetChatsChatIdMessages,
  usePostChatsChatIdMessages,
} from "@/api/chat/chat"
import type { GithubCom4H1RZooraInternalDomainLiveRoomMessage } from "@/api/model"

import { decodeRoomEvent } from "./room-events"
import { useRoomChannel } from "./use-room-channel"

// Backend-persisted room chat. History (and late-join catch-up) comes from a
// slow GET; new messages arrive in realtime over the LiveKit data channel, which
// the backend fans out server-side on every send/delete. Server-side fanout
// means all roles receive instantly with no per-client publish grant — the GET
// stays only as a mount-time backfill + reconnect safety net.
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

  // Messages received live over the data channel, keyed by id (dedup). Survives
  // panel unmount because this hook is mounted at room level (RoomShell), not in
  // the chat tab — matching the pattern in use-room-polls.ts.
  const [liveMessages, setLiveMessages] = useState<Record<string, RoomChatMessage>>({})
  const [deletedIds, setDeletedIds] = useState<Record<string, true>>({})

  useRoomChannel(undefined, (msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return

    if (event.type === "chat_message") {
      const d = event.data
      setLiveMessages((prev) => ({
        ...prev,
        [d.id]: {
          id: d.id,
          content: d.content,
          senderName: d.sender?.name ?? "—",
          senderId: d.sender_id,
          createdAt: d.created_at,
          isDeleted: false,
        },
      }))
    } else if (event.type === "chat_message_deleted") {
      setDeletedIds((prev) => ({ ...prev, [event.data.id]: true }))
    }
  })

  const { data } = useGetChatsChatIdMessages(
    chatId ?? "",
    undefined,
    {
      query: {
        enabled: !!chatId,
        // Realtime arrives via the data channel; this slow poll only backfills
        // history on mount and recovers packets missed during a reconnect.
        refetchInterval: 30000,
      },
    },
  )

  const rawMessages =
    data?.status === 200
      ? ((data.data.data?.items ?? []) as GithubCom4H1RZooraInternalDomainLiveRoomMessage[])
      : []

  // Merge persisted history with live messages, deduping by id. The persisted
  // copy wins when both exist (authoritative sender, edits, etc.).
  const merged = new Map<string, RoomChatMessage>()
  for (const m of Object.values(liveMessages)) {
    merged.set(m.id, m)
  }
  for (const msg of rawMessages) {
    const id = msg.id ?? ""
    if (!id) continue
    merged.set(id, {
      id,
      content: msg.content ?? "",
      senderName: msg.sender?.name ?? "—",
      senderId: msg.sender_id,
      createdAt: msg.created_at ?? "",
      isDeleted: false,
    })
  }

  const messages: RoomChatMessage[] = [...merged.values()]
    .filter((m) => !deletedIds[m.id])
    .sort((a, b) => {
      if (!a.createdAt) return -1
      if (!b.createdAt) return 1
      return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    })

  // The sender never receives its own data-channel packet (LiveKit doesn't echo
  // to the publisher), so refetch to surface the sender's own message/delete.
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
      onSuccess: (_res, vars) => {
        setDeletedIds((prev) => ({ ...prev, [vars.messageId]: true }))
        invalidate()
      },
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

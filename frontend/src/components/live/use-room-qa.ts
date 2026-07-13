import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"

import type { GithubCom4H1RZooraInternalDomainQAQuestionView } from "@/api/model"
import {
  getGetQaQueryKey,
  useDeleteQaId,
  useGetQa,
  usePostQa,
  usePostQaIdDismiss,
  usePostQaIdReopen,
  usePostQaIdResolve,
  usePostQaIdVote,
} from "@/api/qa/qa"

import { decodeRoomEvent } from "./room-events"
import { useRoomChannel } from "./use-room-channel"

// Audience Q&A for a live room. History (and late-join catch-up) comes from a
// slow GET; the backend fans out lightweight events over the LiveKit data
// channel on every ask/vote/status change. Created events carry a full row so
// other participants see a new question instantly; vote/status events are cheap
// overlays applied on top of the authoritative GET. Matches use-room-chat.ts.
const QA_MODEL_TYPE = "live_session"

// Sentinel status emitted by the backend on delete (qa_status_changed).
const QA_STATUS_DELETED = "deleted"

export interface RoomQuestion {
  id: string
  text: string
  authorName: string
  authorId: string
  status: string // "open" | "resolved" | "dismissed"
  voteCount: number
  votedByMe: boolean
  createdAt: string
}

function errorStatus(err: unknown): number | undefined {
  return (
    (err as { status?: number })?.status ??
    (err as { response?: { status?: number } })?.response?.status
  )
}

export function useRoomQa(liveId: string) {
  const queryClient = useQueryClient()

  // Questions received live over the data channel before the GET catches up,
  // plus per-question vote/status overlays. Survive panel unmount because this
  // hook is mounted at room level (RoomShell), not in the Q&A tab.
  const [liveCreated, setLiveCreated] = useState<Record<string, RoomQuestion>>({})
  const [voteCounts, setVoteCounts] = useState<Record<string, number>>({})
  const [statusOverrides, setStatusOverrides] = useState<Record<string, string>>({})

  useRoomChannel(undefined, (msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return

    if (event.type === "qa_question_created") {
      const d = event.data
      setLiveCreated((prev) => ({
        ...prev,
        [d.id]: {
          id: d.id,
          text: d.text,
          authorName: d.author_name,
          authorId: d.user_id,
          status: d.status,
          voteCount: d.vote_count ?? 0,
          // A broadcast can't know the receiver's own vote; the GET reconciles.
          votedByMe: false,
          createdAt: d.created_at,
        },
      }))
    } else if (event.type === "qa_question_voted") {
      setVoteCounts((prev) => ({ ...prev, [event.data.id]: event.data.vote_count }))
    } else if (event.type === "qa_status_changed") {
      setStatusOverrides((prev) => ({ ...prev, [event.data.id]: event.data.status }))
    }
  })

  const params = { model_type: QA_MODEL_TYPE, model_id: liveId }
  const { data } = useGetQa(params, {
    query: {
      enabled: !!liveId,
      // Realtime arrives via the data channel; this slow poll only backfills on
      // mount and recovers packets missed during a reconnect.
      refetchInterval: 30000,
    },
  })

  const rawItems: GithubCom4H1RZooraInternalDomainQAQuestionView[] =
    data?.status === 200 ? (data.data.data?.items ?? []) : []

  // Merge persisted history with live-created questions, deduping by id. The
  // persisted copy wins (authoritative vote_count/voted_by_me/ordering).
  const merged = new Map<string, RoomQuestion>()
  for (const q of Object.values(liveCreated)) merged.set(q.id, q)
  for (const item of rawItems) {
    const id = item.id ?? ""
    if (!id) continue
    merged.set(id, {
      id,
      text: item.text ?? "",
      authorName: item.author_name ?? "—",
      authorId: item.user_id ?? "",
      status: item.status ?? "open",
      voteCount: item.vote_count ?? 0,
      votedByMe: item.voted_by_me ?? false,
      createdAt: item.created_at ?? "",
    })
  }

  const questions: RoomQuestion[] = [...merged.values()]
    .map((q) => ({
      ...q,
      voteCount: voteCounts[q.id] ?? q.voteCount,
      status: statusOverrides[q.id] ?? q.status,
    }))
    .filter((q) => q.status !== QA_STATUS_DELETED)
    // Match backend ordering: open first, then vote count desc, then oldest first.
    .sort((a, b) => {
      const ao = a.status === "open" ? 0 : 1
      const bo = b.status === "open" ? 0 : 1
      if (ao !== bo) return ao - bo
      if (b.voteCount !== a.voteCount) return b.voteCount - a.voteCount
      return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    })

  const openCount = questions.filter((q) => q.status === "open").length

  // The sender never receives its own data-channel packet (LiveKit doesn't echo
  // to the publisher), so refetch to surface the caller's own ask/vote/moderation.
  const invalidate = () => {
    if (!liveId) return
    void queryClient.invalidateQueries({ queryKey: getGetQaQueryKey(params) })
  }

  const askMutation = usePostQa({ mutation: { onSuccess: invalidate } })
  const voteMutation = usePostQaIdVote({ mutation: { onSuccess: invalidate } })
  const resolveMutation = usePostQaIdResolve({ mutation: { onSuccess: invalidate } })
  const dismissMutation = usePostQaIdDismiss({ mutation: { onSuccess: invalidate } })
  const reopenMutation = usePostQaIdReopen({ mutation: { onSuccess: invalidate } })
  const deleteMutation = useDeleteQaId({ mutation: { onSuccess: invalidate } })

  const ask = (text: string, onError?: (status?: number) => void) => {
    if (!liveId) return
    askMutation.mutate(
      { data: { model_type: QA_MODEL_TYPE, model_id: liveId, text } },
      { onError: (err) => onError?.(errorStatus(err)) },
    )
  }
  const vote = (id: string) => voteMutation.mutate({ id })
  const resolve = (id: string) => resolveMutation.mutate({ id })
  const dismiss = (id: string) => dismissMutation.mutate({ id })
  const reopen = (id: string) => reopenMutation.mutate({ id })
  const remove = (id: string) => deleteMutation.mutate({ id })

  return {
    questions,
    openCount,
    ask,
    isAsking: askMutation.isPending,
    vote,
    resolve,
    dismiss,
    reopen,
    remove,
  }
}

export type RoomQa = ReturnType<typeof useRoomQa>

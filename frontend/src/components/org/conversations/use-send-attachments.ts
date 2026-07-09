import type { ChatMessage, LocalAttachment } from "./lib/messages"
import type { PendingFile, PendingSendInput } from "./upload/pending-registry"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { useAccess } from "react-access-engine"
import { uuidv7 } from "uuidv7"

import {
  type postConversationsIdMessagesResponse,
  usePostConversationsIdMessages,
} from "@/api/conversations/conversations"
import { useGetUsersMe } from "@/api/users/users"
import type { GithubCom4H1RZooraInternalDomainSendConversationMessageDTO as SendMessageDTO } from "@/api/model"

import {
  allAttachmentsSucceeded,
  attachmentsOf,
  markAttachmentDone,
  markAttachmentError,
  removeAttachment,
  resetAttachmentUploading,
  resolvedMediaIds,
  updateAttachmentProgress,
} from "./lib/attachments"
import { insertOptimistic, markStatus, removeMessage, replaceMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { isImage } from "./upload/compress"
import {
  abortAllPending,
  cancelPending,
  clearPending,
  getPending,
  isAbortError,
  revokeAttachmentBlobs,
  setPending,
} from "./upload/pending-registry"
import { capFiles, uploadFile } from "./upload/upload-manager"

type MessagesCache = InfiniteData<ChatMessage[]>

// Keep blob previews alive briefly after the confirmed message swaps in so the
// real (download-url) images have time to decode before we release memory.
const BLOB_REVOKE_DELAY_MS = 30_000

export interface SendWithAttachmentsInput {
  content: string
  files: File[]
  replyToMessageId?: string
  mentions?: string[]
}

function buildDto(id: string, input: PendingSendInput, mediaIds: string[]): SendMessageDTO {
  const dto: SendMessageDTO = { id, content: input.content }
  if (input.replyToMessageId) dto.reply_to_message_id = input.replyToMessageId
  if (input.mentions && input.mentions.length > 0) dto.mentions = input.mentions
  if (mediaIds.length > 0) dto.media_ids = mediaIds
  return dto
}

function serverMessage(res: postConversationsIdMessagesResponse): ChatMessage | undefined {
  return res.status === 201 ? (res.data.data as ChatMessage | undefined) : undefined
}

/**
 * Telegram-style attachment send: insert the optimistic bubble immediately with
 * local blob/blurhash previews, upload the files in the background (progress
 * rings render off the cached `_attachments`), and only once EVERY upload
 * resolves fire the real POST with the resolved `media_ids` — reusing the same
 * client message id so the server treats it idempotently. Cancel and retry are
 * driven from the module-level pending registry so the message bubble (a
 * different hook instance) can reach the live controllers + original files.
 */
export function useSendAttachments(convId: string) {
  const queryClient = useQueryClient()
  const { user } = useAccess()
  const { data: meData } = useGetUsersMe()

  const selfId = user.id
  const selfName = (meData?.status === 200 && meData.data.data?.name) || ""

  const mutation = usePostConversationsIdMessages()
  const key = chatKeys.messages(convId)

  const setCache = (fn: (old: MessagesCache | undefined) => MessagesCache | undefined) =>
    queryClient.setQueryData<MessagesCache>(key, fn)
  const getCache = () => queryClient.getQueryData<MessagesCache>(key)

  // Fire the real message POST once uploads are done. On success reconcile the
  // server copy (dropping the local previews) and schedule the blob revoke; on
  // error flip the bubble to "failed" so Retry can re-run.
  function post(msgId: string, input: PendingSendInput, mediaIds: string[]) {
    mutation.mutate(
      { id: convId, data: buildDto(msgId, input, mediaIds) },
      {
        onSuccess: (res) => {
          const server = serverMessage(res)
          if (!server) return
          const previews = attachmentsOf(getCache(), msgId)
          setCache((old) => replaceMessage(old, server))
          clearPending(msgId)
          // Seamless swap: let the confirmed images load, then release blobs.
          setTimeout(() => revokeAttachmentBlobs(previews), BLOB_REVOKE_DELAY_MS)
        },
        onError: () => {
          setCache((old) => markStatus(old, msgId, "failed"))
        },
      }
    )
  }

  // Upload each pending file, wiring progress + terminal state into the cached
  // `_attachments`. Resolves once every file settles (fulfilled, errored, or
  // canceled) so `finalize` can decide the message's fate.
  function uploadAll(msgId: string, pendings: PendingFile[]) {
    const tasks = pendings.map((p) =>
      uploadFile(p.file, convId, {
        signal: p.controller.signal,
        onProgress: (prog) => setCache((old) => updateAttachmentProgress(old, msgId, p.localId, prog)),
      })
        .then((res) => {
          setCache((old) => markAttachmentDone(old, msgId, p.localId, res))
        })
        .catch((err) => {
          if (isAbortError(err)) {
            // Individual cancel: revoke its blob and drop it from the bubble.
            const current = attachmentsOf(getCache(), msgId).find((a) => a.localId === p.localId)
            if (current?.blobUrl) URL.revokeObjectURL(current.blobUrl)
            setCache((old) => removeAttachment(old, msgId, p.localId))
          } else {
            setCache((old) => markAttachmentError(old, msgId, p.localId))
          }
        })
    )
    Promise.all(tasks).then(() => finalize(msgId))
  }

  // Decide what happens once all uploads have settled.
  function finalize(msgId: string) {
    const atts = attachmentsOf(getCache(), msgId)
    const entry = getPending(msgId)

    // Everything was canceled → abort the send entirely (drop the bubble).
    if (atts.length === 0) {
      revokeAttachmentBlobs(atts)
      setCache((old) => removeMessage(old, msgId))
      clearPending(msgId)
      return
    }

    if (allAttachmentsSucceeded(atts)) {
      const input = entry?.input ?? { content: getMessageContent(msgId) }
      post(msgId, input, resolvedMediaIds(atts))
      return
    }

    // Some uploads failed → surface a failed bubble; Retry re-runs the failures.
    setCache((old) => markStatus(old, msgId, "failed"))
  }

  function getMessageContent(msgId: string): string {
    return getCache()?.pages.flat().find((m) => m.id === msgId)?.content ?? ""
  }

  function sendWithAttachments({ content, files, replyToMessageId, mentions }: SendWithAttachmentsInput) {
    const capped = capFiles(files)
    if (capped.length === 0) return

    const id = uuidv7()
    const attachments: LocalAttachment[] = capped.map((file) => ({
      localId: uuidv7(),
      name: file.name,
      contentType: file.type || "application/octet-stream",
      size: file.size,
      blobUrl: isImage(file) ? URL.createObjectURL(file) : undefined,
      blurhash: null,
      progress: 0,
      status: "uploading",
    }))

    const optimistic: ChatMessage = {
      id,
      conversation_id: convId,
      sender_id: selfId,
      sender: { id: selfId, name: selfName },
      content,
      reply_to_message_id: replyToMessageId,
      media_ids: [],
      created_at: new Date().toISOString(),
      _status: "sending",
      _attachments: attachments,
    }
    setCache((old) => insertOptimistic(old, optimistic))

    const pendings: PendingFile[] = capped.map((file, i) => ({
      localId: attachments[i].localId,
      file,
      controller: new AbortController(),
    }))
    setPending(id, { files: pendings, input: { content, replyToMessageId, mentions } })
    uploadAll(id, pendings)
  }

  // Cancel one in-flight upload from its bubble; aborting the last one aborts
  // the whole send (bubble removed by `finalize`).
  function cancelAttachment(msgId: string, localId: string) {
    cancelPending(msgId, localId)
  }

  // Retry a failed attachment bubble: re-upload only the errored attachments
  // (keeping the ones that already succeeded) with fresh controllers, then
  // re-run the settle → POST pipeline.
  function retry(msgId: string) {
    const entry = getPending(msgId)
    if (!entry) return
    const atts = attachmentsOf(getCache(), msgId)
    const failedIds = new Set(atts.filter((a) => a.status !== "done").map((a) => a.localId))
    if (failedIds.size === 0) return

    setCache((old) => markStatus(old, msgId, "sending"))

    const retried: PendingFile[] = entry.files
      .filter((f) => failedIds.has(f.localId))
      .map((f) => {
        setCache((old) => resetAttachmentUploading(old, msgId, f.localId))
        return { ...f, controller: new AbortController() }
      })

    // Keep already-succeeded pendings in the registry so `finalize` still sees
    // the full input; swap in fresh controllers for the retried ones.
    const nextFiles = entry.files.map((f) => retried.find((r) => r.localId === f.localId) ?? f)
    setPending(msgId, { ...entry, files: nextFiles })
    uploadAll(msgId, retried)
  }

  // Discard a failed attachment bubble: abort anything still running, revoke
  // blobs, and drop the bubble from the cache.
  function discard(msgId: string) {
    const previews = attachmentsOf(getCache(), msgId)
    abortAllPending(msgId)
    revokeAttachmentBlobs(previews)
    setCache((old) => removeMessage(old, msgId))
  }

  return { sendWithAttachments, cancelAttachment, retry, discard }
}

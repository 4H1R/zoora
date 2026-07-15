import type { UploadResult } from "../upload/upload-manager"
import type { ChatMessage, LocalAttachment } from "./messages"
import type { InfiniteData } from "@tanstack/react-query"

/**
 * Pure cache helpers for the client-only `_attachments` array carried on an
 * optimistic bubble. They mirror the shape of the `optimistic.ts` helpers
 * (`markStatus` et al.): each finds the message by id across the infinite
 * message cache and returns a NEW cache with that message's `_attachments`
 * transformed — or the original reference unchanged when nothing matched, so
 * React Query never sees a spurious identity change.
 */
type MessagesCache = InfiniteData<ChatMessage[]>

function clamp01(n: number): number {
  if (Number.isNaN(n)) return 0
  return Math.min(1, Math.max(0, n))
}

/**
 * Map the `_attachments` of the message with `msgId` through `update`,
 * preserving cache identity when the message (or its attachments) is absent.
 */
function updateAttachments(
  old: MessagesCache | undefined,
  msgId: string,
  update: (atts: LocalAttachment[]) => LocalAttachment[]
): MessagesCache | undefined {
  if (!old) return old
  let changed = false
  const pages = old.pages.map((page) => {
    const idx = page.findIndex((m) => m.id === msgId)
    if (idx === -1) return page
    const msg = page[idx]
    if (!msg._attachments) return page
    changed = true
    const next = page.slice()
    next[idx] = { ...msg, _attachments: update(msg._attachments) }
    return next
  })
  return changed ? { ...old, pages } : old
}

/** Set the upload fraction (0..1) of a single attachment. */
export function updateAttachmentProgress(
  old: MessagesCache | undefined,
  msgId: string,
  localId: string,
  progress: number
): MessagesCache | undefined {
  return updateAttachments(old, msgId, (atts) =>
    atts.map((a) => (a.localId === localId ? { ...a, progress: clamp01(progress) } : a))
  )
}

/**
 * Mark a single attachment uploaded: fold in the resolved media id, dimensions
 * and (server-agreeing) blurhash, pin progress to 1 and flip status to "done".
 */
export function markAttachmentDone(
  old: MessagesCache | undefined,
  msgId: string,
  localId: string,
  result: UploadResult
): MessagesCache | undefined {
  return updateAttachments(old, msgId, (atts) =>
    atts.map((a) =>
      a.localId === localId
        ? {
            ...a,
            status: "done",
            progress: 1,
            mediaId: result.mediaId,
            blurhash: result.blurhash ?? a.blurhash,
            width: result.width ?? a.width,
            height: result.height ?? a.height,
          }
        : a
    )
  )
}

/** Flip a single attachment to the "error" status (upload failed). */
export function markAttachmentError(
  old: MessagesCache | undefined,
  msgId: string,
  localId: string
): MessagesCache | undefined {
  return updateAttachments(old, msgId, (atts) =>
    atts.map((a) => (a.localId === localId ? { ...a, status: "error" } : a))
  )
}

/** Reset a single attachment back to a fresh "uploading" state (for retry). */
export function resetAttachmentUploading(
  old: MessagesCache | undefined,
  msgId: string,
  localId: string
): MessagesCache | undefined {
  return updateAttachments(old, msgId, (atts) =>
    atts.map((a) => (a.localId === localId ? { ...a, status: "uploading", progress: 0 } : a))
  )
}

/** Drop a single attachment entirely (individual cancel). */
export function removeAttachment(
  old: MessagesCache | undefined,
  msgId: string,
  localId: string
): MessagesCache | undefined {
  return updateAttachments(old, msgId, (atts) => atts.filter((a) => a.localId !== localId))
}

/**
 * True when there is at least one attachment and EVERY attachment finished
 * uploading with a resolved media id — the gate for firing the real send.
 */
export function allAttachmentsSucceeded(attachments: LocalAttachment[]): boolean {
  return attachments.length > 0 && attachments.every((a) => a.status === "done" && !!a.mediaId)
}

/** Resolved media ids, in order, for every "done" attachment. */
export function resolvedMediaIds(attachments: LocalAttachment[]): string[] {
  const ids: string[] = []
  for (const a of attachments) {
    if (a.status === "done" && a.mediaId) ids.push(a.mediaId)
  }
  return ids
}

/** Read the (fresh) attachments off the cached message, or `[]` if absent. */
export function attachmentsOf(old: MessagesCache | undefined, msgId: string): LocalAttachment[] {
  const msg = old?.pages.flat().find((m) => m.id === msgId)
  return msg?._attachments ?? []
}

/**
 * How to retry a failed attachment bubble.
 * - `resend: true`  → every upload already succeeded, so the message POST itself
 *   is what failed; re-fire the POST with the already-resolved `mediaIds` (no
 *   re-upload). This is the path that was previously a dead-end.
 * - `resend: false` → at least one upload failed/errored; re-upload just the
 *   `failedIds`, then run the settle → POST pipeline again.
 */
export interface AttachmentRetryPlan {
  resend: boolean
  failedIds: string[]
  mediaIds: string[]
}

export function planAttachmentRetry(attachments: LocalAttachment[]): AttachmentRetryPlan {
  const failedIds = attachments.filter((a) => a.status !== "done").map((a) => a.localId)
  // Only re-send directly when there is something to send AND nothing to retry.
  if (failedIds.length === 0 && attachments.length > 0) {
    return { resend: true, failedIds: [], mediaIds: resolvedMediaIds(attachments) }
  }
  return { resend: false, failedIds, mediaIds: [] }
}

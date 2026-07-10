import type { LocalAttachment } from "../lib/messages"

/**
 * Module-level registry of in-flight / retryable attachment uploads, keyed by
 * the optimistic message id.
 *
 * The upload orchestration lives in the `useSendAttachments` hook, but its
 * refs are scoped to whichever component instance called `sendWithAttachments`.
 * Cancel and retry, however, are triggered from a DIFFERENT instance (the
 * message bubble). React state can't cross that boundary and `File`/
 * `AbortController` objects can't live in the React Query cache — so the raw
 * upload material (plus the original send input) is parked here, in module
 * scope, reachable from any instance (they all share the one QueryClient for
 * the cache side).
 */
export interface PendingFile {
  localId: string
  file: File
  /** Live controller for the current attempt; replaced on retry. */
  controller: AbortController
}

export interface PendingSendInput {
  content: string
  replyToMessageId?: string
  mentions?: string[]
  /** Telegram "Send as a document": force file-chip rendering for every media. */
  asDocument?: boolean
}

interface PendingEntry {
  files: PendingFile[]
  input: PendingSendInput
}

const registry = new Map<string, PendingEntry>()

export function setPending(msgId: string, entry: PendingEntry): void {
  registry.set(msgId, entry)
}

export function getPending(msgId: string): PendingEntry | undefined {
  return registry.get(msgId)
}

export function clearPending(msgId: string): void {
  registry.delete(msgId)
}

/** Abort a single file's upload and drop it from the pending set. */
export function cancelPending(msgId: string, localId: string): void {
  const entry = registry.get(msgId)
  if (!entry) return
  const target = entry.files.find((f) => f.localId === localId)
  target?.controller.abort()
  const rest = entry.files.filter((f) => f.localId !== localId)
  if (rest.length === 0) registry.delete(msgId)
  else registry.set(msgId, { ...entry, files: rest })
}

/** Abort every in-flight upload for a message and forget it. */
export function abortAllPending(msgId: string): void {
  const entry = registry.get(msgId)
  if (!entry) return
  for (const f of entry.files) f.controller.abort()
  registry.delete(msgId)
}

/** Revoke every object URL held by a message's attachment previews. */
export function revokeAttachmentBlobs(attachments: LocalAttachment[] | undefined): void {
  if (!attachments) return
  for (const a of attachments) {
    if (a.blobUrl) URL.revokeObjectURL(a.blobUrl)
  }
}

/** True for an AbortController-driven cancellation (vs. a genuine failure). */
export function isAbortError(reason: unknown): boolean {
  return reason instanceof DOMException && reason.name === "AbortError"
}

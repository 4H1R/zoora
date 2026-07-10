import { postMediaPresign } from "@/api/media/media"
import type { GithubCom4H1RZooraInternalDomainPresignUploadDTO } from "@/api/model"

import { compressImage, isImage } from "./compress"
import { encodeBlurhash, imageDimensions } from "./blurhash"

// Polymorphic media identifiers — MUST match the backend constants in
// internal/domain/media.go (MediaModelConversation / MediaCollectionAttach).
export const MEDIA_MODEL_CONVERSATION = "conversation"
export const MEDIA_COLLECTION_ATTACHMENTS = "attachments"

// The message send endpoint accepts at most 20 media_ids.
export const MAX_MEDIA_PER_MESSAGE = 20

export interface UploadResult {
  mediaId: string
  blurhash: string | null
  width?: number
  height?: number
  /** Original display name, independent of any compression rename. */
  name: string
  /** Mime type of the bytes actually uploaded. */
  contentType: string
  /** Byte size of the bytes actually uploaded. */
  size: number
}

export interface UploadOptions {
  onProgress?: (progress: number) => void
  signal?: AbortSignal
}

/**
 * Build the /media/presign request body for a conversation attachment. Pure —
 * reads only name/type/size, so it's trivially unit-testable with a fake File.
 */
export function buildPresignPayload(
  file: Pick<File, "name" | "type" | "size">,
  convId: string
): GithubCom4H1RZooraInternalDomainPresignUploadDTO {
  return {
    model_type: MEDIA_MODEL_CONVERSATION,
    model_id: convId,
    collection_name: MEDIA_COLLECTION_ATTACHMENTS,
    file_name: file.name,
    mime_type: file.type || "application/octet-stream",
    size: file.size,
  }
}

/**
 * Cap a file list to the per-message media limit. Files beyond the cap are
 * dropped (the caller is responsible for surfacing that to the user).
 */
export function capFiles<T>(files: T[], max = MAX_MEDIA_PER_MESSAGE): T[] {
  return files.slice(0, Math.max(0, max))
}

const abortError = () =>
  typeof DOMException === "function"
    ? new DOMException("aborted", "AbortError")
    : Object.assign(new Error("aborted"), { name: "AbortError" })

/**
 * PUT the bytes to a presigned S3/RustFS URL with a raw XMLHttpRequest so we
 * get upload progress and cancellation — neither of which fetch/redaxios
 * expose. The backend only ever issues presigned PUT URLs (no POST-form
 * variant), so PUT is the only method handled here.
 *
 * jsdom's XHR does not perform real network I/O, so this path is exercised in
 * the browser only, not in unit tests.
 */
export function putToPresignedUrl(
  url: string,
  body: Blob,
  contentType: string,
  opts: UploadOptions = {}
): Promise<void> {
  const { onProgress, signal } = opts
  return new Promise<void>((resolve, reject) => {
    if (signal?.aborted) {
      reject(abortError())
      return
    }

    const xhr = new XMLHttpRequest()
    xhr.open("PUT", url, true)
    if (contentType) xhr.setRequestHeader("Content-Type", contentType)

    const onAbort = () => xhr.abort()
    const cleanup = () => signal?.removeEventListener("abort", onAbort)

    xhr.upload.onprogress = (e: ProgressEvent) => {
      if (onProgress && e.lengthComputable && e.total > 0) {
        onProgress(e.loaded / e.total)
      }
    }
    xhr.onload = () => {
      cleanup()
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress?.(1)
        resolve()
      } else {
        reject(new Error(`upload failed: ${xhr.status}`))
      }
    }
    xhr.onerror = () => {
      cleanup()
      reject(new Error("upload network error"))
    }
    xhr.onabort = () => {
      cleanup()
      reject(abortError())
    }

    signal?.addEventListener("abort", onAbort)
    xhr.send(body)
  })
}

/**
 * Full single-file pipeline: compress (images), compute blurhash + dimensions,
 * presign, then upload the bytes with progress + cancellation support.
 */
export async function uploadFile(
  file: File,
  convId: string,
  opts: UploadOptions = {}
): Promise<UploadResult> {
  const { onProgress, signal } = opts
  if (signal?.aborted) throw abortError()

  const img = isImage(file)

  // Compression and blurhash/dimensions are independent — run them together.
  // Each best-effort helper resolves to null on failure, so the upload still
  // proceeds without a placeholder.
  const [compressed, blurhash, dims] = await Promise.all([
    compressImage(file),
    img ? encodeBlurhash(file) : Promise.resolve(null),
    img ? imageDimensions(file) : Promise.resolve(null),
  ])

  if (signal?.aborted) throw abortError()

  const payload = buildPresignPayload(compressed, convId)
  const res = await postMediaPresign(payload)
  if (res.status !== 201) {
    throw new Error(`presign failed: ${res.status}`)
  }
  const data = res.data.data
  const mediaId = data?.media?.id
  const uploadUrl = data?.upload_url
  if (!mediaId || !uploadUrl) {
    throw new Error("presign response missing media id or upload url")
  }

  await putToPresignedUrl(uploadUrl, compressed, payload.mime_type, { onProgress, signal })

  return {
    mediaId,
    blurhash,
    width: dims?.width,
    height: dims?.height,
    name: file.name,
    contentType: payload.mime_type,
    size: compressed.size,
  }
}

export interface UploadHandle {
  file: File
  /** Abort this file's upload individually. */
  controller: AbortController
  promise: Promise<UploadResult>
}

export interface UploadFilesResult {
  handles: UploadHandle[]
  /** Resolves once every (capped) upload settles — never rejects. */
  settled: Promise<PromiseSettledResult<UploadResult>[]>
}

/**
 * Upload up to MAX_MEDIA_PER_MESSAGE files at once. Files beyond the cap are
 * ignored. Each file gets its own AbortController so the UI can cancel one
 * without touching the others. `perFileCallbacks(file, index)` supplies the
 * per-file progress handler.
 */
export function uploadFiles(
  files: File[],
  convId: string,
  perFileCallbacks?: (file: File, index: number) => UploadOptions | undefined
): UploadFilesResult {
  const capped = capFiles(files)
  const handles: UploadHandle[] = capped.map((file, index) => {
    const controller = new AbortController()
    const cb = perFileCallbacks?.(file, index)
    const promise = uploadFile(file, convId, {
      onProgress: cb?.onProgress,
      signal: controller.signal,
    })
    return { file, controller, promise }
  })

  const settled = Promise.allSettled(handles.map((h) => h.promise))
  return { handles, settled }
}

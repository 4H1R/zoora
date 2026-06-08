import type { usePostMediaPresign } from "@/api/media/media"

export const OFFLINE_MODEL_TYPE = "offline_room"
export const OFFLINE_ATTACHMENTS_COLLECTION = "attachments"

type PresignAsync = ReturnType<typeof usePostMediaPresign>["mutateAsync"]

/**
 * Presign a media record for an offline room, then PUT the file to S3/RustFS.
 * Throws on any failure so callers can surface a toast.
 */
export async function uploadOfflineAttachment(
  presignAsync: PresignAsync,
  offlineId: string,
  file: File
): Promise<void> {
  const mime = file.type || "application/octet-stream"
  const res = await presignAsync({
    data: {
      model_type: OFFLINE_MODEL_TYPE,
      model_id: offlineId,
      collection_name: OFFLINE_ATTACHMENTS_COLLECTION,
      file_name: file.name,
      mime_type: mime,
      size: file.size,
    },
  })
  const uploadURL = res.status === 201 ? res.data.data?.upload_url : undefined
  if (!uploadURL) throw new Error("presign failed")
  const put = await fetch(uploadURL, {
    method: "PUT",
    body: file,
    headers: { "Content-Type": mime },
  })
  if (!put.ok) throw new Error(`upload failed: ${put.status}`)
}

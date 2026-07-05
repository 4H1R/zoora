import type { usePostMediaPresign } from "@/api/media/media"

type PresignAsync = ReturnType<typeof usePostMediaPresign>["mutateAsync"]

export const SHARED_MODEL_TYPE = "organization"
export const SHARED_COLLECTION = "shared"
export const MAX_SHARED_UPLOAD_BYTES = 200 * 1024 * 1024

/**
 * Presign a Shared-folder media record, then PUT the file to S3/RustFS.
 * Throws on any failure so callers can surface a toast.
 */
export async function uploadSharedFile(presignAsync: PresignAsync, orgId: string, file: File): Promise<void> {
  const mime = file.type || "application/octet-stream"
  const res = await presignAsync({
    data: {
      model_type: SHARED_MODEL_TYPE,
      model_id: orgId,
      collection_name: SHARED_COLLECTION,
      file_name: file.name,
      mime_type: mime,
      size: file.size,
    },
  })
  const uploadURL = res.status === 201 ? res.data.data?.upload_url : undefined
  if (!uploadURL) throw new Error("presign failed")
  const put = await fetch(uploadURL, { method: "PUT", body: file, headers: { "Content-Type": mime } })
  if (!put.ok) throw new Error(`upload failed: ${put.status}`)
}

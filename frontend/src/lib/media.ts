import { useQuery } from "@tanstack/react-query"

import { apiClient } from "@/api/mutator/custom-instance"

export const MEDIA_MODEL_PRACTICE = "practice"
export const MEDIA_MODEL_PRACTICE_SUBMISSION = "practice_submission"
export const MEDIA_COLLECTION_ATTACHMENTS = "attachments"

interface MediaInfo {
  name: string
  url: string | null
}

/**
 * Resolves a media record's display filename and a presigned download URL by id.
 * Shared by the attachment list (read) and the uploader (existing-item hydration).
 */
export function useMediaInfo(mediaID: string, enabled = true) {
  return useQuery<MediaInfo>({
    queryKey: ["media", "info", mediaID],
    enabled: enabled && !!mediaID,
    staleTime: 25 * 60 * 1000,
    queryFn: async () => {
      const metaRes = await apiClient(`/media/${mediaID}`, { method: "GET" })
      const meta = (metaRes.data as { data?: { file_name?: string; name?: string } }).data ?? {}
      const urlRes = await apiClient(`/media/${mediaID}/download-url`, { method: "GET" })
      const url = (urlRes.data as { data?: { url?: string } }).data?.url ?? null
      return { name: meta.file_name || meta.name || mediaID, url }
    },
  })
}

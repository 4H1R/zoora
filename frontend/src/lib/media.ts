import { useQuery } from "@tanstack/react-query"

import { getMediaId, getMediaIdDownloadUrl } from "@/api/media/media"

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
      const metaRes = await getMediaId(mediaID)
      const meta = metaRes.status === 200 ? (metaRes.data.data ?? {}) : {}
      const urlRes = await getMediaIdDownloadUrl(mediaID)
      const url = urlRes.status === 200 ? (urlRes.data.data?.url ?? null) : null
      return { name: meta.file_name || meta.name || mediaID, url }
    },
  })
}

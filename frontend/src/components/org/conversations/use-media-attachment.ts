import { useQuery } from "@tanstack/react-query"

import { getMediaId, getMediaIdDownloadUrl } from "@/api/media/media"

// Presigned download URLs are long-lived but not eternal — refetch lazily on a
// generous window rather than on every mount. Metadata (name/mime/size) never
// changes for a media id, so it can be cached indefinitely.
const URL_STALE_MS = 10 * 60 * 1000

/**
 * Lazily resolve a media id's presigned display URL. Keyed so every bubble
 * referencing the same id shares one request. Returns `null` when the id can't
 * be resolved (deleted / forbidden) so the caller can fall back gracefully.
 */
export function useMediaUrl(id: string) {
  return useQuery({
    queryKey: ["media", "download-url", id],
    queryFn: async () => {
      const res = await getMediaIdDownloadUrl(id)
      return res.status === 200 ? (res.data.data?.url ?? null) : null
    },
    enabled: !!id,
    staleTime: URL_STALE_MS,
  })
}

/**
 * Lazily resolve a media id's metadata (name, mime type, size) — used to decide
 * image vs. file rendering and to label file chips for CONFIRMED attachments.
 */
export function useMediaMeta(id: string) {
  return useQuery({
    queryKey: ["media", "meta", id],
    queryFn: async () => {
      const res = await getMediaId(id)
      return res.status === 200 ? (res.data.data ?? null) : null
    },
    enabled: !!id,
    staleTime: Infinity,
  })
}

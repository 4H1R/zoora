import { useQuery } from "@tanstack/react-query"
import { Loader2Icon } from "lucide-react"

import { getMediaIdDownloadUrl } from "@/api/media/media"
import { cn } from "@/lib/utils"

/**
 * SystemImage renders an anti-cheat question/option image (a server-rendered PNG
 * of the text). Unlike OptionImageThumb — a fixed 48px cropped avatar — these are
 * wide text strips, so they display full-width and object-contain. Resolves the
 * media id to a presigned URL, reusing the shared ["media","download-url",id]
 * cache key.
 */
export function SystemImage({ mediaID, className }: { mediaID: string; className?: string }) {
  const { data: url } = useQuery({
    queryKey: ["media", "download-url", mediaID],
    queryFn: async () => {
      const res = await getMediaIdDownloadUrl(mediaID)
      return res.status === 200 ? (res.data.data?.url ?? null) : null
    },
    staleTime: 30 * 60 * 1000,
  })

  if (!url) {
    return (
      <div className={cn("bg-muted flex h-16 w-full max-w-md items-center justify-center rounded-md", className)}>
        <Loader2Icon className="size-4 animate-spin opacity-60" />
      </div>
    )
  }
  return <img src={url} alt="" className={cn("max-w-full rounded-md", className)} />
}

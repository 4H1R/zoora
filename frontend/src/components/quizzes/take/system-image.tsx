import { useQuery } from "@tanstack/react-query"
import { Loader2Icon } from "lucide-react"

import { apiClient } from "@/api/mutator/custom-instance"
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
      const res = await apiClient(`/media/${mediaID}/download-url`, { method: "GET" })
      return (res.data as { data?: { url?: string } }).data?.url ?? null
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

import { useQuery } from "@tanstack/react-query"
import { DownloadIcon, FileIcon, Loader2Icon } from "lucide-react"

import { apiClient } from "@/api/mutator/custom-instance"
import { cn } from "@/lib/utils"

interface MediaMeta {
  file_name?: string
  name?: string
}

function useMediaChip(mediaID: string) {
  return useQuery({
    queryKey: ["media", "chip", mediaID],
    queryFn: async () => {
      const metaRes = await apiClient(`/media/${mediaID}`, { method: "GET" })
      const meta = (metaRes.data as { data?: MediaMeta }).data ?? {}
      const urlRes = await apiClient(`/media/${mediaID}/download-url`, { method: "GET" })
      const url = (urlRes.data as { data?: { url?: string } }).data?.url ?? null
      return { name: meta.file_name || meta.name || mediaID, url }
    },
    staleTime: 25 * 60 * 1000,
  })
}

function AttachmentChip({ mediaID }: { mediaID: string }) {
  const { data, isPending } = useMediaChip(mediaID)

  return (
    <a
      href={data?.url ?? undefined}
      target="_blank"
      rel="noreferrer"
      aria-disabled={isPending || !data?.url}
      className={cn(
        "border-foreground/10 bg-muted/40 text-foreground group inline-flex max-w-full items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs transition-colors",
        data?.url
          ? "hover:bg-muted hover:border-foreground/20 cursor-pointer"
          : "pointer-events-none opacity-60"
      )}
    >
      {isPending ? (
        <Loader2Icon className="size-3.5 shrink-0 animate-spin opacity-50" />
      ) : (
        <FileIcon className="size-3.5 shrink-0 opacity-70" />
      )}
      <span className="truncate">
        {isPending ? (
          <span className="bg-foreground/10 inline-block h-3 w-24 animate-pulse rounded" />
        ) : (
          (data?.name ?? mediaID)
        )}
      </span>
      <DownloadIcon
        className={cn(
          "size-3 shrink-0 opacity-60 transition-opacity",
          data?.url && "group-hover:opacity-100"
        )}
      />
    </a>
  )
}

interface MediaAttachmentListProps {
  mediaIds: string[]
  className?: string
}

export function MediaAttachmentList({ mediaIds, className }: MediaAttachmentListProps) {
  if (mediaIds.length === 0) return null
  return (
    <div className={className}>
      <div className="flex flex-wrap gap-1.5">
        {mediaIds.map((id) => (
          <AttachmentChip key={id} mediaID={id} />
        ))}
      </div>
    </div>
  )
}

import { DownloadIcon, FileIcon, Loader2Icon } from "lucide-react"

import { useMediaInfo } from "@/lib/media"
import { cn } from "@/lib/utils"

function AttachmentChip({ mediaID }: { mediaID: string }) {
  const { data, isPending } = useMediaInfo(mediaID)

  return (
    <a
      href={data?.url ?? undefined}
      target="_blank"
      rel="noreferrer"
      aria-disabled={isPending || !data?.url}
      tabIndex={data?.url ? undefined : -1}
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

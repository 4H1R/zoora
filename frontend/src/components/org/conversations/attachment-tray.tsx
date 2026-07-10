import { FileIcon, XIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { formatBytes } from "@/components/org/files/utils"
import { cn } from "@/lib/utils"

import { isImage } from "./upload/compress"

interface AttachmentTrayProps {
  files: File[]
  onRemove: (index: number) => void
}

/**
 * Pre-send tray shown inside the composer: one square cell per selected file —
 * a blob thumbnail for images, an icon + name/size chip for everything else —
 * each with a hover X to drop it before sending. Upload progress itself renders
 * later on the optimistic bubble (see `attachment-bubble`); here the files are
 * only staged.
 */
export function AttachmentTray({ files, onRemove }: AttachmentTrayProps) {
  const { t } = useTranslation()
  if (files.length === 0) return null

  return (
    <div className="flex flex-wrap gap-2 px-1 pb-1" aria-label={t("conversations.attachments.tray")}>
      {files.map((file, index) => (
        <TrayItem key={`${file.name}-${file.size}-${index}`} file={file} onRemove={() => onRemove(index)} />
      ))}
    </div>
  )
}

function TrayItem({ file, onRemove }: { file: File; onRemove: () => void }) {
  const { t } = useTranslation()
  const image = isImage(file)
  const [blobUrl, setBlobUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!image) return
    const url = URL.createObjectURL(file)
    setBlobUrl(url)
    return () => URL.revokeObjectURL(url)
  }, [file, image])

  return (
    <div className="group/tray relative">
      <div
        className={cn(
          "border-border bg-muted flex size-16 items-center justify-center overflow-hidden rounded-xl border",
          !image && "flex-col gap-1 p-1.5"
        )}
      >
        {image && blobUrl ? (
          <img src={blobUrl} alt={file.name} className="size-full object-cover" />
        ) : (
          <>
            <FileIcon className="text-muted-foreground size-5 shrink-0" />
            <span className="text-muted-foreground line-clamp-2 w-full text-center text-[0.625rem] leading-tight break-all">
              {file.name}
            </span>
          </>
        )}
      </div>
      {!image && (
        <span className="text-muted-foreground pointer-events-none absolute inset-x-0 -bottom-4 truncate text-center text-[0.625rem] tabular-nums">
          {formatBytes(file.size)}
        </span>
      )}
      <button
        type="button"
        onClick={onRemove}
        aria-label={t("conversations.attachments.remove")}
        className="bg-foreground/80 text-background hover:bg-foreground absolute -end-1.5 -top-1.5 flex size-5 items-center justify-center rounded-full opacity-0 shadow-sm transition group-hover/tray:opacity-100 focus-visible:opacity-100"
      >
        <XIcon className="size-3" />
      </button>
    </div>
  )
}

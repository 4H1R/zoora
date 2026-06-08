import { FileIcon, Loader2Icon, PaperclipIcon, Trash2Icon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { usePostMediaPresign } from "@/api/media/media"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export interface PendingAttachment {
  media_id: string
  name: string
  size: number
}

interface MediaAttachmentUploaderProps {
  value: PendingAttachment[]
  onChange: (next: PendingAttachment[]) => void
  modelType: string
  collectionName?: string
  modelId?: string
  disabled?: boolean
}

export function MediaAttachmentUploader({
  value,
  onChange,
  modelType,
  collectionName = "attachments",
  modelId,
  disabled,
}: MediaAttachmentUploaderProps) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [isUploading, setIsUploading] = useState(false)
  const presign = usePostMediaPresign()

  const handlePick = () => inputRef.current?.click()

  const handleFiles = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? [])
    e.target.value = ""
    if (files.length === 0) return

    setIsUploading(true)
    const uploaded: PendingAttachment[] = []
    try {
      for (const file of files) {
        const res = await presign.mutateAsync({
          data: {
            model_type: modelType,
            model_id: modelId ?? crypto.randomUUID(),
            collection_name: collectionName,
            file_name: file.name,
            mime_type: file.type || "application/octet-stream",
            size: file.size,
          },
        })
        if (res.status !== 201 || !res.data.data?.upload_url || !res.data.data?.media?.id) {
          throw new Error("presign failed")
        }
        const put = await fetch(res.data.data.upload_url, {
          method: "PUT",
          body: file,
          headers: { "Content-Type": file.type || "application/octet-stream" },
        })
        if (!put.ok) throw new Error(`upload failed: ${put.status}`)
        uploaded.push({ media_id: res.data.data.media.id, name: file.name, size: file.size })
      }
    } catch (err) {
      console.error(err)
      toast.error(t("media.uploader.uploadError"))
    } finally {
      if (uploaded.length > 0) onChange([...value, ...uploaded])
      setIsUploading(false)
    }
  }

  const handleRemove = (mediaID: string) => {
    onChange(value.filter((m) => m.media_id !== mediaID))
  }

  return (
    <div className="flex flex-col gap-2">
      <input ref={inputRef} type="file" multiple className="hidden" onChange={handleFiles} />

      {value.length > 0 && (
        <ul className="flex flex-col gap-1.5">
          {value.map((m) => (
            <li
              key={m.media_id}
              className="border-foreground/10 bg-muted/40 flex items-center gap-2.5 rounded-lg border px-3 py-2"
            >
              <FileIcon className="text-muted-foreground size-4 shrink-0" />
              <span className="flex-1 truncate text-sm">{m.name}</span>
              <span className="text-muted-foreground shrink-0 font-mono text-xs tabular-nums">
                {formatBytes(m.size)}
              </span>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                onClick={() => handleRemove(m.media_id)}
                disabled={disabled || isUploading}
              >
                <Trash2Icon />
              </Button>
            </li>
          ))}
        </ul>
      )}

      <Button
        type="button"
        variant="outline"
        size="sm"
        className={cn("self-start", value.length === 0 && "w-full justify-center border-dashed py-6")}
        onClick={handlePick}
        disabled={disabled || isUploading}
      >
        {isUploading ? <Loader2Icon className="animate-spin" /> : <PaperclipIcon />}
        {value.length === 0 ? t("media.uploader.empty") : t("media.uploader.add")}
      </Button>
    </div>
  )
}

import { PaperclipIcon, XIcon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getMediaIdDownloadUrl, postMediaPresign } from "@/api/media/media"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"

export type PendingAttachment = { id: string; name: string }

// Presigns against model_type=ticket / model_id=<class id> (the ticket id
// doesn't exist yet for the first message), PUTs the file, and returns the
// media row id the tickets API expects in media_ids.
async function uploadTicketFile(classId: string, file: File): Promise<PendingAttachment> {
  const res = await postMediaPresign({
    model_type: "ticket",
    model_id: classId,
    collection_name: "attachments",
    file_name: file.name,
    mime_type: file.type || "application/octet-stream",
    size: file.size,
  })
  const presign = res.status === 201 ? res.data.data : undefined
  if (!presign?.upload_url || !presign.media?.id) throw new Error(`presign failed (${res.status})`)
  const put = await fetch(presign.upload_url, { method: "PUT", body: file })
  if (!put.ok) throw new Error(`upload failed (${put.status})`)
  return { id: presign.media.id, name: file.name }
}

export function AttachmentPicker({
  classId,
  attachments,
  onChange,
  disabled,
}: {
  classId?: string
  attachments: PendingAttachment[]
  onChange: (next: PendingAttachment[]) => void
  disabled?: boolean
}) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  const pick = async (file: File | undefined) => {
    if (!file || !classId) return
    setUploading(true)
    try {
      const uploaded = await uploadTicketFile(classId, file)
      onChange([...attachments, uploaded])
    } catch {
      toast.error(t("tickets.error"))
    } finally {
      setUploading(false)
      if (inputRef.current) inputRef.current.value = ""
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        onChange={(e) => pick(e.target.files?.[0])}
        disabled={disabled || uploading || !classId}
      />
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={() => inputRef.current?.click()}
        disabled={disabled || uploading || !classId}
        aria-label={t("tickets.form.attach")}
      >
        {uploading ? <Spinner className="size-4" /> : <PaperclipIcon className="size-4" />}
        {t("tickets.form.attach")}
      </Button>
      {attachments.map((a) => (
        <span
          key={a.id}
          className="bg-muted text-muted-foreground inline-flex max-w-48 items-center gap-1 rounded-md px-2 py-1 text-xs"
        >
          <PaperclipIcon className="size-3 shrink-0" />
          <span className="truncate">{a.name}</span>
          <button
            type="button"
            className="hover:text-foreground"
            onClick={() => onChange(attachments.filter((x) => x.id !== a.id))}
            aria-label={t("tickets.form.remove")}
          >
            <XIcon className="size-3" />
          </button>
        </span>
      ))}
    </div>
  )
}

// AttachmentChips renders a message's stored media_ids; clicking resolves a
// presigned download URL on demand (URLs expire, so they're never persisted).
export function AttachmentChips({ mediaIds }: { mediaIds?: unknown }) {
  const { t } = useTranslation()
  const ids = Array.isArray(mediaIds) ? (mediaIds as string[]) : []
  if (ids.length === 0) return null

  const open = async (id: string) => {
    try {
      const res = await getMediaIdDownloadUrl(id)
      const url = res.status === 200 ? res.data.data?.url : undefined
      if (!url) throw new Error("no url")
      window.open(url, "_blank", "noopener")
    } catch {
      toast.error(t("tickets.error"))
    }
  }

  return (
    <div className="mt-1 flex flex-wrap gap-1">
      {ids.map((id, i) => (
        <button
          key={id}
          type="button"
          onClick={() => open(id)}
          className="bg-background/60 text-foreground/80 hover:bg-background inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-xs transition-colors"
        >
          <PaperclipIcon className="size-3" />
          {t("tickets.thread.attachment")} {i + 1}
        </button>
      ))}
    </div>
  )
}

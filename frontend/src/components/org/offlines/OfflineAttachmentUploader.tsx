import { useQueryClient } from "@tanstack/react-query"
import { FileIcon, Loader2Icon, Trash2Icon, UploadIcon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetMediaQueryKey, useDeleteMediaId, useGetMedia, usePostMediaPresign } from "@/api/media/media"
import { Button } from "@/components/ui/button"

import {
  OFFLINE_ATTACHMENTS_COLLECTION,
  OFFLINE_MODEL_TYPE,
  uploadOfflineAttachment,
} from "./attachments"

interface OfflineAttachmentUploaderProps {
  offlineId: string
}

export function OfflineAttachmentUploader({ offlineId }: OfflineAttachmentUploaderProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  const params = {
    model_type: OFFLINE_MODEL_TYPE,
    model_id: offlineId,
    collection: OFFLINE_ATTACHMENTS_COLLECTION,
  }
  const mediaQuery = useGetMedia(params)
  const attachments = (mediaQuery.data?.status === 200 && mediaQuery.data.data.data) || []

  const presign = usePostMediaPresign()
  const deleteMedia = useDeleteMediaId()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetMediaQueryKey(params) })

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ""
    if (!file) return
    setUploading(true)
    try {
      await uploadOfflineAttachment(presign.mutateAsync, offlineId, file)
      invalidate()
    } catch (err) {
      console.error(err)
      toast.error(t("org.session.offlines.form.uploadError"))
    } finally {
      setUploading(false)
    }
  }

  const handleRemove = (id: string) => {
    deleteMedia.mutate(
      { id },
      {
        onSuccess: invalidate,
        onError: () => toast.error(t("org.session.offlines.form.removeError")),
      }
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <input ref={inputRef} type="file" className="hidden" onChange={handleFile} />
      {mediaQuery.isPending ? (
        <div className="text-muted-foreground flex items-center gap-2 text-xs">
          <Loader2Icon className="size-4 animate-spin" />
        </div>
      ) : attachments.length === 0 ? (
        <p className="text-muted-foreground text-xs">{t("org.session.offlines.form.noAttachments")}</p>
      ) : (
        <ul className="flex flex-col gap-1.5">
          {attachments.map((m) => (
            <li
              key={m.id}
              className="border-border flex items-center gap-2 rounded-lg border px-3 py-2 text-sm"
            >
              <FileIcon className="text-muted-foreground size-4 shrink-0" />
              <span className="min-w-0 flex-1 truncate">{m.name || m.file_name}</span>
              <Button
                type="button"
                variant="ghost"
                size="icon-xs"
                className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                disabled={deleteMedia.isPending}
                onClick={() => m.id && handleRemove(m.id)}
              >
                {deleteMedia.isPending ? <Loader2Icon className="animate-spin" /> : <Trash2Icon />}
              </Button>
            </li>
          ))}
        </ul>
      )}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="self-start"
        disabled={uploading}
        onClick={() => inputRef.current?.click()}
      >
        {uploading ? <Loader2Icon className="animate-spin" /> : <UploadIcon />}
        {uploading ? t("org.session.offlines.form.uploading") : t("org.session.offlines.form.addAttachment")}
      </Button>
    </div>
  )
}

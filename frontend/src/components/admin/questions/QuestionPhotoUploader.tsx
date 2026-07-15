import { useQuery } from "@tanstack/react-query"
import { ImagePlusIcon, Loader2Icon, Trash2Icon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getMediaIdDownloadUrl, usePostMediaPresign } from "@/api/media/media"
import { Button } from "@/components/ui/button"

export interface PhotoMetadata {
  type: "photo"
  media_id: string
}

interface QuestionPhotoUploaderProps {
  value: PhotoMetadata[]
  onChange: (next: PhotoMetadata[]) => void
  questionId?: string
}

export function QuestionPhotoUploader({ value, onChange, questionId }: QuestionPhotoUploaderProps) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [isUploading, setIsUploading] = useState(false)

  const presign = usePostMediaPresign()

  const handlePick = () => inputRef.current?.click()

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ""
    if (!file) return
    if (!file.type.startsWith("image/")) {
      toast.error(t("admin.questions.form.photos.notImage"))
      return
    }

    setIsUploading(true)
    try {
      const modelID = questionId ?? crypto.randomUUID()
      const res = await presign.mutateAsync({
        data: {
          model_type: "question",
          model_id: modelID,
          collection_name: "photos",
          file_name: file.name,
          mime_type: file.type,
          size: file.size,
        },
      })
      if (res.status !== 201 || !res.data.data?.upload_url || !res.data.data?.media?.id) {
        throw new Error("presign failed")
      }
      const uploadURL = res.data.data.upload_url
      const mediaID = res.data.data.media.id

      const put = await fetch(uploadURL, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type },
      })
      if (!put.ok) throw new Error(`upload failed: ${put.status}`)

      onChange([...value, { type: "photo", media_id: mediaID }])
      toast.success(t("admin.questions.form.photos.uploadSuccess"))
    } catch (err) {
      console.error(err)
      toast.error(t("admin.questions.form.photos.uploadError"))
    } finally {
      setIsUploading(false)
    }
  }

  const handleRemove = (mediaID: string) => {
    onChange(value.filter((m) => m.media_id !== mediaID))
  }

  return (
    <div className="flex flex-col gap-2">
      <input ref={inputRef} type="file" accept="image/*" className="hidden" onChange={handleFile} />
      <div className="flex flex-wrap gap-2">
        {value.map((m) => (
          <PhotoThumb key={m.media_id} mediaID={m.media_id} onRemove={() => handleRemove(m.media_id)} />
        ))}
        <Button type="button" variant="outline" size="sm" onClick={handlePick} disabled={isUploading}>
          {isUploading ? <Loader2Icon className="animate-spin" /> : <ImagePlusIcon />}
          {t("admin.questions.form.photos.add")}
        </Button>
      </div>
      {value.length === 0 && <p className="text-muted-foreground text-xs">{t("admin.questions.form.photos.empty")}</p>}
    </div>
  )
}

interface PhotoThumbProps {
  mediaID: string
  onRemove: () => void
}

function PhotoThumb({ mediaID, onRemove }: PhotoThumbProps) {
  const { data } = useQuery({
    queryKey: ["media", "download-url", mediaID],
    queryFn: async () => {
      const res = await getMediaIdDownloadUrl(mediaID)
      return res.status === 200 ? (res.data.data?.url ?? null) : null
    },
    staleTime: 30 * 60 * 1000,
  })

  return (
    <div className="border-border bg-muted relative size-20 overflow-hidden rounded-md border">
      {data ? (
        <img src={data} alt="" className="size-full object-cover" />
      ) : (
        <div className="flex size-full items-center justify-center">
          <Loader2Icon className="size-4 animate-spin opacity-60" />
        </div>
      )}
      <Button
        type="button"
        variant="destructive"
        size="icon"
        className="absolute end-1 top-1 size-5"
        onClick={onRemove}
      >
        <Trash2Icon className="size-3" />
      </Button>
    </div>
  )
}

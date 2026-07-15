import { useQuery } from "@tanstack/react-query"
import { Loader2Icon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getMediaIdDownloadUrl, usePostMediaPresign } from "@/api/media/media"

export function OptionImageThumb({ mediaID }: { mediaID: string }) {
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
      <div className="bg-muted flex size-12 items-center justify-center rounded-md">
        <Loader2Icon className="size-4 animate-spin opacity-60" />
      </div>
    )
  }
  return <img src={url} alt="" className="size-12 rounded-md object-cover" />
}

interface OptionImageControlProps {
  value?: string | null
  questionId?: string
  onChange: (mediaID: string | null) => void
}

export function OptionImageControl({ value, questionId, onChange }: OptionImageControlProps) {
  const { t } = useTranslation()
  const presign = usePostMediaPresign()

  const handleFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ""
    if (!file) return
    if (!file.type.startsWith("image/")) {
      toast.error(t("admin.questions.form.photos.notImage"))
      return
    }

    try {
      const modelID = questionId ?? crypto.randomUUID()
      const res = await presign.mutateAsync({
        data: {
          model_type: "question",
          model_id: modelID,
          collection_name: "option-photos",
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
      onChange(mediaID)
    } catch (err) {
      console.error(err)
      toast.error(t("admin.questions.form.photos.uploadError"))
    }
  }

  return (
    <div className="flex items-center gap-2">
      {value ? <OptionImageThumb mediaID={value} /> : null}
      <label className="text-muted-foreground hover:text-foreground cursor-pointer text-xs">
        <input type="file" accept="image/*" className="hidden" onChange={handleFile} disabled={presign.isPending} />
        {value ? t("admin.questions.form.optionImage.replace") : t("admin.questions.form.optionImage.add")}
      </label>
      {value ? (
        <button type="button" className="text-destructive text-xs" onClick={() => onChange(null)}>
          {t("admin.questions.form.optionImage.remove")}
        </button>
      ) : null}
    </div>
  )
}

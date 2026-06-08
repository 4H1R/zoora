import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { FileIcon, Loader2Icon, Trash2Icon, UploadIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { usePostMediaPresign } from "@/api/media/media"
import { getGetOfflinesQueryKey, usePostOfflines, usePutOfflinesId } from "@/api/offlines/offlines"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

import { uploadOfflineAttachment } from "./attachments"
import { OfflineAttachmentUploader } from "./OfflineAttachmentUploader"

const PREFIX = "org.session.offlines.form"

const schema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  published_at: z.string().optional(),
})

type Values = z.infer<typeof schema>

const defaults: Values = { title: "", description: "", published_at: "" }

function isoToLocalInput(iso?: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function localInputToISO(value?: string): string | undefined {
  if (!value) return undefined
  const d = new Date(value)
  if (isNaN(d.getTime())) return undefined
  return d.toISOString()
}

interface OfflineFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  room?: OfflineRoom | null
  classSessionId: string
}

export function OfflineFormDialog({ open, onOpenChange, room, classSessionId }: OfflineFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!room
  const inputRef = useRef<HTMLInputElement>(null)

  const [pendingFiles, setPendingFiles] = useState<File[]>([])
  const [uploading, setUploading] = useState(false)

  const form = useForm<Values>({ resolver: zodResolver(schema), defaultValues: defaults })

  useEffect(() => {
    if (!open) return
    setPendingFiles([])
    form.reset(
      room
        ? { title: room.title ?? "", description: room.description ?? "", published_at: isoToLocalInput(room.published_at) }
        : defaults
    )
  }, [open, room])

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetOfflinesQueryKey() })

  const createMutation = usePostOfflines()
  const updateMutation = usePutOfflinesId()
  const presign = usePostMediaPresign()

  const isLoading = createMutation.isPending || updateMutation.isPending || uploading

  const onSubmit = form.handleSubmit(async (values) => {
    const published_at = localInputToISO(values.published_at)
    if (isEdit && room?.id) {
      await updateMutation.mutateAsync({
        id: room.id,
        data: { title: values.title, description: values.description, published_at },
      })
      toast.success(t(`${PREFIX}.updateSuccess`))
      invalidate()
      onOpenChange(false)
      return
    }
    const res = await createMutation.mutateAsync({
      data: { class_session_id: classSessionId, title: values.title, description: values.description, published_at },
    })
    const newId = res.status === 201 ? res.data.data?.id : undefined
    if (newId && pendingFiles.length > 0) {
      setUploading(true)
      for (const file of pendingFiles) {
        try {
          await uploadOfflineAttachment(presign.mutateAsync, newId, file)
        } catch (err) {
          console.error(err)
          toast.error(t(`${PREFIX}.uploadError`))
        }
      }
      setUploading(false)
    }
    toast.success(t(`${PREFIX}.createSuccess`))
    invalidate()
    onOpenChange(false)
  })

  const errors = form.formState.errors

  const handlePick = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ""
    if (file) setPendingFiles((prev) => [...prev, file])
  }

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t(`${PREFIX}.editTitle`) : t(`${PREFIX}.createTitle`)}
      description={isEdit ? t(`${PREFIX}.editDescription`) : t(`${PREFIX}.createDescription`)}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.title || undefined}>
          <FieldLabel>{t(`${PREFIX}.title`)}</FieldLabel>
          <Input {...form.register("title")} placeholder={t(`${PREFIX}.titlePlaceholder`)} />
          <FieldError errors={[errors.title]} />
        </Field>
        <Field>
          <FieldLabel>{t(`${PREFIX}.description`)}</FieldLabel>
          <Textarea {...form.register("description")} placeholder={t(`${PREFIX}.descriptionPlaceholder`)} rows={3} />
        </Field>
        <Field>
          <FieldLabel>{t(`${PREFIX}.publishedAt`)}</FieldLabel>
          <Input type="datetime-local" {...form.register("published_at")} />
        </Field>
        <Field>
          <FieldLabel>{t(`${PREFIX}.attachments`)}</FieldLabel>
          {isEdit && room?.id ? (
            <OfflineAttachmentUploader offlineId={room.id} />
          ) : (
            <div className="flex flex-col gap-2">
              {pendingFiles.length === 0 ? (
                <p className="text-muted-foreground text-xs">{t(`${PREFIX}.noAttachments`)}</p>
              ) : (
                <ul className="flex flex-col gap-1.5">
                  {pendingFiles.map((f, i) => (
                    <li
                      key={`${f.name}-${i}`}
                      className="border-border flex items-center gap-2 rounded-lg border px-3 py-2 text-sm"
                    >
                      <FileIcon className="text-muted-foreground size-4 shrink-0" />
                      <span className="min-w-0 flex-1 truncate">{f.name}</span>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        onClick={() => setPendingFiles((prev) => prev.filter((_, j) => j !== i))}
                      >
                        <Trash2Icon />
                      </Button>
                    </li>
                  ))}
                </ul>
              )}
              <input ref={inputRef} type="file" className="hidden" onChange={handlePick} />
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="self-start"
                disabled={uploading}
                onClick={() => inputRef.current?.click()}
              >
                {uploading ? <Loader2Icon className="animate-spin" /> : <UploadIcon />}
                {t(`${PREFIX}.addAttachment`)}
              </Button>
            </div>
          )}
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

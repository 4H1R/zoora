import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetPracticesQueryKey,
  usePostPractices,
  usePutPracticesId,
} from "@/api/practices/practices"
import {
  MediaAttachmentUploader,
  type PendingAttachment,
} from "@/components/media/MediaAttachmentUploader"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const TRANSLATION_PREFIX = "org.session.practices.form"

const schema = z
  .object({
    title: z.string().min(2),
    content: z.string().optional(),
    max_score: z.coerce.number().min(0).optional(),
    start_time: z.string().min(1),
    end_time: z.string().min(1),
  })
  .refine((v) => new Date(v.end_time) > new Date(v.start_time), {
    path: ["end_time"],
    message: "end_after_start",
  })

type Values = z.infer<typeof schema>

const defaults: Values = {
  title: "",
  content: "",
  max_score: 0,
  start_time: "",
  end_time: "",
}

function isoToLocalInput(iso?: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const practiceToValues = (practice: PracticeRoom): Values => ({
  title: practice.title ?? "",
  content: practice.content ?? "",
  max_score: practice.max_score ?? 0,
  start_time: isoToLocalInput(practice.start_time),
  end_time: isoToLocalInput(practice.end_time),
})

interface PracticeFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  practice?: PracticeRoom | null
  classSessionId: string
}

export function PracticeFormDialog({
  open,
  onOpenChange,
  practice,
  classSessionId,
}: PracticeFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!practice

  const [attachments, setAttachments] = useState<PendingAttachment[]>([])
  const [newModelId, setNewModelId] = useState(() => crypto.randomUUID())

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: defaults,
  })

  useEffect(() => {
    if (!open) return
    form.reset(practice ? practiceToValues(practice) : defaults)
    setAttachments(
      (practice?.attachments ?? []).map((id) => ({ media_id: id, name: id, size: 0 })),
    )
    if (!practice) setNewModelId(crypto.randomUUID())
  }, [open, practice])

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: getGetPracticesQueryKey() })

  const createMutation = usePostPractices({
    mutation: {
      onSuccess: () => {
        toast.success(t(`${TRANSLATION_PREFIX}.createSuccess`))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutPracticesId({
    mutation: {
      onSuccess: () => {
        toast.success(t(`${TRANSLATION_PREFIX}.updateSuccess`))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = form.handleSubmit((values) => {
    const payload = {
      title: values.title,
      content: values.content,
      max_score: values.max_score,
      start_time: new Date(values.start_time).toISOString(),
      end_time: new Date(values.end_time).toISOString(),
      attachments: attachments.map((a) => a.media_id),
    }
    if (isEdit && practice?.id) {
      updateMutation.mutate({ id: practice.id, data: payload })
    } else {
      createMutation.mutate({ data: { ...payload, class_session_id: classSessionId } })
    }
  })

  const errors = form.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t(`${TRANSLATION_PREFIX}.editTitle`) : t(`${TRANSLATION_PREFIX}.createTitle`)}
      description={
        isEdit ? t(`${TRANSLATION_PREFIX}.editDescription`) : t(`${TRANSLATION_PREFIX}.createDescription`)
      }
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.title || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.title`)}</FieldLabel>
          <Input {...form.register("title")} placeholder={t(`${TRANSLATION_PREFIX}.titlePlaceholder`)} />
          <FieldError errors={[errors.title]} />
        </Field>
        <Field>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.content`)}</FieldLabel>
          <Textarea
            {...form.register("content")}
            placeholder={t(`${TRANSLATION_PREFIX}.contentPlaceholder`)}
            rows={4}
          />
        </Field>
        <Field>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.attachments`)}</FieldLabel>
          <MediaAttachmentUploader
            value={attachments}
            onChange={setAttachments}
            modelType="practice"
            modelId={isEdit ? practice?.id : newModelId}
          />
        </Field>
        <Field data-invalid={!!errors.max_score || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.maxScore`)}</FieldLabel>
          <Input
            type="number"
            min={0}
            step="any"
            {...form.register("max_score")}
            placeholder={t(`${TRANSLATION_PREFIX}.maxScorePlaceholder`)}
          />
          <FieldError errors={[errors.max_score]} />
        </Field>
        <Field data-invalid={!!errors.start_time || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.startTime`)}</FieldLabel>
          <Input type="datetime-local" {...form.register("start_time")} />
          <FieldError errors={[errors.start_time]} />
        </Field>
        <Field data-invalid={!!errors.end_time || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.endTime`)}</FieldLabel>
          <Input type="datetime-local" {...form.register("end_time")} />
          <FieldError errors={[errors.end_time]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

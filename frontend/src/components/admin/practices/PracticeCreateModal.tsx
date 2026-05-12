import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminPracticesQueryKey } from "@/api/admin-practices/admin-practices"
import { usePostPractices, usePutPracticesId } from "@/api/practices/practices"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const createSchema = z
  .object({
    class_id: z.string().uuid(),
    class_session_id: z.string().uuid(),
    title: z.string().min(2),
    content: z.string().optional(),
    max_score: z.coerce.number().min(0).optional(),
    start_time: z.string().min(1),
    end_time: z.string().min(1),
  })
  .refine((v) => new Date(v.end_time) > new Date(v.start_time), {
    path: ["end_time"],
    message: "end_time must be after start_time",
  })

const editSchema = z
  .object({
    title: z.string().min(2),
    content: z.string().optional(),
    max_score: z.coerce.number().min(0).optional(),
    start_time: z.string().min(1),
    end_time: z.string().min(1),
  })
  .refine((v) => new Date(v.end_time) > new Date(v.start_time), {
    path: ["end_time"],
    message: "end_time must be after start_time",
  })

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

interface PracticeCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  practice?: PracticeRoom | null
  defaultClassId?: string
  defaultSessionId?: string
}

function isoToLocalInput(iso?: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function PracticeCreateModal({
  open,
  onOpenChange,
  practice,
  defaultClassId,
  defaultSessionId,
}: PracticeCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!practice

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      class_id: "",
      class_session_id: "",
      title: "",
      content: "",
      max_score: 0,
      start_time: "",
      end_time: "",
    },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: {
      title: "",
      content: "",
      max_score: 0,
      start_time: "",
      end_time: "",
    },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && practice) {
      editForm.reset({
        title: practice.title ?? "",
        content: practice.content ?? "",
        max_score: practice.max_score ?? 0,
        start_time: isoToLocalInput(practice.start_time),
        end_time: isoToLocalInput(practice.end_time),
      })
    } else {
      createForm.reset({
        class_id: defaultClassId ?? "",
        class_session_id: defaultSessionId ?? "",
        title: "",
        content: "",
        max_score: 0,
        start_time: "",
        end_time: "",
      })
    }
  }, [open, practice, isEdit, defaultClassId, defaultSessionId])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminPracticesQueryKey() })
  }

  const createMutation = usePostPractices({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.practices.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutPracticesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.practices.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const selectedClassId = createForm.watch("class_id")
  const selectedSessionId = createForm.watch("class_session_id")

  const onSubmitCreate = createForm.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: values.class_session_id,
        title: values.title,
        content: values.content,
        max_score: values.max_score,
        start_time: new Date(values.start_time).toISOString(),
        end_time: new Date(values.end_time).toISOString(),
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!practice?.id) return
    updateMutation.mutate({
      id: practice.id,
      data: {
        title: values.title,
        content: values.content,
        max_score: values.max_score,
        start_time: new Date(values.start_time).toISOString(),
        end_time: new Date(values.end_time).toISOString(),
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.practices.form.editTitle") : t("admin.practices.form.createTitle")}
      description={
        isEdit
          ? t("admin.practices.form.editDescription")
          : t("admin.practices.form.createDescription")
      }
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {isEdit ? (
          <>
            <Field data-invalid={!!editErrors.title || undefined}>
              <FieldLabel>{t("admin.practices.form.title")}</FieldLabel>
              <Input
                {...editForm.register("title")}
                placeholder={t("admin.practices.form.titlePlaceholder")}
              />
              <FieldError errors={[editErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.practices.form.content")}</FieldLabel>
              <Textarea
                {...editForm.register("content")}
                placeholder={t("admin.practices.form.contentPlaceholder")}
                rows={4}
              />
            </Field>
            <Field data-invalid={!!editErrors.max_score || undefined}>
              <FieldLabel>{t("admin.practices.form.maxScore")}</FieldLabel>
              <Input
                type="number"
                min={0}
                step="any"
                {...editForm.register("max_score")}
                placeholder={t("admin.practices.form.maxScorePlaceholder")}
              />
              <FieldError errors={[editErrors.max_score]} />
            </Field>
            <Field data-invalid={!!editErrors.start_time || undefined}>
              <FieldLabel>{t("admin.practices.form.startTime")}</FieldLabel>
              <Input type="datetime-local" {...editForm.register("start_time")} />
              <FieldError errors={[editErrors.start_time]} />
            </Field>
            <Field data-invalid={!!editErrors.end_time || undefined}>
              <FieldLabel>{t("admin.practices.form.endTime")}</FieldLabel>
              <Input type="datetime-local" {...editForm.register("end_time")} />
              <FieldError errors={[editErrors.end_time]} />
            </Field>
          </>
        ) : (
          <>
            <Field data-invalid={!!createErrors.class_id || undefined}>
              <FieldLabel>{t("admin.practices.form.class")}</FieldLabel>
              <ClassPicker
                value={selectedClassId || undefined}
                onChange={(id) => {
                  createForm.setValue("class_id", id, { shouldValidate: true })
                  createForm.setValue("class_session_id", "", { shouldValidate: true })
                }}
                placeholder={t("admin.practices.form.classPlaceholder")}
              />
              <FieldError errors={[createErrors.class_id]} />
            </Field>
            <Field data-invalid={!!createErrors.class_session_id || undefined}>
              <FieldLabel>{t("admin.practices.form.session")}</FieldLabel>
              <SessionPicker
                classId={selectedClassId || undefined}
                value={selectedSessionId || undefined}
                onChange={(id) =>
                  createForm.setValue("class_session_id", id, { shouldValidate: true })
                }
                placeholder={t("admin.practices.form.sessionPlaceholder")}
              />
              <FieldError errors={[createErrors.class_session_id]} />
            </Field>
            <Field data-invalid={!!createErrors.title || undefined}>
              <FieldLabel>{t("admin.practices.form.title")}</FieldLabel>
              <Input
                {...createForm.register("title")}
                placeholder={t("admin.practices.form.titlePlaceholder")}
              />
              <FieldError errors={[createErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.practices.form.content")}</FieldLabel>
              <Textarea
                {...createForm.register("content")}
                placeholder={t("admin.practices.form.contentPlaceholder")}
                rows={4}
              />
            </Field>
            <Field data-invalid={!!createErrors.max_score || undefined}>
              <FieldLabel>{t("admin.practices.form.maxScore")}</FieldLabel>
              <Input
                type="number"
                min={0}
                step="any"
                {...createForm.register("max_score")}
                placeholder={t("admin.practices.form.maxScorePlaceholder")}
              />
              <FieldError errors={[createErrors.max_score]} />
            </Field>
            <Field data-invalid={!!createErrors.start_time || undefined}>
              <FieldLabel>{t("admin.practices.form.startTime")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("start_time")} />
              <FieldError errors={[createErrors.start_time]} />
            </Field>
            <Field data-invalid={!!createErrors.end_time || undefined}>
              <FieldLabel>{t("admin.practices.form.endTime")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("end_time")} />
              <FieldError errors={[createErrors.end_time]} />
            </Field>
          </>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

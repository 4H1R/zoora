import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminOfflinesQueryKey } from "@/api/admin-offlines/admin-offlines"
import { usePostOfflines, usePutOfflinesId } from "@/api/offlines/offlines"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const createSchema = z.object({
  class_id: z.string().uuid(),
  class_session_id: z.string().uuid(),
  title: z.string().min(2),
  description: z.string().optional(),
  published_at: z.string().optional(),
})

const editSchema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  published_at: z.string().optional(),
})

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

interface OfflineCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  room?: OfflineRoom | null
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

function localInputToISO(value?: string): string | undefined {
  if (!value) return undefined
  const d = new Date(value)
  if (isNaN(d.getTime())) return undefined
  return d.toISOString()
}

export function OfflineCreateModal({
  open,
  onOpenChange,
  room,
  defaultClassId,
  defaultSessionId,
}: OfflineCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!room

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: { class_id: "", class_session_id: "", title: "", description: "", published_at: "" },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { title: "", description: "", published_at: "" },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && room) {
      editForm.reset({
        title: room.title ?? "",
        description: room.description ?? "",
        published_at: isoToLocalInput(room.published_at),
      })
    } else {
      createForm.reset({
        class_id: defaultClassId ?? "",
        class_session_id: defaultSessionId ?? "",
        title: "",
        description: "",
        published_at: "",
      })
    }
  }, [open, room, isEdit, defaultClassId, defaultSessionId])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminOfflinesQueryKey() })
  }

  const createMutation = usePostOfflines({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.offlines.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutOfflinesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.offlines.form.updateSuccess"))
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
        description: values.description,
        published_at: localInputToISO(values.published_at),
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!room?.id) return
    updateMutation.mutate({
      id: room.id,
      data: {
        title: values.title,
        description: values.description,
        published_at: localInputToISO(values.published_at),
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.offlines.form.editTitle") : t("admin.offlines.form.createTitle")}
      description={
        isEdit ? t("admin.offlines.form.editDescription") : t("admin.offlines.form.createDescription")
      }
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {isEdit ? (
          <>
            <Field data-invalid={!!editErrors.title || undefined}>
              <FieldLabel>{t("admin.offlines.form.title")}</FieldLabel>
              <Input
                {...editForm.register("title")}
                placeholder={t("admin.offlines.form.titlePlaceholder")}
              />
              <FieldError errors={[editErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.description")}</FieldLabel>
              <Textarea
                {...editForm.register("description")}
                placeholder={t("admin.offlines.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.publishedAt")}</FieldLabel>
              <Input type="datetime-local" {...editForm.register("published_at")} />
            </Field>
          </>
        ) : (
          <>
            <Field data-invalid={!!createErrors.class_id || undefined}>
              <FieldLabel>{t("admin.offlines.form.class")}</FieldLabel>
              <ClassPicker
                value={selectedClassId || undefined}
                onChange={(id) => {
                  createForm.setValue("class_id", id, { shouldValidate: true })
                  createForm.setValue("class_session_id", "", { shouldValidate: true })
                }}
                placeholder={t("admin.offlines.form.classPlaceholder")}
              />
              <FieldError errors={[createErrors.class_id]} />
            </Field>
            <Field data-invalid={!!createErrors.class_session_id || undefined}>
              <FieldLabel>{t("admin.offlines.form.session")}</FieldLabel>
              <SessionPicker
                classId={selectedClassId || undefined}
                value={selectedSessionId || undefined}
                onChange={(id) => createForm.setValue("class_session_id", id, { shouldValidate: true })}
                placeholder={t("admin.offlines.form.sessionPlaceholder")}
              />
              <FieldError errors={[createErrors.class_session_id]} />
            </Field>
            <Field data-invalid={!!createErrors.title || undefined}>
              <FieldLabel>{t("admin.offlines.form.title")}</FieldLabel>
              <Input
                {...createForm.register("title")}
                placeholder={t("admin.offlines.form.titlePlaceholder")}
              />
              <FieldError errors={[createErrors.title]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.description")}</FieldLabel>
              <Textarea
                {...createForm.register("description")}
                placeholder={t("admin.offlines.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field>
              <FieldLabel>{t("admin.offlines.form.publishedAt")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("published_at")} />
            </Field>
          </>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

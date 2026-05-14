import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminQuizzesQueryKey } from "@/api/admin-quizzes/admin-quizzes"
import {
  getGetQuizzesQueryKey,
  postQuizzesIdRooms,
  usePostQuizzes,
  usePutQuizzesId,
} from "@/api/quizzes/quizzes"
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
  duration_minutes: z.coerce.number().int().gt(0),
})

const editSchema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().int().gt(0),
})

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

interface QuizCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  quiz?: Quiz | null
  defaultClassId?: string
}

export function QuizCreateModal({
  open,
  onOpenChange,
  quiz,
  defaultClassId,
}: QuizCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!quiz

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      class_id: defaultClassId ?? "",
      class_session_id: "",
      title: "",
      description: "",
      duration_minutes: 30,
    },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { title: "", description: "", duration_minutes: 30 },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && quiz) {
      editForm.reset({
        title: quiz.title ?? "",
        description: quiz.description ?? "",
        duration_minutes: quiz.duration_minutes ?? 30,
      })
    } else {
      createForm.reset({
        class_id: defaultClassId ?? "",
        class_session_id: "",
        title: "",
        description: "",
        duration_minutes: 30,
      })
    }
  }, [open, quiz, isEdit, defaultClassId])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetAdminQuizzesQueryKey() })
  }

  const createMutation = usePostQuizzes({
    mutation: {
      onSuccess: async (res) => {
        const sessionId = createForm.getValues("class_session_id")
        const quizId = res.status === 201 ? res.data.data?.id : undefined
        if (sessionId && quizId) {
          try {
            await postQuizzesIdRooms(quizId, { class_session_id: sessionId })
          } catch {
            toast.error(t("admin.quizzes.form.linkRoomFailed"))
          }
        }
        toast.success(t("admin.quizzes.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutQuizzesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.quizzes.form.updateSuccess"))
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
        class_id: values.class_id,
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!quiz?.id) return
    updateMutation.mutate({
      id: quiz.id,
      data: {
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.quizzes.form.editTitle") : t("admin.quizzes.form.createTitle")}
      description={
        isEdit
          ? t("admin.quizzes.form.editDescription")
          : t("admin.quizzes.form.createDescription")
      }
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {!isEdit && (
          <>
            <Field data-invalid={!!createErrors.class_id || undefined}>
              <FieldLabel>{t("admin.quizzes.form.class")}</FieldLabel>
              <ClassPicker
                value={selectedClassId || undefined}
                onChange={(id) => {
                  createForm.setValue("class_id", id, { shouldValidate: true })
                  createForm.setValue("class_session_id", "", { shouldValidate: true })
                }}
                placeholder={t("admin.quizzes.form.classPlaceholder")}
              />
              <FieldError errors={[createErrors.class_id]} />
            </Field>
            <Field data-invalid={!!createErrors.class_session_id || undefined}>
              <FieldLabel>{t("admin.quizzes.form.session")}</FieldLabel>
              <SessionPicker
                classId={selectedClassId || undefined}
                value={selectedSessionId || undefined}
                onChange={(id) =>
                  createForm.setValue("class_session_id", id, { shouldValidate: true })
                }
                placeholder={t("admin.quizzes.form.sessionPlaceholder")}
              />
              <FieldError errors={[createErrors.class_session_id]} />
            </Field>
          </>
        )}

        <Field
          data-invalid={!!(isEdit ? editErrors.title : createErrors.title) || undefined}
        >
          <FieldLabel>{t("admin.quizzes.form.title")}</FieldLabel>
          <Input
            {...(isEdit ? editForm.register("title") : createForm.register("title"))}
            placeholder={t("admin.quizzes.form.titlePlaceholder")}
          />
          <FieldError errors={[isEdit ? editErrors.title : createErrors.title]} />
        </Field>

        <Field>
          <FieldLabel>{t("admin.quizzes.form.description")}</FieldLabel>
          <Textarea
            {...(isEdit
              ? editForm.register("description")
              : createForm.register("description"))}
            placeholder={t("admin.quizzes.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>

        <Field
          data-invalid={
            !!(isEdit ? editErrors.duration_minutes : createErrors.duration_minutes) || undefined
          }
        >
          <FieldLabel>{t("admin.quizzes.form.duration")}</FieldLabel>
          <Input
            type="number"
            min={1}
            {...(isEdit
              ? editForm.register("duration_minutes")
              : createForm.register("duration_minutes"))}
          />
          <FieldError
            errors={[isEdit ? editErrors.duration_minutes : createErrors.duration_minutes]}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

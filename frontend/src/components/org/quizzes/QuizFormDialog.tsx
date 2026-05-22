import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetQuizzesQueryKey,
  postQuizzesIdRooms,
  usePostQuizzes,
  usePutQuizzesId,
} from "@/api/quizzes/quizzes"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const schema = z
  .object({
    title: z.string().min(2),
    description: z.string().optional(),
    duration_minutes: z.coerce.number().int().gt(0),
    started_at: z.string().optional(),
    ended_at: z.string().optional(),
  })
  .refine(
    (v) => {
      if (!v.started_at || !v.ended_at) return true
      return new Date(v.ended_at).getTime() > new Date(v.started_at).getTime()
    },
    { path: ["ended_at"], message: "end_after_start" }
  )

type Values = z.infer<typeof schema>

const emptyDefaults: Values = {
  title: "",
  description: "",
  duration_minutes: 30,
  started_at: "",
  ended_at: "",
}

interface QuizFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  quiz?: Quiz | null
  classId: string
  classSessionId: string
}

export function QuizFormDialog({
  open,
  onOpenChange,
  quiz,
  classId,
  classSessionId,
}: QuizFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!quiz

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: emptyDefaults,
  })

  useEffect(() => {
    if (!open) return
    form.reset(
      isEdit && quiz
        ? {
            ...emptyDefaults,
            title: quiz.title ?? "",
            description: quiz.description ?? "",
            duration_minutes: quiz.duration_minutes ?? 30,
          }
        : emptyDefaults
    )
  }, [open, quiz, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
  }

  const createMutation = usePostQuizzes({
    mutation: {
      onSuccess: async (res) => {
        const quizId = res.status === 201 ? res.data.data?.id : undefined
        const started = form.getValues("started_at")
        const ended = form.getValues("ended_at")
        if (quizId) {
          try {
            await postQuizzesIdRooms(quizId, {
              class_session_id: classSessionId,
              started_at: started ? new Date(started).toISOString() : undefined,
              ended_at: ended ? new Date(ended).toISOString() : undefined,
            })
          } catch {
            toast.error(t("org.session.quizzes.form.linkRoomFailed"))
          }
        }
        toast.success(t("org.session.quizzes.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutQuizzesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.quizzes.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = form.handleSubmit((values) => {
    if (isEdit && quiz?.id) {
      updateMutation.mutate({
        id: quiz.id,
        data: {
          title: values.title,
          description: values.description,
          duration_minutes: values.duration_minutes,
        },
      })
    } else {
      createMutation.mutate({
        data: {
          class_id: classId,
          title: values.title,
          description: values.description,
          duration_minutes: values.duration_minutes,
        },
      })
    }
  })

  const errors = form.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        isEdit
          ? t("org.session.quizzes.form.editTitle")
          : t("org.session.quizzes.form.createTitle")
      }
      description={
        isEdit
          ? t("org.session.quizzes.form.editDescription")
          : t("org.session.quizzes.form.createDescription")
      }
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.title || undefined}>
          <FieldLabel>{t("org.session.quizzes.form.title")}</FieldLabel>
          <Input
            {...form.register("title")}
            placeholder={t("org.session.quizzes.form.titlePlaceholder")}
          />
          <FieldError errors={[errors.title]} />
        </Field>
        <Field>
          <FieldLabel>{t("org.session.quizzes.form.description")}</FieldLabel>
          <Textarea
            {...form.register("description")}
            placeholder={t("org.session.quizzes.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>
        <Field data-invalid={!!errors.duration_minutes || undefined}>
          <FieldLabel>{t("org.session.quizzes.form.duration")}</FieldLabel>
          <Input type="number" min={1} {...form.register("duration_minutes")} />
          <FieldError errors={[errors.duration_minutes]} />
        </Field>
        {!isEdit ? (
          <>
            <Field>
              <FieldLabel>{t("org.session.quizzes.form.startedAt")}</FieldLabel>
              <Input type="datetime-local" {...form.register("started_at")} />
              <p className="text-muted-foreground text-xs">
                {t("org.session.quizzes.form.startedAtHint")}
              </p>
            </Field>
            <Field data-invalid={!!errors.ended_at || undefined}>
              <FieldLabel>{t("org.session.quizzes.form.endedAt")}</FieldLabel>
              <Input type="datetime-local" {...form.register("ended_at")} />
              {errors.ended_at?.message === "end_after_start" ? (
                <p className="text-destructive text-xs">
                  {t("org.session.quizzes.form.endedAtError")}
                </p>
              ) : (
                <p className="text-muted-foreground text-xs">
                  {t("org.session.quizzes.form.endedAtHint")}
                </p>
              )}
            </Field>
          </>
        ) : null}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

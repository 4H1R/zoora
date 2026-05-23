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
import { Checkbox } from "@/components/ui/checkbox"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const baseSchema = {
  title: z.string().min(2),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().int().gt(0),
  no_back_navigation: z.boolean(),
  shuffle_questions: z.boolean(),
}

const createSchema = z
  .object({
    ...baseSchema,
    started_at: z.string().min(1),
    ended_at: z.string().min(1),
  })
  .refine(
    (v) => new Date(v.ended_at).getTime() > new Date(v.started_at).getTime(),
    { path: ["ended_at"], message: "end_after_start" }
  )

const editSchema = z.object(baseSchema)

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

const emptyCreate: CreateValues = {
  title: "",
  description: "",
  duration_minutes: 30,
  no_back_navigation: false,
  shuffle_questions: false,
  started_at: "",
  ended_at: "",
}

const emptyEdit: EditValues = {
  title: "",
  description: "",
  duration_minutes: 30,
  no_back_navigation: false,
  shuffle_questions: false,
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

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: emptyCreate,
  })
  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: emptyEdit,
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && quiz) {
      editForm.reset({
        title: quiz.title ?? "",
        description: quiz.description ?? "",
        duration_minutes: quiz.duration_minutes ?? 30,
        no_back_navigation: quiz.no_back_navigation ?? false,
        shuffle_questions: quiz.shuffle_questions ?? false,
      })
    } else {
      createForm.reset(emptyCreate)
    }
  }, [open, quiz, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
  }

  const createMutation = usePostQuizzes({
    mutation: {
      onSuccess: async (res) => {
        const quizId = res.status === 201 ? res.data.data?.id : undefined
        const started = createForm.getValues("started_at")
        const ended = createForm.getValues("ended_at")
        if (quizId) {
          try {
            await postQuizzesIdRooms(quizId, {
              class_session_id: classSessionId,
              started_at: new Date(started).toISOString(),
              ended_at: new Date(ended).toISOString(),
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

  const onSubmit = isEdit
    ? editForm.handleSubmit((values) => {
        if (!quiz?.id) return
        updateMutation.mutate({
          id: quiz.id,
          data: {
            title: values.title,
            description: values.description,
            duration_minutes: values.duration_minutes,
            no_back_navigation: values.no_back_navigation,
            shuffle_questions: values.shuffle_questions,
          },
        })
      })
    : createForm.handleSubmit((values) => {
        createMutation.mutate({
          data: {
            class_id: classId,
            title: values.title,
            description: values.description,
            duration_minutes: values.duration_minutes,
            no_back_navigation: values.no_back_navigation,
            shuffle_questions: values.shuffle_questions,
          },
        })
      })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors
  const createNoBack = createForm.watch("no_back_navigation")
  const createShuffle = createForm.watch("shuffle_questions")
  const editNoBack = editForm.watch("no_back_navigation")
  const editShuffle = editForm.watch("shuffle_questions")

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
        <Field data-invalid={!!(isEdit ? editErrors.title : createErrors.title) || undefined}>
          <FieldLabel>{t("org.session.quizzes.form.title")}</FieldLabel>
          <Input
            {...(isEdit ? editForm.register("title") : createForm.register("title"))}
            placeholder={t("org.session.quizzes.form.titlePlaceholder")}
          />
          <FieldError errors={[isEdit ? editErrors.title : createErrors.title]} />
        </Field>
        <Field>
          <FieldLabel>{t("org.session.quizzes.form.description")}</FieldLabel>
          <Textarea
            {...(isEdit
              ? editForm.register("description")
              : createForm.register("description"))}
            placeholder={t("org.session.quizzes.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>
        <Field
          data-invalid={
            !!(isEdit ? editErrors.duration_minutes : createErrors.duration_minutes) || undefined
          }
        >
          <FieldLabel>{t("org.session.quizzes.form.duration")}</FieldLabel>
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

        {!isEdit ? (
          <>
            <Field data-invalid={!!createErrors.started_at || undefined}>
              <FieldLabel>{t("org.session.quizzes.form.startedAt")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("started_at")} />
              <p className="text-muted-foreground text-xs">
                {t("org.session.quizzes.form.startedAtHint")}
              </p>
              <FieldError errors={[createErrors.started_at]} />
            </Field>
            <Field data-invalid={!!createErrors.ended_at || undefined}>
              <FieldLabel>{t("org.session.quizzes.form.endedAt")}</FieldLabel>
              <Input type="datetime-local" {...createForm.register("ended_at")} />
              {createErrors.ended_at?.message === "end_after_start" ? (
                <p className="text-destructive text-xs">
                  {t("org.session.quizzes.form.endedAtError")}
                </p>
              ) : (
                <p className="text-muted-foreground text-xs">
                  {t("org.session.quizzes.form.endedAtHint")}
                </p>
              )}
              <FieldError errors={[createErrors.ended_at]} />
            </Field>
          </>
        ) : null}

        <div className="border-foreground/10 flex flex-col gap-3 rounded-md border border-dashed p-3">
          <Field className="flex-row items-start gap-3 space-y-0">
            <Checkbox
              checked={isEdit ? editNoBack : createNoBack}
              onCheckedChange={(c) => {
                if (isEdit) editForm.setValue("no_back_navigation", !!c)
                else createForm.setValue("no_back_navigation", !!c)
              }}
              className="mt-0.5"
            />
            <div className="flex flex-col gap-0.5">
              <FieldLabel className="cursor-pointer">
                {t("org.session.quizzes.form.noBackNavigation")}
              </FieldLabel>
              <p className="text-muted-foreground text-xs leading-relaxed">
                {t("org.session.quizzes.form.noBackNavigationHint")}
              </p>
            </div>
          </Field>
          <Field className="flex-row items-start gap-3 space-y-0">
            <Checkbox
              checked={isEdit ? editShuffle : createShuffle}
              onCheckedChange={(c) => {
                if (isEdit) editForm.setValue("shuffle_questions", !!c)
                else createForm.setValue("shuffle_questions", !!c)
              }}
              className="mt-0.5"
            />
            <div className="flex flex-col gap-0.5">
              <FieldLabel className="cursor-pointer">
                {t("org.session.quizzes.form.shuffleQuestions")}
              </FieldLabel>
              <p className="text-muted-foreground text-xs leading-relaxed">
                {t("org.session.quizzes.form.shuffleQuestionsHint")}
              </p>
            </div>
          </Field>
        </div>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

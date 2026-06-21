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
import {
  QuizCoreFields,
  QuizFlagsFields,
  QuizScheduleFields,
} from "@/components/quizzes/quiz-form-fields"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"

const TRANSLATION_PREFIX = "admin.quizzes.form"

const createSchema = z
  .object({
    class_id: z.string().uuid(),
    class_session_id: z.string().uuid(),
    title: z.string().min(2),
    description: z.string().optional(),
    duration_minutes: z.coerce.number().int().gt(0),
    no_back_navigation: z.boolean(),
    shuffle_questions: z.boolean(),
    started_at: z.string().min(1),
    ended_at: z.string().min(1),
  })
  .refine((v) => new Date(v.ended_at).getTime() > new Date(v.started_at).getTime(), {
    path: ["ended_at"],
    message: "end_after_start",
  })

const editSchema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().int().gt(0),
  no_back_navigation: z.boolean(),
  shuffle_questions: z.boolean(),
})

type CreateInput = z.input<typeof createSchema>
type CreateValues = z.infer<typeof createSchema>
type EditInput = z.input<typeof editSchema>
type EditValues = z.infer<typeof editSchema>

const buildCreateDefaults = (classId?: string): CreateValues => ({
  class_id: classId ?? "",
  class_session_id: "",
  title: "",
  description: "",
  duration_minutes: 30,
  no_back_navigation: false,
  shuffle_questions: false,
  started_at: "",
  ended_at: "",
})

const quizToEditValues = (quiz: Quiz): EditValues => ({
  title: quiz.title ?? "",
  description: quiz.description ?? "",
  duration_minutes: quiz.duration_minutes ?? 30,
  no_back_navigation: quiz.no_back_navigation ?? false,
  shuffle_questions: quiz.shuffle_questions ?? false,
})

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

  return isEdit && quiz ? (
    <EditDialog
      key={quiz.id}
      open={open}
      onOpenChange={onOpenChange}
      quiz={quiz}
      onInvalidate={() => invalidateQuizCaches(queryClient)}
      t={t}
    />
  ) : (
    <CreateDialog
      open={open}
      onOpenChange={onOpenChange}
      defaultClassId={defaultClassId}
      onInvalidate={() => invalidateQuizCaches(queryClient)}
      t={t}
    />
  )
}

function invalidateQuizCaches(queryClient: ReturnType<typeof useQueryClient>) {
  queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
  queryClient.invalidateQueries({ queryKey: getGetAdminQuizzesQueryKey() })
}

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  defaultClassId?: string
  onInvalidate: () => void
  t: (key: string) => string
}

function CreateDialog({
  open,
  onOpenChange,
  defaultClassId,
  onInvalidate,
  t,
}: CreateDialogProps) {
  const form = useForm<CreateInput, unknown, CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: buildCreateDefaults(defaultClassId),
  })

  useEffect(() => {
    if (open) form.reset(buildCreateDefaults(defaultClassId))
  }, [open, defaultClassId])

  const mutation = usePostQuizzes({
    mutation: {
      onSuccess: async (res) => {
        const sessionId = form.getValues("class_session_id")
        const quizId = res.status === 201 ? res.data.data?.id : undefined
        if (sessionId && quizId) {
          try {
            await postQuizzesIdRooms(quizId, {
              class_session_id: sessionId,
              started_at: new Date(form.getValues("started_at")).toISOString(),
              ended_at: new Date(form.getValues("ended_at")).toISOString(),
            })
          } catch {
            toast.error(t(`${TRANSLATION_PREFIX}.linkRoomFailed`))
          }
        }
        toast.success(t(`${TRANSLATION_PREFIX}.createSuccess`))
        onInvalidate()
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    mutation.mutate({
      data: {
        class_id: values.class_id,
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
        no_back_navigation: values.no_back_navigation,
        shuffle_questions: values.shuffle_questions,
      },
    })
  })

  const errors = form.formState.errors
  const classId = form.watch("class_id")
  const sessionId = form.watch("class_session_id")
  const noBack = form.watch("no_back_navigation")
  const shuffle = form.watch("shuffle_questions")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t(`${TRANSLATION_PREFIX}.createTitle`)}
      description={t(`${TRANSLATION_PREFIX}.createDescription`)}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.class_id || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.class`)}</FieldLabel>
          <ClassPicker
            value={classId || undefined}
            onChange={(id) => {
              form.setValue("class_id", id, { shouldValidate: true })
              form.setValue("class_session_id", "", { shouldValidate: true })
            }}
            placeholder={t(`${TRANSLATION_PREFIX}.classPlaceholder`)}
          />
          <FieldError errors={[errors.class_id]} />
        </Field>
        <Field data-invalid={!!errors.class_session_id || undefined}>
          <FieldLabel>{t(`${TRANSLATION_PREFIX}.session`)}</FieldLabel>
          <SessionPicker
            classId={classId || undefined}
            value={sessionId || undefined}
            onChange={(id) =>
              form.setValue("class_session_id", id, { shouldValidate: true })
            }
            placeholder={t(`${TRANSLATION_PREFIX}.sessionPlaceholder`)}
          />
          <FieldError errors={[errors.class_session_id]} />
        </Field>

        <QuizCoreFields register={form.register} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizScheduleFields register={form.register} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizFlagsFields
          prefix={TRANSLATION_PREFIX}
          noBackNavigation={noBack}
          shuffleQuestions={shuffle}
          onNoBackNavigationChange={(v) => form.setValue("no_back_navigation", v)}
          onShuffleQuestionsChange={(v) => form.setValue("shuffle_questions", v)}
        />
      </FieldGroup>
    </ResourceFormDialog>
  )
}

interface EditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  quiz: Quiz
  onInvalidate: () => void
  t: (key: string) => string
}

function EditDialog({ open, onOpenChange, quiz, onInvalidate, t }: EditDialogProps) {
  const form = useForm<EditInput, unknown, EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: quizToEditValues(quiz),
  })

  useEffect(() => {
    if (open) form.reset(quizToEditValues(quiz))
  }, [open, quiz])

  const mutation = usePutQuizzesId({
    mutation: {
      onSuccess: () => {
        toast.success(t(`${TRANSLATION_PREFIX}.updateSuccess`))
        onInvalidate()
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    if (!quiz.id) return
    mutation.mutate({ id: quiz.id, data: values })
  })

  const errors = form.formState.errors
  const noBack = form.watch("no_back_navigation")
  const shuffle = form.watch("shuffle_questions")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t(`${TRANSLATION_PREFIX}.editTitle`)}
      description={t(`${TRANSLATION_PREFIX}.editDescription`)}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <QuizCoreFields register={form.register} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizFlagsFields
          prefix={TRANSLATION_PREFIX}
          noBackNavigation={noBack}
          shuffleQuestions={shuffle}
          onNoBackNavigationChange={(v) => form.setValue("no_back_navigation", v)}
          onShuffleQuestionsChange={(v) => form.setValue("shuffle_questions", v)}
        />
      </FieldGroup>
    </ResourceFormDialog>
  )
}

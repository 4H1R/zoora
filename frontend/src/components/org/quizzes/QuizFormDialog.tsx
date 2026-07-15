import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetQuizzesQueryKey, postQuizzesIdRooms, usePostQuizzes, usePutQuizzesId } from "@/api/quizzes/quizzes"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import {
  antiCheatDefaults,
  antiCheatFromQuiz,
  antiCheatSchemaShape,
  QuizCoreFields,
  QuizFlagsFields,
  QuizScheduleFields,
} from "@/components/quizzes/quiz-form-fields"
import { FieldGroup } from "@/components/ui/field"
import { isPlanError } from "@/lib/plan-errors"

const TRANSLATION_PREFIX = "org.session.quizzes.form"

const coreFields = {
  title: z.string().min(2),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().int().gt(0),
  ...antiCheatSchemaShape,
}

const createSchema = z
  .object({
    ...coreFields,
    started_at: z.string().min(1),
    ended_at: z.string().min(1),
  })
  .refine((v) => new Date(v.ended_at).getTime() > new Date(v.started_at).getTime(), {
    path: ["ended_at"],
    params: { i18n: "validation.endAfterStart" },
  })

const editSchema = z.object(coreFields)

type CreateInput = z.input<typeof createSchema>
type CreateValues = z.infer<typeof createSchema>
type EditInput = z.input<typeof editSchema>
type EditValues = z.infer<typeof editSchema>

const createDefaults: CreateValues = {
  title: "",
  description: "",
  duration_minutes: 30,
  ...antiCheatDefaults,
  started_at: "",
  ended_at: "",
}

const quizToEditValues = (quiz: Quiz): EditValues => ({
  title: quiz.title ?? "",
  description: quiz.description ?? "",
  duration_minutes: quiz.duration_minutes ?? 30,
  ...antiCheatFromQuiz(quiz),
})

interface QuizFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  quiz?: Quiz | null
  classId: string
  classSessionId: string
}

export function QuizFormDialog({ open, onOpenChange, quiz, classId, classSessionId }: QuizFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })

  return quiz ? (
    <EditDialog key={quiz.id} open={open} onOpenChange={onOpenChange} quiz={quiz} onInvalidate={invalidate} t={t} />
  ) : (
    <CreateDialog
      open={open}
      onOpenChange={onOpenChange}
      classId={classId}
      classSessionId={classSessionId}
      onInvalidate={invalidate}
      t={t}
    />
  )
}

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  classSessionId: string
  onInvalidate: () => void
  t: (key: string) => string
}

function CreateDialog({ open, onOpenChange, classId, classSessionId, onInvalidate, t }: CreateDialogProps) {
  const form = useForm<CreateInput, unknown, CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: createDefaults,
  })

  useEffect(() => {
    if (open) form.reset(createDefaults)
  }, [open])

  const mutation = usePostQuizzes({
    mutation: {
      onSuccess: async (res) => {
        const quizId = res.status === 201 ? res.data.data?.id : undefined
        if (quizId) {
          try {
            await postQuizzesIdRooms(quizId, {
              class_session_id: classSessionId,
              started_at: form.getValues("started_at"),
              ended_at: form.getValues("ended_at"),
            })
          } catch {
            toast.error(t(`${TRANSLATION_PREFIX}.linkRoomFailed`))
          }
        }
        toast.success(t(`${TRANSLATION_PREFIX}.createSuccess`))
        onInvalidate()
        onOpenChange(false)
      },
      // Plan-gate 402s (e.g. advanced anti-cheat) get a central upgrade toast
      // (main.tsx); only surface the generic failure here.
      onError: (error) => {
        if (isPlanError(error)) return
        toast.error(t(`${TRANSLATION_PREFIX}.createFailed`))
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    mutation.mutate({
      data: {
        class_id: classId,
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
        ...antiCheatFromQuiz(values),
      },
    })
  })

  const errors = form.formState.errors

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
        <QuizCoreFields register={form.register} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizScheduleFields control={form.control as never} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizFlagsFields values={antiCheatFromQuiz(form.watch())} onChange={(k, v) => form.setValue(k, v)} />
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
        <QuizFlagsFields values={antiCheatFromQuiz(form.watch())} onChange={(k, v) => form.setValue(k, v)} />
      </FieldGroup>
    </ResourceFormDialog>
  )
}

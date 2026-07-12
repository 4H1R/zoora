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
  antiCheatDefaults,
  antiCheatFromQuiz,
  antiCheatSchemaShape,
  QuizCoreFields,
  QuizFlagsFields,
  QuizScheduleFields,
} from "@/components/quizzes/quiz-form-fields"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { isPlanError } from "@/lib/plan-errors"

const TRANSLATION_PREFIX = "admin.quizzes.form"

const NEGATIVE_MODES = ["none", "per_wrong", "accumulative"] as const
type NegativeMode = (typeof NEGATIVE_MODES)[number]

interface QuizNegativeFieldsProps {
  mode: NegativeMode
  negativeValue: number
  wrongsPerPoint: number
  onModeChange: (mode: NegativeMode) => void
  onNegativeValueChange: (value: number) => void
  onWrongsPerPointChange: (value: number) => void
  t: (key: string) => string
}

function QuizNegativeFields({
  mode,
  negativeValue,
  wrongsPerPoint,
  onModeChange,
  onNegativeValueChange,
  onWrongsPerPointChange,
  t,
}: QuizNegativeFieldsProps) {
  return (
    <Field>
      <FieldLabel>{t("admin.quizzes.form.negativeMark.label")}</FieldLabel>
      <Select value={mode} onValueChange={(v) => onModeChange(v as NegativeMode)}>
        <SelectTrigger>
          <SelectValue>
            {(value: NegativeMode) => t(`admin.questions.form.negativeMark.modes.${value}`)}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {NEGATIVE_MODES.map((m) => (
            <SelectItem key={m} value={m}>
              {t(`admin.questions.form.negativeMark.modes.${m}`)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {mode === "per_wrong" && (
        <div className="mt-2 flex flex-col gap-1">
          <FieldLabel>{t("admin.questions.form.negativeMark.negativeValue")}</FieldLabel>
          <Input
            type="number"
            step="any"
            min={0}
            value={negativeValue}
            onChange={(e) => onNegativeValueChange(Number(e.target.value))}
          />
        </div>
      )}
      {mode === "accumulative" && (
        <div className="mt-2 flex flex-col gap-1">
          <FieldLabel>{t("admin.questions.form.negativeMark.wrongsPerPoint")}</FieldLabel>
          <Input
            type="number"
            min={2}
            max={5}
            step={1}
            value={wrongsPerPoint}
            onChange={(e) => onWrongsPerPointChange(Number(e.target.value))}
          />
        </div>
      )}
      <p className="text-muted-foreground text-xs">
        {t("admin.quizzes.form.negativeMark.hint")}
      </p>
    </Field>
  )
}

const createSchema = z
  .object({
    class_id: z.string().uuid(),
    class_session_id: z.string().uuid(),
    title: z.string().min(2),
    description: z.string().optional(),
    duration_minutes: z.coerce.number().int().gt(0),
    ...antiCheatSchemaShape,
    started_at: z.string().min(1),
    ended_at: z.string().min(1),
    negative_mark_mode: z.enum(NEGATIVE_MODES),
    negative_value: z.coerce.number(),
    wrongs_per_point: z.coerce.number().int(),
  })
  .refine((v) => new Date(v.ended_at).getTime() > new Date(v.started_at).getTime(), {
    path: ["ended_at"],
    params: { i18n: "validation.endAfterStart" },
  })

const editSchema = z.object({
  title: z.string().min(2),
  description: z.string().optional(),
  duration_minutes: z.coerce.number().int().gt(0),
  ...antiCheatSchemaShape,
  negative_mark_mode: z.enum(NEGATIVE_MODES),
  negative_value: z.coerce.number(),
  wrongs_per_point: z.coerce.number().int(),
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
  ...antiCheatDefaults,
  started_at: "",
  ended_at: "",
  negative_mark_mode: "none",
  negative_value: 0,
  wrongs_per_point: 0,
})

const quizToEditValues = (quiz: Quiz): EditValues => ({
  title: quiz.title ?? "",
  description: quiz.description ?? "",
  duration_minutes: quiz.duration_minutes ?? 30,
  ...antiCheatFromQuiz(quiz),
  negative_mark_mode: (quiz.negative_mark_mode as NegativeMode) ?? "none",
  negative_value: quiz.negative_value ?? 0,
  wrongs_per_point: quiz.wrongs_per_point ?? 0,
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

// negativePayload normalizes the quiz-wide negative fields so only the fields
// relevant to the selected mode are non-zero (mirrors backend normalization).
function negativePayload(values: {
  negative_mark_mode: NegativeMode
  negative_value: number
  wrongs_per_point: number
}) {
  const mode = values.negative_mark_mode
  return {
    negative_mark_mode: mode,
    negative_value: mode === "per_wrong" ? values.negative_value : 0,
    wrongs_per_point: mode === "accumulative" ? values.wrongs_per_point : 0,
  }
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
        class_id: values.class_id,
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
        ...antiCheatFromQuiz(values),
        ...negativePayload(values),
      },
    })
  })

  const errors = form.formState.errors
  const classId = form.watch("class_id")
  const sessionId = form.watch("class_session_id")
  const negMode = form.watch("negative_mark_mode") as NegativeMode
  const negValue = form.watch("negative_value") as number
  const negWpp = form.watch("wrongs_per_point") as number

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
        <QuizScheduleFields control={form.control as never} errors={errors} prefix={TRANSLATION_PREFIX} />
        <QuizFlagsFields
          values={antiCheatFromQuiz(form.watch())}
          onChange={(k, v) => form.setValue(k, v)}
        />
        <QuizNegativeFields
          mode={negMode}
          negativeValue={negValue}
          wrongsPerPoint={negWpp}
          onModeChange={(v) => form.setValue("negative_mark_mode", v)}
          onNegativeValueChange={(v) => form.setValue("negative_value", v)}
          onWrongsPerPointChange={(v) => form.setValue("wrongs_per_point", v)}
          t={t}
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
    mutation.mutate({
      id: quiz.id,
      data: {
        title: values.title,
        description: values.description,
        duration_minutes: values.duration_minutes,
        ...antiCheatFromQuiz(values),
        ...negativePayload(values),
      },
    })
  })

  const errors = form.formState.errors
  const negMode = form.watch("negative_mark_mode") as NegativeMode
  const negValue = form.watch("negative_value") as number
  const negWpp = form.watch("wrongs_per_point") as number

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
          values={antiCheatFromQuiz(form.watch())}
          onChange={(k, v) => form.setValue(k, v)}
        />
        <QuizNegativeFields
          mode={negMode}
          negativeValue={negValue}
          wrongsPerPoint={negWpp}
          onModeChange={(v) => form.setValue("negative_mark_mode", v)}
          onNegativeValueChange={(v) => form.setValue("negative_value", v)}
          onWrongsPerPointChange={(v) => form.setValue("wrongs_per_point", v)}
          t={t}
        />
      </FieldGroup>
    </ResourceFormDialog>
  )
}

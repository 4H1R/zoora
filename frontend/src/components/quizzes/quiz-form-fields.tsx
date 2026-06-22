import type {
  Control,
  FieldError as RHFFieldError,
  FieldErrors,
  UseFormRegister,
} from "react-hook-form"

import { Controller } from "react-hook-form"
import { useTranslation } from "react-i18next"

import { BooleanFieldGroup, BooleanFieldRow } from "@/components/form/boolean-field-row"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

export interface QuizCoreValues {
  title: string
  description?: string
  // `unknown` because callers feed in the zod *input* type, where
  // `z.coerce.number()` fields are typed `unknown` before coercion.
  duration_minutes: unknown
  no_back_navigation: boolean
  shuffle_questions: boolean
}

export interface QuizScheduleValues {
  started_at: string
  ended_at: string
}

interface QuizCoreFieldsProps {
  // Loose register so callers with stricter generics can pass it without casts.
  register: UseFormRegister<any>
  errors: FieldErrors<QuizCoreValues>
  prefix: string
}

export function QuizCoreFields({ register, errors, prefix }: QuizCoreFieldsProps) {
  const { t } = useTranslation()
  return (
    <>
      <Field data-invalid={!!errors.title || undefined}>
        <FieldLabel>{t(`${prefix}.title`)}</FieldLabel>
        <Input {...register("title")} placeholder={t(`${prefix}.titlePlaceholder`)} />
        <FieldError errors={[errors.title as RHFFieldError | undefined]} />
      </Field>
      <Field>
        <FieldLabel>{t(`${prefix}.description`)}</FieldLabel>
        <Textarea
          {...register("description")}
          placeholder={t(`${prefix}.descriptionPlaceholder`)}
          rows={3}
        />
      </Field>
      <Field data-invalid={!!errors.duration_minutes || undefined}>
        <FieldLabel>{t(`${prefix}.duration`)}</FieldLabel>
        <Input type="number" min={1} {...register("duration_minutes")} />
        <FieldError errors={[errors.duration_minutes as RHFFieldError | undefined]} />
      </Field>
    </>
  )
}

interface QuizScheduleFieldsProps {
  // Loose control so callers with stricter generics can pass it without casts.
  control: Control<any, any, any>
  errors: FieldErrors<QuizScheduleValues>
  prefix: string
}

export function QuizScheduleFields({ control, errors, prefix }: QuizScheduleFieldsProps) {
  const { t } = useTranslation()
  const endedAtError = errors.ended_at as RHFFieldError | undefined
  return (
    <div className="grid gap-3 sm:grid-cols-2">
      <Field data-invalid={!!errors.started_at || undefined}>
        <FieldLabel>{t(`${prefix}.startedAt`)}</FieldLabel>
        <Controller
          control={control}
          name="started_at"
          render={({ field, fieldState }) => (
            <DateTimePicker
              value={field.value || undefined}
              onChange={(v) => field.onChange(v ?? "")}
              invalid={fieldState.invalid}
            />
          )}
        />
        <p className="text-muted-foreground text-xs">{t(`${prefix}.startedAtHint`)}</p>
        <FieldError errors={[errors.started_at as RHFFieldError | undefined]} />
      </Field>
      <Field data-invalid={!!endedAtError || undefined}>
        <FieldLabel>{t(`${prefix}.endedAt`)}</FieldLabel>
        <Controller
          control={control}
          name="ended_at"
          render={({ field, fieldState }) => (
            <DateTimePicker
              value={field.value || undefined}
              onChange={(v) => field.onChange(v ?? "")}
              invalid={fieldState.invalid}
            />
          )}
        />
        <p className="text-muted-foreground text-xs">{t(`${prefix}.endedAtHint`)}</p>
        <FieldError errors={[endedAtError]} />
      </Field>
    </div>
  )
}

interface QuizFlagsFieldsProps {
  prefix: string
  noBackNavigation: boolean
  shuffleQuestions: boolean
  onNoBackNavigationChange: (value: boolean) => void
  onShuffleQuestionsChange: (value: boolean) => void
}

export function QuizFlagsFields({
  prefix,
  noBackNavigation,
  shuffleQuestions,
  onNoBackNavigationChange,
  onShuffleQuestionsChange,
}: QuizFlagsFieldsProps) {
  const { t } = useTranslation()
  return (
    <BooleanFieldGroup>
      <BooleanFieldRow
        label={t(`${prefix}.noBackNavigation`)}
        hint={t(`${prefix}.noBackNavigationHint`)}
        checked={noBackNavigation}
        onCheckedChange={onNoBackNavigationChange}
      />
      <BooleanFieldRow
        label={t(`${prefix}.shuffleQuestions`)}
        hint={t(`${prefix}.shuffleQuestionsHint`)}
        checked={shuffleQuestions}
        onCheckedChange={onShuffleQuestionsChange}
      />
    </BooleanFieldGroup>
  )
}

import type { GithubCom4H1RZooraInternalDomainQuestion as Question } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { PlusIcon, Trash2Icon } from "lucide-react"
import { useEffect } from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetQuestionBanksIdQuestionsQueryKey,
  usePostQuestionBanksIdQuestions,
  usePutQuestionBanksQuestionsQuestionId,
} from "@/api/question-banks/question-banks"
import { BankPicker } from "@/components/admin/forms/BankPicker"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

import { OptionImageControl } from "./OptionImage"
import { QuestionPhotoUploader } from "./QuestionPhotoUploader"

const TYPE_VALUES = ["descriptive", "short_answer", "choice"] as const
type QType = (typeof TYPE_VALUES)[number]

const NEGATIVE_MODES = ["none", "per_wrong", "accumulative"] as const
type NegativeMode = (typeof NEGATIVE_MODES)[number]

const optionSchema = z.object({
  id: z.string().min(1),
  value: z.string(),
  score: z.coerce.number(),
  image_media_id: z.string().nullable().optional(),
})

// fractionFor mirrors the backend domain.FractionFor: suggested per-wrong
// fraction for an option count {2:0.5,3:0.33,4:0.25,5:0.2} else 1/n.
function fractionFor(optionCount: number): number {
  switch (optionCount) {
    case 2:
      return 0.5
    case 3:
      return 0.33
    case 4:
      return 0.25
    case 5:
      return 0.2
  }
  if (optionCount <= 0) return 0
  return 1 / optionCount
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

// suggestNegative mirrors the backend auto-prefill: per_wrong uses fractionFor,
// accumulative uses clamp(optionCount, 2, 5).
function suggestNegative(mode: NegativeMode, optionCount: number): {
  negative_value: number
  wrongs_per_point: number
} {
  if (mode === "per_wrong") {
    return { negative_value: fractionFor(optionCount), wrongs_per_point: 0 }
  }
  if (mode === "accumulative") {
    return { negative_value: 0, wrongs_per_point: clamp(optionCount, 2, 5) }
  }
  return { negative_value: 0, wrongs_per_point: 0 }
}

const metadataSchema = z.object({
  type: z.literal("photo"),
  media_id: z.string().uuid(),
})

const baseSchema = z.object({
  bank_id: z.string().uuid().optional(),
  text: z.string().min(1),
  type: z.enum(TYPE_VALUES),
  options: z.array(optionSchema),
  metadata: z.array(metadataSchema),
  negative_mark_mode: z.enum(NEGATIVE_MODES),
  negative_value: z.coerce.number(),
  wrongs_per_point: z.coerce.number().int(),
})

type FormInput = z.input<typeof baseSchema>
type FormValues = z.infer<typeof baseSchema>
type FormOption = FormValues["options"][number]
type FormMetadata = FormValues["metadata"][number]

interface QuestionCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  question?: Question | null
  defaultBankId?: string
}

function nextOptId() {
  return crypto.randomUUID()
}

function defaultOptionsFor(type: QType): FormOption[] {
  switch (type) {
    case "choice":
      return [
        { id: nextOptId(), value: "", score: 1 },
        { id: nextOptId(), value: "", score: 0 },
      ]
    case "short_answer":
      return [{ id: nextOptId(), value: "", score: 1 }]
    case "descriptive":
      return [{ id: nextOptId(), value: "", score: 1 }]
  }
}

function validateOptionsFor(type: QType, options: FormOption[]): string | null {
  if (type === "choice") {
    if (options.length < 2) return "choice"
    for (const o of options) if (!o.value?.trim()) return "valueRequired"
  } else if (type === "short_answer") {
    if (options.length < 1) return "shortAnswer"
    for (const o of options) {
      if (!o.value?.trim()) return "valueRequired"
      if ((o.score ?? 0) < 0) return "negative"
    }
  } else {
    if (options.length < 1) return "descriptive"
    for (const o of options) if ((o.score ?? 0) < 0) return "negative"
  }
  return null
}

export function QuestionCreateModal({
  open,
  onOpenChange,
  question,
  defaultBankId,
}: QuestionCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!question

  const form = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(baseSchema),
    defaultValues: {
      bank_id: "",
      text: "",
      type: "descriptive",
      options: defaultOptionsFor("descriptive"),
      metadata: [],
      negative_mark_mode: "none",
      negative_value: 0,
      wrongs_per_point: 0,
    },
  })

  const optsArr = useFieldArray({ control: form.control, name: "options" })

  useEffect(() => {
    if (!open) return
    if (isEdit && question) {
      const type = ((question.type as QType) ?? "descriptive") as QType
      const opts = question.options?.length
        ? question.options.map((o) => ({
            id: o.id ?? nextOptId(),
            value: o.value ?? "",
            score: o.score ?? 0,
            image_media_id: o.image_media_id ?? null,
          }))
        : defaultOptionsFor(type)
      form.reset({
        bank_id: question.bank_id ?? "",
        text: question.text ?? "",
        type,
        options: opts,
        metadata: (question.metadata ?? []).map((m) => ({
          type: "photo" as const,
          media_id: m.media_id ?? "",
        })),
        negative_mark_mode: (question.negative_mark_mode as NegativeMode) ?? "none",
        negative_value: question.negative_value ?? 0,
        wrongs_per_point: question.wrongs_per_point ?? 0,
      })
    } else {
      form.reset({
        bank_id: defaultBankId ?? "",
        text: "",
        type: "descriptive",
        options: defaultOptionsFor("descriptive"),
        metadata: [],
        negative_mark_mode: "none",
        negative_value: 0,
        wrongs_per_point: 0,
      })
    }
  }, [open, question, isEdit, defaultBankId])

  const invalidate = (bankId?: string) => {
    const id = bankId ?? question?.bank_id ?? form.getValues("bank_id")
    if (id) {
      queryClient.invalidateQueries({
        queryKey: getGetQuestionBanksIdQuestionsQueryKey(id),
      })
    }
    queryClient.invalidateQueries({ queryKey: ["getAdminQuestions"] })
  }

  const createMutation = usePostQuestionBanksIdQuestions({
    mutation: {
      onSuccess: (_, variables) => {
        toast.success(t("admin.questions.form.createSuccess"))
        invalidate(variables.id)
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutQuestionBanksQuestionsQuestionId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.questions.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const type = form.watch("type") as QType
  const bankId = form.watch("bank_id")
  const metadata = form.watch("metadata") as FormMetadata[]
  const negativeMode = form.watch("negative_mark_mode") as NegativeMode

  const handleNegativeModeChange = (next: NegativeMode) => {
    form.setValue("negative_mark_mode", next, { shouldDirty: true })
    const optionCount = (form.getValues("options") as FormOption[]).length
    const suggested = suggestNegative(next, optionCount)
    form.setValue("negative_value", suggested.negative_value, { shouldDirty: true })
    form.setValue("wrongs_per_point", suggested.wrongs_per_point, { shouldDirty: true })
  }

  const handleTypeChange = (next: QType) => {
    form.setValue("type", next, { shouldValidate: true })
    const current = form.getValues("options") as FormOption[]
    if (next === "choice" && current.length < 2) {
      form.setValue("options", defaultOptionsFor("choice"))
    } else if ((next === "short_answer" || next === "descriptive") && current.length < 1) {
      form.setValue("options", defaultOptionsFor(next))
    } else if (next === "descriptive") {
      form.setValue(
        "options",
        current.map((o) => ({ ...o, value: "" }))
      )
    }
  }

  const onSubmit = form.handleSubmit((values) => {
    const err = validateOptionsFor(values.type as QType, values.options as FormOption[])
    if (err) {
      toast.error(t(`admin.questions.form.errors.${err}`))
      return
    }

    const isChoice = values.type === "choice"
    const negativeMode: NegativeMode = isChoice ? values.negative_mark_mode : "none"
    const negativeValue = isChoice && negativeMode === "per_wrong" ? values.negative_value : 0
    const wrongsPerPoint = isChoice && negativeMode === "accumulative" ? values.wrongs_per_point : 0
    const options = values.options.map((o) => ({
      ...o,
      image_media_id: isChoice ? (o.image_media_id ?? undefined) : undefined,
    }))

    if (isEdit) {
      if (!question?.id) return
      updateMutation.mutate({
        questionId: question.id,
        data: {
          text: values.text,
          type: values.type,
          options,
          metadata: values.metadata,
          negative_mark_mode: negativeMode,
          negative_value: negativeValue,
          wrongs_per_point: wrongsPerPoint,
        },
      })
    } else {
      if (!values.bank_id) {
        form.setError("bank_id", {
          message: t("validation.required", { attribute: t("validation.attributes.bank_id") }),
        })
        return
      }
      createMutation.mutate({
        id: values.bank_id,
        data: {
          text: values.text,
          type: values.type,
          options,
          metadata: values.metadata,
          negative_mark_mode: negativeMode,
          negative_value: negativeValue,
          wrongs_per_point: wrongsPerPoint,
        },
      })
    }
  })

  const errors = form.formState.errors
  const showValueField = type !== "descriptive"
  const minOptions = type === "choice" ? 2 : 1

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        isEdit ? t("admin.questions.form.editTitle") : t("admin.questions.form.createTitle")
      }
      description={
        isEdit
          ? t("admin.questions.form.editDescription")
          : t("admin.questions.form.createDescription")
      }
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
      contentClassName="sm:max-w-2xl"
    >
      <FieldGroup>
        {!isEdit && (
          <Field data-invalid={!!errors.bank_id || undefined}>
            <FieldLabel>{t("admin.questions.form.bank")}</FieldLabel>
            <BankPicker
              value={bankId || undefined}
              onChange={(id) => form.setValue("bank_id", id, { shouldValidate: true })}
              placeholder={t("admin.questions.form.bankPlaceholder")}
            />
            <FieldError errors={[errors.bank_id]} />
          </Field>
        )}

        <Field data-invalid={!!errors.text || undefined}>
          <FieldLabel>{t("admin.questions.form.text")}</FieldLabel>
          <Textarea
            {...form.register("text")}
            placeholder={t("admin.questions.form.textPlaceholder")}
            rows={3}
          />
          <FieldError errors={[errors.text]} />
        </Field>

        <Field data-invalid={!!errors.type || undefined}>
          <FieldLabel>{t("admin.questions.form.type")}</FieldLabel>
          <Select value={type} onValueChange={(v) => handleTypeChange(v as QType)}>
            <SelectTrigger>
              <SelectValue>
                {(value: QType) => t(`admin.questions.types.${value}`)}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {TYPE_VALUES.map((v) => (
                <SelectItem key={v} value={v}>
                  {t(`admin.questions.types.${v}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.type]} />
        </Field>

        <Field>
          <FieldLabel className="flex items-center justify-between">
            <span>
              {type === "descriptive"
                ? t("admin.questions.form.maxScore")
                : t("admin.questions.form.options")}
            </span>
            {type !== "descriptive" && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() =>
                  optsArr.append({ id: nextOptId(), value: "", score: 0 })
                }
              >
                <PlusIcon data-icon="inline-start" />
                {t("admin.questions.form.addOption")}
              </Button>
            )}
          </FieldLabel>
          <div className="flex flex-col gap-2">
            {optsArr.fields.map((field, idx) => (
              <div key={field.id} className="flex flex-col gap-2">
                <div className="flex items-start gap-2">
                  {showValueField && (
                    <Input
                      className="flex-1"
                      placeholder={t("admin.questions.form.optionValuePlaceholder")}
                      {...form.register(`options.${idx}.value`)}
                    />
                  )}
                  <Input
                    type="number"
                    step="any"
                    className={showValueField ? "w-24" : "flex-1"}
                    placeholder={t("admin.questions.form.scorePlaceholder")}
                    {...form.register(`options.${idx}.score`, { valueAsNumber: true })}
                  />
                  {type !== "descriptive" && optsArr.fields.length > minOptions && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                      onClick={() => optsArr.remove(idx)}
                    >
                      <Trash2Icon />
                    </Button>
                  )}
                </div>
                {type === "choice" && (
                  <div className="ps-1">
                    <OptionImageControl
                      value={form.watch(`options.${idx}.image_media_id`)}
                      questionId={question?.id}
                      onChange={(id) =>
                        form.setValue(`options.${idx}.image_media_id`, id, {
                          shouldDirty: true,
                        })
                      }
                    />
                  </div>
                )}
              </div>
            ))}
            <p className="text-muted-foreground text-xs">
              {t(`admin.questions.form.hints.${type}`)}
            </p>
          </div>
        </Field>

        {type === "choice" && (
          <Field>
            <FieldLabel>{t("admin.questions.form.negativeMark.label")}</FieldLabel>
            <Select
              value={negativeMode}
              onValueChange={(v) => handleNegativeModeChange(v as NegativeMode)}
            >
              <SelectTrigger>
                <SelectValue>
                  {(value: NegativeMode) =>
                    t(`admin.questions.form.negativeMark.modes.${value}`)
                  }
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
            {negativeMode === "per_wrong" && (
              <div className="mt-2 flex flex-col gap-1">
                <FieldLabel>{t("admin.questions.form.negativeMark.negativeValue")}</FieldLabel>
                <Input
                  type="number"
                  step="any"
                  min={0}
                  {...form.register("negative_value", { valueAsNumber: true })}
                />
                <p className="text-muted-foreground text-xs">
                  {t("admin.questions.form.negativeMark.perWrongHint")}
                </p>
              </div>
            )}
            {negativeMode === "accumulative" && (
              <div className="mt-2 flex flex-col gap-1">
                <FieldLabel>{t("admin.questions.form.negativeMark.wrongsPerPoint")}</FieldLabel>
                <Input
                  type="number"
                  min={2}
                  max={5}
                  step={1}
                  {...form.register("wrongs_per_point", { valueAsNumber: true })}
                />
                <p className="text-muted-foreground text-xs">
                  {t("admin.questions.form.negativeMark.accumulativeHint")}
                </p>
              </div>
            )}
          </Field>
        )}

        <Field>
          <FieldLabel>{t("admin.questions.form.photos.label")}</FieldLabel>
          <QuestionPhotoUploader
            value={metadata}
            onChange={(next) =>
              form.setValue("metadata", next, { shouldValidate: false, shouldDirty: true })
            }
            questionId={question?.id}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

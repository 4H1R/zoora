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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

import { OptionImageControl } from "./OptionImage"
import { QuestionPhotoUploader } from "./QuestionPhotoUploader"

const TYPE_VALUES = ["descriptive", "short_answer", "choice"] as const
type QType = (typeof TYPE_VALUES)[number]

const optionSchema = z.object({
  id: z.string().min(1),
  value: z.string(),
  score: z.coerce.number(),
  image_media_id: z.string().nullable().optional(),
})

const metadataSchema = z.object({
  type: z.literal("photo"),
  media_id: z.string().uuid(),
})

const baseSchema = z.object({
  bank_id: z.string().uuid().optional(),
  text: z.string().min(1),
  type: z.enum(TYPE_VALUES),
  options: z.array(optionSchema),
  model_answer: z.string().optional(),
  metadata: z.array(metadataSchema),
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
      // Single score-holder option: the point value the free-text answer is
      // graded out of. No rubric concepts — grading is manual.
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

export function QuestionCreateModal({ open, onOpenChange, question, defaultBankId }: QuestionCreateModalProps) {
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
      model_answer: "",
      metadata: [],
    },
  })

  const optsArr = useFieldArray({ control: form.control, name: "options" })

  useEffect(() => {
    if (!open) return
    if (isEdit && question) {
      const type = ((question.type as QType) ?? "descriptive") as QType
      let opts: FormOption[]
      if (type === "descriptive") {
        // Descriptive carries a single score-holder option. Collapse any legacy
        // rubric concepts into one Points field, preserving the total weight.
        const total = (question.options ?? []).reduce((s, o) => s + Math.max(0, o.score ?? 0), 0)
        opts = [{ id: question.options?.[0]?.id ?? nextOptId(), value: "", score: total || 1 }]
      } else {
        opts = question.options?.length
          ? question.options.map((o) => ({
              id: o.id ?? nextOptId(),
              value: o.value ?? "",
              score: o.score ?? 0,
              image_media_id: o.image_media_id ?? null,
            }))
          : defaultOptionsFor(type)
      }
      form.reset({
        bank_id: question.bank_id ?? "",
        text: question.text ?? "",
        type,
        options: opts,
        model_answer: question.model_answer ?? "",
        metadata: (question.metadata ?? []).map((m) => ({
          type: "photo" as const,
          media_id: m.media_id ?? "",
        })),
      })
    } else {
      form.reset({
        bank_id: defaultBankId ?? "",
        text: "",
        type: "descriptive",
        options: defaultOptionsFor("descriptive"),
        model_answer: "",
        metadata: [],
      })
    }
  }, [open, question, isEdit, defaultBankId, form])

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

  const handleTypeChange = (next: QType) => {
    form.setValue("type", next, { shouldValidate: true })
    const current = form.getValues("options") as FormOption[]
    if (next === "choice") {
      if (current.length < 2) form.setValue("options", defaultOptionsFor("choice"))
    } else if (next === "short_answer") {
      if (current.length < 1) form.setValue("options", defaultOptionsFor("short_answer"))
    } else {
      // descriptive: a single score-holder option (the Points field). Collapse
      // whatever was there, preserving the total weight as the point value.
      const total = current.reduce((s, o) => s + Math.max(0, Number(o.score) || 0), 0)
      form.setValue("options", [{ id: current[0]?.id ?? nextOptId(), value: "", score: total || 1 }])
    }
  }

  const onSubmit = form.handleSubmit((values) => {
    const err = validateOptionsFor(values.type as QType, values.options as FormOption[])
    if (err) {
      toast.error(t(`admin.questions.form.errors.${err}`))
      return
    }

    const options =
      values.type === "descriptive"
        ? // Single score-holder option: no value, just the point total.
          [{ id: values.options[0]?.id ?? nextOptId(), value: "", score: values.options[0]?.score ?? 0 }]
        : values.options.map((o) => ({
            id: o.id,
            value: o.value,
            score: o.score,
            image_media_id: values.type === "choice" ? (o.image_media_id ?? undefined) : undefined,
          }))
    const modelAnswer = values.type === "descriptive" ? (values.model_answer ?? "") : ""

    if (isEdit) {
      if (!question?.id) return
      updateMutation.mutate({
        questionId: question.id,
        data: {
          text: values.text,
          type: values.type,
          options,
          model_answer: modelAnswer,
          metadata: values.metadata,
          negative_mark_mode: "none",
          negative_value: 0,
          wrongs_per_point: 0,
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
          model_answer: modelAnswer,
          metadata: values.metadata,
          negative_mark_mode: "none",
          negative_value: 0,
          wrongs_per_point: 0,
        },
      })
    }
  })

  const errors = form.formState.errors
  const minOptions = type === "choice" ? 2 : 1

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.questions.form.editTitle") : t("admin.questions.form.createTitle")}
      description={isEdit ? t("admin.questions.form.editDescription") : t("admin.questions.form.createDescription")}
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
          <Textarea {...form.register("text")} placeholder={t("admin.questions.form.textPlaceholder")} rows={3} />
          <FieldError errors={[errors.text]} />
        </Field>

        <Field data-invalid={!!errors.type || undefined}>
          <FieldLabel>{t("admin.questions.form.type")}</FieldLabel>
          <Select value={type} onValueChange={(v) => handleTypeChange(v as QType)}>
            <SelectTrigger>
              <SelectValue>{(value: QType) => t(`admin.questions.types.${value}`)}</SelectValue>
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

        {type === "descriptive" ? (
          // Descriptive answers are graded manually — the teacher just sets the
          // point value the free-text answer is marked out of.
          <Field>
            <FieldLabel>{t("admin.questions.form.points")}</FieldLabel>
            <Input
              type="number"
              step="any"
              className="w-32"
              placeholder={t("admin.questions.form.scorePlaceholder")}
              {...form.register("options.0.score", { valueAsNumber: true })}
            />
            <p className="text-muted-foreground text-xs">{t("admin.questions.form.hints.descriptive")}</p>
          </Field>
        ) : (
          <Field>
            <FieldLabel className="flex items-center justify-between">
              <span>{t("admin.questions.form.options")}</span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => optsArr.append({ id: nextOptId(), value: "", score: type === "choice" ? 0 : 1 })}
              >
                <PlusIcon data-icon="inline-start" />
                {t("admin.questions.form.addOption")}
              </Button>
            </FieldLabel>
            <div className="flex flex-col gap-2">
              {optsArr.fields.map((field, idx) => (
                <div key={field.id} className="flex flex-col gap-2">
                  <div className="flex items-start gap-2">
                    <Input
                      className="flex-1"
                      placeholder={t("admin.questions.form.optionValuePlaceholder")}
                      {...form.register(`options.${idx}.value`)}
                    />
                    <Input
                      type="number"
                      step="any"
                      className="w-24"
                      placeholder={t("admin.questions.form.scorePlaceholder")}
                      {...form.register(`options.${idx}.score`, { valueAsNumber: true })}
                    />
                    {optsArr.fields.length > minOptions && (
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
              <p className="text-muted-foreground text-xs">{t(`admin.questions.form.hints.${type}`)}</p>
            </div>
          </Field>
        )}

        {type === "descriptive" && (
          <Field>
            <FieldLabel>{t("admin.questions.form.modelAnswer")}</FieldLabel>
            <Textarea
              {...form.register("model_answer")}
              placeholder={t("admin.questions.form.modelAnswerPlaceholder")}
              rows={3}
            />
          </Field>
        )}

        <Field>
          <FieldLabel>{t("admin.questions.form.photos.label")}</FieldLabel>
          <QuestionPhotoUploader
            value={metadata}
            onChange={(next) => form.setValue("metadata", next, { shouldValidate: false, shouldDirty: true })}
            questionId={question?.id}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

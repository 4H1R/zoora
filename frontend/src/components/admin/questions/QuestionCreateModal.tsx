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

const TYPE_VALUES = ["descriptive", "short_answer", "choice"] as const

const optionSchema = z.object({
  id: z.string().min(1),
  value: z.string().min(1),
  score: z.coerce.number(),
})

const createSchema = z.object({
  bank_id: z.string().uuid(),
  text: z.string().min(1),
  type: z.enum(TYPE_VALUES),
  options: z.array(optionSchema).optional(),
})

const editSchema = z.object({
  text: z.string().min(1),
  type: z.enum(TYPE_VALUES),
  options: z.array(optionSchema).optional(),
})

type CreateValues = z.infer<typeof createSchema>
type EditValues = z.infer<typeof editSchema>

interface QuestionCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  question?: Question | null
  defaultBankId?: string
}

function nextOptId() {
  return crypto.randomUUID().slice(0, 8)
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

  const createForm = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      bank_id: "",
      text: "",
      type: "descriptive",
      options: [],
    },
  })

  const editForm = useForm<EditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { text: "", type: "descriptive", options: [] },
  })

  const createOpts = useFieldArray({ control: createForm.control, name: "options" })
  const editOpts = useFieldArray({ control: editForm.control, name: "options" })

  useEffect(() => {
    if (!open) return
    if (isEdit && question) {
      editForm.reset({
        text: question.text ?? "",
        type: (question.type as EditValues["type"]) ?? "descriptive",
        options:
          question.options?.map((o) => ({
            id: o.id ?? nextOptId(),
            value: o.value ?? "",
            score: o.score ?? 0,
          })) ?? [],
      })
    } else {
      createForm.reset({
        bank_id: defaultBankId ?? "",
        text: "",
        type: "descriptive",
        options: [],
      })
    }
  }, [open, question, isEdit, defaultBankId])

  const invalidate = (bankId?: string) => {
    const id = bankId ?? question?.bank_id ?? createForm.getValues("bank_id")
    if (id) {
      queryClient.invalidateQueries({
        queryKey: getGetQuestionBanksIdQuestionsQueryKey(id),
      })
    }
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

  const createBankId = createForm.watch("bank_id")
  const createType = createForm.watch("type")
  const editType = editForm.watch("type")

  const onSubmitCreate = createForm.handleSubmit((values) => {
    createMutation.mutate({
      id: values.bank_id,
      data: {
        text: values.text,
        type: values.type,
        options: values.type === "descriptive" ? [] : values.options ?? [],
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (!question?.id) return
    updateMutation.mutate({
      questionId: question.id,
      data: {
        text: values.text,
        type: values.type,
        options: values.type === "descriptive" ? [] : values.options ?? [],
      },
    })
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  const showOptions = (isEdit ? editType : createType) !== "descriptive"

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
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {!isEdit && (
          <Field data-invalid={!!createErrors.bank_id || undefined}>
            <FieldLabel>{t("admin.questions.form.bank")}</FieldLabel>
            <BankPicker
              value={createBankId || undefined}
              onChange={(id) => createForm.setValue("bank_id", id, { shouldValidate: true })}
              placeholder={t("admin.questions.form.bankPlaceholder")}
            />
            <FieldError errors={[createErrors.bank_id]} />
          </Field>
        )}

        <Field data-invalid={!!(isEdit ? editErrors.text : createErrors.text) || undefined}>
          <FieldLabel>{t("admin.questions.form.text")}</FieldLabel>
          <Textarea
            {...(isEdit ? editForm.register("text") : createForm.register("text"))}
            placeholder={t("admin.questions.form.textPlaceholder")}
            rows={3}
          />
          <FieldError errors={[isEdit ? editErrors.text : createErrors.text]} />
        </Field>

        <Field data-invalid={!!(isEdit ? editErrors.type : createErrors.type) || undefined}>
          <FieldLabel>{t("admin.questions.form.type")}</FieldLabel>
          <Select
            value={isEdit ? editType : createType}
            onValueChange={(v) => {
              if (isEdit) {
                editForm.setValue("type", v as EditValues["type"], { shouldValidate: true })
              } else {
                createForm.setValue("type", v as CreateValues["type"], {
                  shouldValidate: true,
                })
              }
            }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TYPE_VALUES.map((v) => (
                <SelectItem key={v} value={v}>
                  {t(`admin.questions.types.${v}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[isEdit ? editErrors.type : createErrors.type]} />
        </Field>

        {showOptions && (
          <Field>
            <FieldLabel className="flex items-center justify-between">
              <span>{t("admin.questions.form.options")}</span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  if (isEdit) {
                    editOpts.append({ id: nextOptId(), value: "", score: 0 })
                  } else {
                    createOpts.append({ id: nextOptId(), value: "", score: 0 })
                  }
                }}
              >
                <PlusIcon data-icon="inline-start" />
                {t("admin.questions.form.addOption")}
              </Button>
            </FieldLabel>
            <div className="flex flex-col gap-2">
              {(isEdit ? editOpts.fields : createOpts.fields).map((field, idx) => (
                <div key={field.id} className="flex items-start gap-2">
                  <Input
                    className="flex-1"
                    placeholder={t("admin.questions.form.optionValuePlaceholder")}
                    {...(isEdit
                      ? editForm.register(`options.${idx}.value`)
                      : createForm.register(`options.${idx}.value`))}
                  />
                  <Input
                    type="number"
                    step="any"
                    className="w-24"
                    placeholder={t("admin.questions.form.scorePlaceholder")}
                    {...(isEdit
                      ? editForm.register(`options.${idx}.score`)
                      : createForm.register(`options.${idx}.score`))}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                    onClick={() => (isEdit ? editOpts.remove(idx) : createOpts.remove(idx))}
                  >
                    <Trash2Icon />
                  </Button>
                </div>
              ))}
              {(isEdit ? editOpts.fields : createOpts.fields).length === 0 && (
                <p className="text-muted-foreground text-xs">
                  {t("admin.questions.form.optionsEmpty")}
                </p>
              )}
              <p className="text-muted-foreground text-xs">
                {t("admin.questions.form.optionsHint")}
              </p>
            </div>
          </Field>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

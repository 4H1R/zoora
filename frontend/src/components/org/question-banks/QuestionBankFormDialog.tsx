import type { GithubCom4H1RZooraInternalDomainQuestionBank as Bank } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetQuestionBanksQueryKey,
  usePostQuestionBanks,
  usePutQuestionBanksId,
} from "@/api/question-banks/question-banks"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const schema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
})

type Values = z.infer<typeof schema>

const emptyDefaults: Values = { name: "", description: "" }

interface QuestionBankFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bank?: Bank | null
}

export function QuestionBankFormDialog({ open, onOpenChange, bank }: QuestionBankFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!bank

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: emptyDefaults,
  })

  useEffect(() => {
    if (!open) return
    form.reset(isEdit && bank ? { name: bank.name ?? "", description: bank.description ?? "" } : emptyDefaults)
  }, [open, bank, isEdit, form])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetQuestionBanksQueryKey() })
  }

  const createMutation = usePostQuestionBanks({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutQuestionBanksId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = form.handleSubmit((values) => {
    if (isEdit && bank?.id) {
      updateMutation.mutate({
        id: bank.id,
        data: { name: values.name, description: values.description },
      })
    } else {
      createMutation.mutate({
        data: { name: values.name, description: values.description },
      })
    }
  })

  const errors = form.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("org.session.questionBanks.form.editTitle") : t("org.session.questionBanks.form.createTitle")}
      description={
        isEdit
          ? t("org.session.questionBanks.form.editDescription")
          : t("org.session.questionBanks.form.createDescription")
      }
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("org.session.questionBanks.form.name")}</FieldLabel>
          <Input {...form.register("name")} placeholder={t("org.session.questionBanks.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>
        <Field>
          <FieldLabel>{t("org.session.questionBanks.form.description")}</FieldLabel>
          <Textarea
            {...form.register("description")}
            placeholder={t("org.session.questionBanks.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

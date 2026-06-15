import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminQuestionBanksQueryKey,
  usePostAdminQuestionBanks,
} from "@/api/admin-questionbanks/admin-questionbanks"
import { getGetQuestionBanksQueryKey } from "@/api/question-banks/question-banks"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useAdminStore } from "@/stores/admin"

const schema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
})

type FormValues = z.infer<typeof schema>

interface BankCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated?: (bankId: string) => void
}

export function BankCreateModal({ open, onOpenChange, onCreated }: BankCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const activeOrganizationId = useAdminStore((s) => s.activeOrganizationId)

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", description: "" },
  })

  useEffect(() => {
    if (open) form.reset({ name: "", description: "" })
  }, [open])

  const mutation = usePostAdminQuestionBanks({
    mutation: {
      onSuccess: (res) => {
        toast.success(t("admin.questionBanks.form.createSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminQuestionBanksQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetQuestionBanksQueryKey() })
        const id = res.status === 201 ? res.data.data?.id : undefined
        if (id && onCreated) onCreated(id)
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    if (!activeOrganizationId) {
      toast.error(t("admin.questionBanks.form.noActiveOrg"))
      return
    }
    mutation.mutate({
      data: {
        organization_id: activeOrganizationId,
        name: values.name,
        description: values.description,
      },
    })
  })

  const errors = form.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.questionBanks.form.createTitle")}
      description={t("admin.questionBanks.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.questionBanks.form.name")}</FieldLabel>
          <Input
            {...form.register("name")}
            placeholder={t("admin.questionBanks.form.namePlaceholder")}
          />
          <FieldError errors={[errors.name]} />
        </Field>
        <Field>
          <FieldLabel>{t("admin.questionBanks.form.description")}</FieldLabel>
          <Textarea
            {...form.register("description")}
            placeholder={t("admin.questionBanks.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

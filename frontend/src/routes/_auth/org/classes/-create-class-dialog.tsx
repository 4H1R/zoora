import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { useAccess } from "react-access-engine"
import { toast } from "sonner"
import { z } from "zod"

import { getGetClassesQueryKey, usePostClasses } from "@/api/classes/classes"
import { UserSelect } from "@/components/form/user-select"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const schema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  total_users: z.coerce.number().min(0).optional(),
  user_id: z.string().uuid().optional(),
})

type FormInput = z.input<typeof schema>
type FormValues = z.infer<typeof schema>

interface CreateClassDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateClassDialog({ open, onOpenChange }: CreateClassDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { can } = useAccess()
  const canCreateAny = can("classes:create_any")

  const form = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", description: "", total_users: 0, user_id: undefined },
  })

  useEffect(() => {
    if (open) {
      form.reset({ name: "", description: "", total_users: 0, user_id: undefined })
    }
  }, [open])

  const mutation = usePostClasses({
    mutation: {
      onSuccess: () => {
        toast.success(t("classesPage.form.createSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetClassesQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    mutation.mutate({
      data: {
        name: values.name,
        description: values.description,
        total_users: values.total_users,
        user_id: values.user_id,
      },
    })
  })

  const { errors } = form.formState
  const selectedUserId = form.watch("user_id")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("classesPage.form.createTitle")}
      description={t("classesPage.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("classesPage.form.name")}</FieldLabel>
          <Input {...form.register("name")} placeholder={t("classesPage.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>
        <Field>
          <FieldLabel>{t("classesPage.form.description")}</FieldLabel>
          <Textarea
            {...form.register("description")}
            placeholder={t("classesPage.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>
        <Field data-invalid={!!errors.total_users || undefined}>
          <FieldLabel>{t("classesPage.form.capacity")}</FieldLabel>
          <Input
            {...form.register("total_users")}
            type="number"
            min={0}
            placeholder={t("classesPage.form.capacityPlaceholder")}
          />
          <FieldError errors={[errors.total_users]} />
        </Field>
        {canCreateAny && (
          <Field data-invalid={!!errors.user_id || undefined}>
            <FieldLabel>{t("classesPage.form.teacher")}</FieldLabel>
            <UserSelect
              value={selectedUserId || undefined}
              onChange={(id) => form.setValue("user_id", id, { shouldValidate: true })}
              placeholder={t("classesPage.form.teacherPlaceholder")}
            />
            <FieldError errors={[errors.user_id]} />
          </Field>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

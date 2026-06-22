import type { GithubCom4H1RZooraInternalDomainClass as Class } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminClassesQueryKey,
  usePostAdminClasses,
  usePutAdminClassesId,
} from "@/api/admin-classes/admin-classes"
import { OrganizationSelect } from "@/components/form/organization-select"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { UserSelect } from "@/components/form/user-select"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

const createSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  total_users: z.coerce.number().min(0).optional(),
  organization_id: z.string().uuid(),
  user_id: z.string().uuid(),
})

const editSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  total_users: z.coerce.number().min(0).optional(),
})

type ClassCreateInput = z.input<typeof createSchema>
type ClassCreateValues = z.infer<typeof createSchema>
type ClassEditInput = z.input<typeof editSchema>
type ClassEditValues = z.infer<typeof editSchema>

interface ClassFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  cls?: Class | null
}

export function ClassFormDialog({ open, onOpenChange, cls }: ClassFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!cls

  const createForm = useForm<ClassCreateInput, unknown, ClassCreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: { name: "", description: "", total_users: 0, organization_id: "", user_id: "" },
  })

  const editForm = useForm<ClassEditInput, unknown, ClassEditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { name: "", description: "", total_users: 0 },
  })

  useEffect(() => {
    if (open) {
      if (isEdit) {
        editForm.reset({
          name: cls?.name ?? "",
          description: cls?.description ?? "",
          total_users: cls?.total_users ?? 0,
        })
      } else {
        createForm.reset({ name: "", description: "", total_users: 0, organization_id: "", user_id: "" })
      }
    }
  }, [open, cls, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminClassesQueryKey() })
  }

  const createMutation = usePostAdminClasses({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.classes.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutAdminClassesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.classes.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const selectedOrgId = createForm.watch("organization_id")
  const selectedUserId = createForm.watch("user_id")

  const onSubmitCreate = createForm.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        name: values.name,
        description: values.description,
        total_users: values.total_users,
        organization_id: values.organization_id,
        user_id: values.user_id,
      },
    })
  })

  const onSubmitEdit = editForm.handleSubmit((values) => {
    if (cls?.id) {
      updateMutation.mutate({
        id: cls.id,
        data: { name: values.name, description: values.description, total_users: values.total_users },
      })
    }
  })

  const createErrors = createForm.formState.errors
  const editErrors = editForm.formState.errors

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.classes.form.editTitle") : t("admin.classes.form.createTitle")}
      description={isEdit ? t("admin.classes.form.editDescription") : t("admin.classes.form.createDescription")}
      onSubmit={isEdit ? onSubmitEdit : onSubmitCreate}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        {isEdit ? (
          <>
            <Field data-invalid={!!editErrors.name || undefined}>
              <FieldLabel>{t("admin.classes.form.name")}</FieldLabel>
              <Input {...editForm.register("name")} placeholder={t("admin.classes.form.namePlaceholder")} />
              <FieldError errors={[editErrors.name]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.classes.form.description")}</FieldLabel>
              <Textarea
                {...editForm.register("description")}
                placeholder={t("admin.classes.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field data-invalid={!!editErrors.total_users || undefined}>
              <FieldLabel>{t("admin.classes.form.capacity")}</FieldLabel>
              <Input
                {...editForm.register("total_users")}
                type="number"
                min={0}
                placeholder={t("admin.classes.form.capacityPlaceholder")}
              />
              <FieldError errors={[editErrors.total_users]} />
            </Field>
          </>
        ) : (
          <>
            <Field data-invalid={!!createErrors.name || undefined}>
              <FieldLabel>{t("admin.classes.form.name")}</FieldLabel>
              <Input {...createForm.register("name")} placeholder={t("admin.classes.form.namePlaceholder")} />
              <FieldError errors={[createErrors.name]} />
            </Field>
            <Field>
              <FieldLabel>{t("admin.classes.form.description")}</FieldLabel>
              <Textarea
                {...createForm.register("description")}
                placeholder={t("admin.classes.form.descriptionPlaceholder")}
                rows={3}
              />
            </Field>
            <Field data-invalid={!!createErrors.total_users || undefined}>
              <FieldLabel>{t("admin.classes.form.capacity")}</FieldLabel>
              <Input
                {...createForm.register("total_users")}
                type="number"
                min={0}
                placeholder={t("admin.classes.form.capacityPlaceholder")}
              />
              <FieldError errors={[createErrors.total_users]} />
            </Field>
            <Field data-invalid={!!createErrors.organization_id || undefined}>
              <FieldLabel>{t("admin.classes.form.organization")}</FieldLabel>
              <OrganizationSelect
                value={selectedOrgId || undefined}
                onChange={(id) => createForm.setValue("organization_id", id, { shouldValidate: true })}
                placeholder={t("admin.classes.form.organizationPlaceholder")}
              />
              <FieldError errors={[createErrors.organization_id]} />
            </Field>
            <Field data-invalid={!!createErrors.user_id || undefined}>
              <FieldLabel>{t("admin.classes.form.teacher")}</FieldLabel>
              <UserSelect
                scope="admin"
                value={selectedUserId || undefined}
                onChange={(id) => createForm.setValue("user_id", id, { shouldValidate: true })}
                placeholder={t("admin.classes.form.teacherPlaceholder")}
                organizationId={selectedOrgId || undefined}
              />
              <FieldError errors={[createErrors.user_id]} />
            </Field>
          </>
        )}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

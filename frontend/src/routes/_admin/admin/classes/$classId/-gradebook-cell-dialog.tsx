import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetClassesIdGradebookQueryKey,
  usePostClassesIdGradebookColumnsColumnIdCells,
} from "@/api/gradebook/gradebook"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"

const schema = z.object({
  value: z.string().min(1),
})

type FormValues = z.infer<typeof schema>

interface GradebookCellDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  columnId: string | undefined
  studentId: string | undefined
  studentName?: string
  columnTitle?: string
  initialValue?: string
}

export function GradebookCellDialog({
  open,
  onOpenChange,
  classId,
  columnId,
  studentId,
  studentName,
  columnTitle,
  initialValue,
}: GradebookCellDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { value: "" },
  })

  useEffect(() => {
    if (open) form.reset({ value: initialValue ?? "" })
  }, [open, initialValue])

  const upsertMutation = usePostClassesIdGradebookColumnsColumnIdCells({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.gradebook.form.cellUpdateSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdGradebookQueryKey(classId) })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    if (!columnId || !studentId) return
    upsertMutation.mutate({
      id: classId,
      columnId,
      data: { student_id: studentId, value: values.value },
    })
  })

  const errors = form.formState.errors
  const description = [columnTitle, studentName].filter(Boolean).join(" · ")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.gradebook.form.cellTitle")}
      description={description || undefined}
      onSubmit={onSubmit}
      isLoading={upsertMutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.value || undefined}>
          <FieldLabel>{t("admin.gradebook.form.value")}</FieldLabel>
          <Input
            {...form.register("value")}
            placeholder={t("admin.gradebook.form.valuePlaceholder")}
            autoFocus
          />
          <FieldError errors={[errors.value]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

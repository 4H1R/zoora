import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminAttendanceQueryKey,
  usePutAdminAttendanceId,
} from "@/api/admin-attendance/admin-attendance"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

const ATTENDANCE_STATUSES = ["present", "absent", "late", "excused"] as const
type AttendanceStatus = (typeof ATTENDANCE_STATUSES)[number]

const editSchema = z.object({
  status: z.enum(ATTENDANCE_STATUSES),
  remarks: z.string().optional(),
})

type AttendanceEditValues = z.infer<typeof editSchema>

interface AttendanceFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  attendance?: Attendance | null
}

export function AttendanceFormDialog({ open, onOpenChange, attendance }: AttendanceFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<AttendanceEditValues>({
    resolver: zodResolver(editSchema),
    defaultValues: { status: "present", remarks: "" },
  })

  useEffect(() => {
    if (open && attendance) {
      form.reset({
        status: (attendance.status as AttendanceStatus | undefined) ?? "present",
        remarks: attendance.remarks ?? "",
      })
    }
  }, [open, attendance, form])

  const updateMutation = usePutAdminAttendanceId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.attendance.form.updateSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminAttendanceQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    if (!attendance?.id) return
    updateMutation.mutate({
      id: attendance.id,
      data: { status: values.status, remarks: values.remarks },
    })
  })

  const errors = form.formState.errors
  const statusValue = form.watch("status")

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.attendance.form.editTitle")}
      description={t("admin.attendance.form.editDescription")}
      onSubmit={onSubmit}
      isLoading={updateMutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.status || undefined}>
          <FieldLabel>{t("admin.attendance.form.status")}</FieldLabel>
          <Select
            value={statusValue ?? null}
            onValueChange={(val) => {
              if (val) form.setValue("status", val as AttendanceStatus, { shouldValidate: true })
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.attendance.form.statusPlaceholder")}>
                {(v: AttendanceStatus) => t(`common.statuses.attendance.${v}`)}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {ATTENDANCE_STATUSES.map((s) => (
                <SelectItem key={s} value={s}>
                  {t(`common.statuses.attendance.${s}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.status]} />
        </Field>
        <Field>
          <FieldLabel>{t("admin.attendance.form.remarks")}</FieldLabel>
          <Textarea
            {...form.register("remarks")}
            placeholder={t("admin.attendance.form.remarksPlaceholder")}
            rows={3}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

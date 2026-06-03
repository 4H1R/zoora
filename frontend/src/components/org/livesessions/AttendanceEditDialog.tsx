import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsSessionIdAttendanceQueryKey,
  usePutAttendanceAttendanceId,
} from "@/api/attendance/attendance"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

const STATUSES = ["present", "absent", "late", "excused"] as const
type Status = (typeof STATUSES)[number]

interface AttendanceEditDialogProps {
  attendance: Attendance | null
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  classSessionId: string
}

export function AttendanceEditDialog({
  attendance,
  open,
  onOpenChange,
  classId,
  classSessionId,
}: AttendanceEditDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [status, setStatus] = useState<Status>("present")
  const [remarks, setRemarks] = useState("")

  useEffect(() => {
    if (open && attendance) {
      setStatus((attendance.status as Status) ?? "present")
      setRemarks(attendance.remarks ?? "")
    }
  }, [open, attendance])

  const mutation = usePutAttendanceAttendanceId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.attendance.form.updateSuccess"))
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!attendance?.id) return
    mutation.mutate({ attendanceId: attendance.id, data: { status, remarks: remarks || undefined } })
  }

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("org.session.attendance.form.editTitle")}
      description={attendance?.user?.name ?? ""}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <Field>
          <FieldLabel>{t("org.session.attendance.form.status")}</FieldLabel>
          <Select value={status} onValueChange={(v) => setStatus(v as Status)}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {STATUSES.map((s) => (
                <SelectItem key={s} value={s}>
                  {t(`org.session.attendance.status.${s}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
        <Field>
          <FieldLabel>{t("org.session.attendance.form.remarks")}</FieldLabel>
          <Textarea
            value={remarks}
            onChange={(e) => setRemarks(e.target.value)}
            placeholder={t("org.session.attendance.form.remarksPlaceholder")}
            rows={3}
          />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

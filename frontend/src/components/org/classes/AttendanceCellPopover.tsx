import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdAttendanceMatrixQueryKey,
  usePostClassesIdSessionsSessionIdAttendance,
  usePutAttendanceAttendanceId,
} from "@/api/attendance/attendance"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

const STATUSES = ["present", "absent", "late", "excused"] as const
export type AttendanceStatus = (typeof STATUSES)[number]

interface AttendanceCellPopoverProps {
  classId: string
  sessionId: string
  studentId: string
  /** Existing attendance record id, or undefined when no record yet. */
  attendanceId?: string
  status?: AttendanceStatus
  disabled?: boolean
  children: React.ReactNode
}

export function AttendanceCellPopover({
  classId,
  sessionId,
  studentId,
  attendanceId,
  status,
  disabled,
  children,
}: AttendanceCellPopoverProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const onSuccess = () => {
    toast.success(t("org.class.attendance.updateSuccess"))
    queryClient.invalidateQueries({ queryKey: getGetClassesIdAttendanceMatrixQueryKey(classId) })
    setOpen(false)
  }

  const createMut = usePostClassesIdSessionsSessionIdAttendance({ mutation: { onSuccess } })
  const updateMut = usePutAttendanceAttendanceId({ mutation: { onSuccess } })
  const isPending = createMut.isPending || updateMut.isPending

  const setStatus = (next: AttendanceStatus) => {
    if (attendanceId) {
      updateMut.mutate({ attendanceId, data: { status: next } })
    } else {
      createMut.mutate({ id: classId, sessionId, data: { user_id: studentId, status: next } })
    }
  }

  if (disabled) return <>{children}</>

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger render={<button type="button" className="inline-flex size-full items-center justify-center" />}>
        {children}
      </PopoverTrigger>
      <PopoverContent className="w-44 p-1" align="center">
        <div className="flex flex-col">
          {STATUSES.map((s) => (
            <Button
              key={s}
              variant={s === status ? "secondary" : "ghost"}
              size="sm"
              className="justify-start"
              disabled={isPending}
              onClick={() => setStatus(s)}
            >
              {t(`org.class.attendance.status.${s}`)}
            </Button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}

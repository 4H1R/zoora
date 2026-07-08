import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { ClipboardCheckIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetAdminAttendance } from "@/api/admin-attendance/admin-attendance"
import { AttendanceFormDialog } from "@/components/admin/attendance/AttendanceFormDialog"
import { AttendanceTable } from "@/components/admin/attendance/AttendanceTable"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema } from "@/lib/data-table"

const ATTENDANCE_STATUSES = ["present", "absent", "late", "excused"] as const
type AttendanceStatus = (typeof ATTENDANCE_STATUSES)[number]

export const Route = createFileRoute("/_admin/admin/attendance/")({
  head: () => adminHead("admin.attendance.title"),
  validateSearch: adminSearchSchema,
  component: AttendancePage,
})

function AttendancePage() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page } = Route.useSearch()
  const currentPage = page ?? 1

  const [classId, setClassId] = useState<string | undefined>(undefined)
  const [sessionId, setSessionId] = useState<string | undefined>(undefined)
  const [status, setStatus] = useState<AttendanceStatus | undefined>(undefined)

  const [editTarget, setEditTarget] = useState<Attendance | null>(null)

  const hasFilters = !!classId || !!sessionId || !!status

  const handleClearFilters = () => {
    setClassId(undefined)
    setSessionId(undefined)
    setStatus(undefined)
  }

  const handleClassChange = (id: string) => {
    setClassId(id || undefined)
    setSessionId(undefined)
  }

  const { data, isLoading } = useGetAdminAttendance({
    class_id: classId || undefined,
    class_session_id: sessionId || undefined,
    status: status || undefined,
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const attendanceData = (data?.status === 200 && data.data.data) || undefined
  const items = attendanceData?.items ?? []
  const total = attendanceData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const statCards = [
    {
      icon: <ClipboardCheckIcon />,
      label: t("admin.attendance.stats.total"),
      value: total,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.attendance.title")} />
      <StatCards stats={statCards} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.attendance.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={handleClassChange} />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.attendance.filter.session")}
          </label>
          <SessionPicker
            classId={classId}
            value={sessionId}
            onChange={(id) => setSessionId(id || undefined)}
          />
        </div>
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.attendance.filter.status")}
          </label>
          <Select
            value={status ?? null}
            onValueChange={(val) => setStatus((val as AttendanceStatus) || undefined)}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.attendance.filter.allStatuses")}>
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
        </div>
        {hasFilters && (
          <Button variant="outline" size="sm" onClick={handleClearFilters}>
            <XIcon data-icon="inline-start" />
            {t("admin.attendance.filter.clear")}
          </Button>
        )}
      </Card>
      <AttendanceTable
        items={items}
        total={total}
        isLoading={isLoading}
        sorting={sorting}
        onEdit={(a) => setEditTarget(a)}
      />
      <AttendanceFormDialog
        open={!!editTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setEditTarget(null)
        }}
        attendance={editTarget}
      />
    </div>
  )
}

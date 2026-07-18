import type { StatItem } from "@/components/data-table/stat-cards"

import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { CalendarCheckIcon, CalendarXIcon, ClockIcon, ShieldCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetAttendanceMe } from "@/api/attendance/attendance"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { StatusTabs } from "@/components/data-table/status-tabs"
import { TableFilter } from "@/components/data-table/table-filter"
import { ClassFilterSelect, SessionFilterSelect } from "@/components/org/class-session-filters"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useAttendanceColumns } from "./-columns"

const attendanceSearchSchema = adminSearchSchema.extend({
  class_id: z.string().optional(),
  class_session_id: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/attendance/")({
  head: () => orgHead("org.nav.attendance"),
  validateSearch: attendanceSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, status, class_id, class_session_id, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["attendance:view"])

  const attendanceQ = useGetAttendanceMe(
    {
      search: search || undefined,
      status: (status as "present" | "absent" | "late" | "excused" | undefined) || undefined,
      class_id: class_id || undefined,
      class_session_id: class_session_id || undefined,
      order_by: order_by || undefined,
      order_dir: order_dir || undefined,
      page: page ?? 1,
      page_size: page_size ?? 20,
    },
    { query: { enabled: allowed } }
  )
  const attendance = (attendanceQ.data?.status === 200 && attendanceQ.data.data.data) || undefined
  const summary = attendance?.summary
  const items = attendance?.items ?? []
  const total = attendance?.total ?? 0
  const loading = attendanceQ.isPending

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useAttendanceColumns()
  const table = useAdminTable({ data: items, columns, rowCount: total, sorting })

  // Changing class invalidates any chosen session — sessions belong to one class.
  const setClass = (classId?: string) =>
    navigate({
      to: ".",
      search: (prev) => ({ ...prev, class_id: classId, class_session_id: undefined, page: 1 }),
    })
  const setSession = (sessionId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_session_id: sessionId, page: 1 }) })

  if (!allowed) return null

  const summaryTotal = (summary?.present ?? 0) + (summary?.absent ?? 0) + (summary?.late ?? 0) + (summary?.excused ?? 0)

  const stats: StatItem[] = [
    { icon: <CalendarCheckIcon />, label: t("common.statuses.attendance.present"), value: summary?.present, loading },
    { icon: <CalendarXIcon />, label: t("common.statuses.attendance.absent"), value: summary?.absent, loading },
    { icon: <ClockIcon />, label: t("common.statuses.attendance.late"), value: summary?.late, loading },
    { icon: <ShieldCheckIcon />, label: t("common.statuses.attendance.excused"), value: summary?.excused, loading },
  ]

  const statusTabs = [
    { value: "all", label: t("org.attendance.tabs.all"), count: summaryTotal },
    { value: "present", label: t("common.statuses.attendance.present"), count: summary?.present },
    { value: "absent", label: t("common.statuses.attendance.absent"), count: summary?.absent },
    { value: "late", label: t("common.statuses.attendance.late"), count: summary?.late },
    { value: "excused", label: t("common.statuses.attendance.excused"), count: summary?.excused },
  ]

  const hasFilters = !!(search || status || class_id || class_session_id)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.attendance.title")} />
        <p className="text-muted-foreground text-sm">{t("org.attendance.subtitle")}</p>
      </div>

      <StatCards stats={stats} className="grid-cols-2 lg:grid-cols-4" />

      <StatusTabs tabs={statusTabs} />

      <TableFilter
        table={table}
        searchPlaceholder={t("org.attendance.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      >
        <ClassFilterSelect value={class_id} onChange={setClass} />
        <SessionFilterSelect classId={class_id} value={class_session_id} onChange={setSession} />
      </TableFilter>

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={loading}
            emptyIcon={<CalendarCheckIcon className="size-8 opacity-40" />}
            emptyTitle={hasFilters ? t("org.attendance.noResults") : t("org.attendance.empty")}
            emptyHint={hasFilters ? t("org.attendance.noResultsHint") : undefined}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}

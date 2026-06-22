import { createFileRoute } from "@tanstack/react-router"
import { CalendarCheckIcon, CalendarXIcon, ClockIcon, ShieldCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetAttendanceMe } from "@/api/attendance/attendance"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards, type StatItem } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useClientTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"

import { useAttendanceColumns } from "./-columns"

export const Route = createFileRoute("/_auth/org/$orgId/attendance/")({
  head: () => orgHead("org.nav.attendance"),
  validateSearch: adminSearchSchema,
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["attendance:view"])

  const attendanceQ = useGetAttendanceMe(undefined, { query: { enabled: allowed } })
  const attendance = (attendanceQ.data?.status === 200 && attendanceQ.data.data.data) || undefined
  const summary = attendance?.summary
  const items = attendance?.items ?? []
  const loading = attendanceQ.isPending

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []
  const columns = useAttendanceColumns()
  const table = useClientTable({
    data: items,
    columns,
    sorting,
    globalFilter: search,
    page,
    pageSize: page_size ?? 8,
  })

  if (!allowed) return null

  const stats: StatItem[] = [
    { icon: <CalendarCheckIcon />, label: t("org.attendance.summary.present"), value: summary?.present, loading },
    { icon: <CalendarXIcon />, label: t("org.attendance.summary.absent"), value: summary?.absent, loading },
    { icon: <ClockIcon />, label: t("org.attendance.summary.late"), value: summary?.late, loading },
    { icon: <ShieldCheckIcon />, label: t("org.attendance.summary.excused"), value: summary?.excused, loading },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.attendance.title")} />

      <StatCards stats={stats} className="grid-cols-2 lg:grid-cols-4" />

      <TableFilter
        table={table}
        searchPlaceholder={t("org.attendance.searchPlaceholder")}
        sortLabel={t("org.attendance.toolbar.sort")}
        columnsLabel={t("org.attendance.toolbar.columns")}
        toggleColumnsLabel={t("org.attendance.toolbar.toggleColumns")}
      />

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={loading}
            emptyIcon={<CalendarCheckIcon className="size-8 opacity-40" />}
            emptyTitle={t("org.attendance.empty")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}

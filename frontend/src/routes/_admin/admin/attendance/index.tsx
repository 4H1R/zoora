import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { ClipboardCheckIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminAttendanceQueryKey,
  useDeleteAdminAttendanceId,
  useGetAdminAttendance,
} from "@/api/admin-attendance/admin-attendance"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { StatCards } from "@/components/data-table/stat-cards"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"
import { adminSearchSchema, useAdminTable } from "@/lib/data-table"

import { AttendanceFormDialog } from "./-attendance-form-dialog"
import { useAttendanceColumns } from "./-columns"

export const Route = createFileRoute("/_admin/admin/attendance/")({
  head: () => adminHead("admin.attendance.title"),
  validateSearch: adminSearchSchema,
  component: AttendancePage,
})

function AttendancePage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { search, order_by, order_dir, page } = Route.useSearch()

  const currentPage = page ?? 1

  const [editTarget, setEditTarget] = useState<Attendance | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Attendance | null>(null)

  const { data, isLoading } = useGetAdminAttendance({
    search: search || undefined,
    page: currentPage,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
  })

  const deleteMutation = useDeleteAdminAttendanceId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.attendance.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminAttendanceQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  const attendanceData = (data?.status === 200 && data.data.data) || undefined
  const items = attendanceData?.items ?? []
  const total = attendanceData?.total ?? 0

  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const columns = useAttendanceColumns({
    onEdit: (a) => setEditTarget(a),
    onDelete: (a) => setDeleteTarget(a),
  })

  const table = useAdminTable({ data: items, columns, rowCount: total, sorting })

  const statCards = [
    {
      icon: <ClipboardCheckIcon />,
      label: t("admin.attendance.stats.total"),
      value: total,
      loading: isLoading,
    },
    {
      icon: <ClipboardCheckIcon />,
      label: t("admin.attendance.stats.present"),
      value: items.filter((a) => a.status === "present").length,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.attendance.title")} />
      <StatCards stats={statCards} />
      <TableFilter
        table={table}
        searchPlaceholder={t("admin.attendance.searchPlaceholder")}
        sortLabel={t("admin.attendance.toolbar.sort")}
        columnsLabel={t("admin.attendance.toolbar.columns")}
        toggleColumnsLabel={t("admin.attendance.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<ClipboardCheckIcon className="size-8 opacity-40" />}
            emptyTitle={t("admin.attendance.noResults")}
            emptyHint={t("admin.attendance.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <AttendanceFormDialog
        open={!!editTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setEditTarget(null)
        }}
        attendance={editTarget}
      />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget?.user?.name ?? deleteTarget?.id ?? ""}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}

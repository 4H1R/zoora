import type { GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView } from "@/api/model"

import { useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetPractices } from "@/api/practices/practices"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { ClassFilterSelect, SessionFilterSelect } from "@/components/org/class-session-filters"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAdminTable } from "@/lib/data-table"
import { Route } from "@/routes/_auth/org/practices/index"

import { ManagerSubmissionsDialog } from "./ManagerSubmissionsDialog"
import { usePracticeHubColumns } from "./practice-hub-columns"

const WINDOW_OPTIONS = ["all", "upcoming", "open", "ended"] as const

export function ManagerPracticesView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, needs_grading, window: windowState, class_id, class_session_id, order_by, order_dir, page, page_size } =
    Route.useSearch()

  // Default landing filter is "needs grading" — the teacher's actual job queue.
  const needsGrading = needs_grading ?? true
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const [submissionsTarget, setSubmissionsTarget] = useState<PracticeRoomView | null>(null)

  const { data, isLoading } = useGetPractices({
    search: search || undefined,
    needs_grading: needsGrading || undefined,
    window: windowState,
    class_id: class_id || undefined,
    class_session_id: class_session_id || undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
    page: page ?? 1,
    page_size: page_size ?? 20,
  })

  const listData = (data?.status === 200 && data.data.data) || undefined
  const rows = listData?.items ?? []
  const total = listData?.total ?? 0

  const columns = usePracticeHubColumns({ onViewSubmissions: setSubmissionsTarget })
  const table = useAdminTable({ data: rows, columns, rowCount: total, sorting })

  const setNeedsGrading = (value: boolean) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, needs_grading: value || undefined, page: 1 }) })
  // Changing class invalidates any chosen session — sessions belong to one class.
  const setClass = (classId?: string) =>
    navigate({
      to: ".",
      search: (prev) => ({ ...prev, class_id: classId, class_session_id: undefined, page: 1 }),
    })
  const setSession = (sessionId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_session_id: sessionId, page: 1 }) })
  const windowItems = WINDOW_OPTIONS.map((value) => ({
    value,
    label: value === "all" ? t("org.practices.filter.windowAll") : t(`org.practices.filter.window.${value}`),
  }))
  const setWindow = (value: string | null) =>
    navigate({
      to: ".",
      search: (prev) => ({
        ...prev,
        window: value && value !== "all" ? (value as "upcoming" | "open" | "ended") : undefined,
        page: 1,
      }),
    })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.practices.title")} />
        <p className="text-muted-foreground text-sm">{t("org.practices.subtitle")}</p>
      </div>

      <TableFilter
        table={table}
        searchPlaceholder={t("org.practices.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      >
        <Button size="sm" variant={needsGrading ? "default" : "outline"} onClick={() => setNeedsGrading(!needsGrading)}>
          {t("org.practices.filter.needsGrading")}
        </Button>
        <ClassFilterSelect value={class_id} onChange={setClass} />
        <SessionFilterSelect classId={class_id} value={class_session_id} onChange={setSession} />
        <Select items={windowItems} value={windowState ?? "all"} onValueChange={setWindow}>
          <SelectTrigger size="sm" className="w-36">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {windowItems.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </TableFilter>

      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyTitle={t("org.practices.noResults")}
            emptyHint={t("org.practices.noResultsHint")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <ManagerSubmissionsDialog
        open={!!submissionsTarget}
        onOpenChange={(open) => !open && setSubmissionsTarget(null)}
        practice={submissionsTarget}
      />
    </div>
  )
}

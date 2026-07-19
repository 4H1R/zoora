import type { GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView } from "@/api/model"

import { useNavigate } from "@tanstack/react-router"
import { CalendarClockIcon, NotebookPenIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetPractices } from "@/api/practices/practices"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { ClassFilterSelect, SessionFilterSelect } from "@/components/org/class-session-filters"
import { PracticeSubmitDialog } from "@/components/org/practices/PracticeSubmitDialog"
import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useViewMode, ViewModeToggle } from "@/components/view-mode-toggle"
import { useAdminTable } from "@/lib/data-table"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"
import { Route } from "@/routes/_auth/org/practices/index"

import { PracticeStatusBadge } from "./practice-status-badge"
import { usePracticeStudentColumns } from "./practice-student-columns"

const STATUS_FILTERS = ["all", "to_submit", "submitted", "graded", "upcoming", "missed"] as const

export function StudentPracticesView() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const { search, status, class_id, class_session_id, order_by, order_dir, page, page_size } = Route.useSearch()
  const { canSubmit } = usePracticePermissions()

  const activeStatus = status ?? "to_submit"
  const currentPage = page ?? 1
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const { viewMode, setViewMode, isTable } = useViewMode()
  const [submitTarget, setSubmitTarget] = useState<PracticeRoomView | null>(null)

  const { data, isLoading } = useGetPractices({
    search: search || undefined,
    status: activeStatus === "all" ? undefined : activeStatus,
    class_id: class_id || undefined,
    class_session_id: class_session_id || undefined,
    order_by: order_by || undefined,
    order_dir: order_dir || undefined,
    page: currentPage,
    page_size: page_size ?? 20,
  })

  const listData = (data?.status === 200 && data.data.data) || undefined
  const items = listData?.items ?? []
  const total = listData?.total ?? 0

  const columns = usePracticeStudentColumns({ canSubmit, onSubmit: setSubmitTarget })
  const table = useAdminTable({ data: items, columns, rowCount: total, sorting })

  const setStatus = (value: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, status: value === "all" ? undefined : value, page: 1 }) })
  // Changing class invalidates any chosen session — sessions belong to one class.
  const setClass = (classId?: string) =>
    navigate({
      to: ".",
      search: (prev) => ({ ...prev, class_id: classId, class_session_id: undefined, page: 1 }),
    })
  const setSession = (sessionId?: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_session_id: sessionId, page: 1 }) })

  const renderContent = () => {
    if (isTable) {
      return (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="overflow-x-auto">
            <DataTable
              table={table}
              isLoading={isLoading}
              emptyIcon={<NotebookPenIcon className="size-8 opacity-40" />}
              emptyTitle={t("org.practices.noResults")}
              emptyHint={t("org.practices.noResultsHint")}
            />
          </div>
          <DataTablePagination table={table} />
        </Card>
      )
    }

    return (
      <Card className="gap-0 overflow-hidden p-0">
        {isLoading ? (
          <div className="divide-y">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4 px-4 py-4">
                <Skeleton className="size-9 rounded-lg" />
                <div className="flex flex-1 flex-col gap-2">
                  <Skeleton className="h-4 w-48" />
                  <Skeleton className="h-3 w-32" />
                </div>
                <Skeleton className="h-6 w-20 rounded-full" />
              </div>
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center gap-2 px-4 py-16 text-center">
            <NotebookPenIcon className="text-muted-foreground/40 size-8" />
            <p className="text-sm font-medium">{t("org.practices.noResults")}</p>
            <p className="text-muted-foreground text-xs">{t("org.practices.noResultsHint")}</p>
          </div>
        ) : (
          <ul className="divide-y">
            {items.map((practice) => {
              const graded = practice.status === "graded"
              const showSubmit = canSubmit && practice.can_submit
              return (
                <li key={practice.id} className="flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:gap-4">
                  <div className="bg-primary/10 text-primary hidden size-9 shrink-0 items-center justify-center rounded-lg sm:flex">
                    <NotebookPenIcon className="size-4" />
                  </div>
                  <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <span className="truncate font-medium">{practice.title}</span>
                    <div className="text-muted-foreground flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
                      {practice.class?.name && <span className="truncate">{practice.class.name}</span>}
                      <span className="inline-flex items-center gap-1">
                        <CalendarClockIcon className="size-3.5" />
                        {practice.end_time
                          ? `${t("org.practices.due")}: ${formatSessionDate(practice.end_time, i18n.language, "short")}`
                          : t("org.practices.noDueDate")}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    {graded && (
                      <span className="text-sm font-semibold tabular-nums">
                        {formatScore(practice.my_submission?.score)}
                        <span className="text-muted-foreground font-normal">
                          {" / "}
                          {formatScore(practice.max_score ?? 0)}
                        </span>
                      </span>
                    )}
                    <PracticeStatusBadge status={practice.status} />
                    {showSubmit && (
                      <Button size="sm" onClick={() => setSubmitTarget(practice)}>
                        {t("org.practices.actions.submit")}
                      </Button>
                    )}
                  </div>
                </li>
              )
            })}
          </ul>
        )}
        <DataTablePagination table={table} />
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.practices.title")} />
        <p className="text-muted-foreground text-sm">{t("org.practices.subtitle")}</p>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex-1">
          <TableFilter
            table={table}
            searchPlaceholder={t("org.practices.searchPlaceholder")}
            sortLabel={t("common.toolbar.sort")}
            columnsLabel={t("common.toolbar.columns")}
            toggleColumnsLabel={t("common.toolbar.toggleColumns")}
            showColumnsToggle={isTable}
          >
            <div className="flex flex-wrap gap-1.5">
              {STATUS_FILTERS.map((value) => (
                <Button
                  key={value}
                  size="sm"
                  variant={activeStatus === value ? "default" : "outline"}
                  onClick={() => setStatus(value)}
                >
                  {value === "all" ? t("org.practices.filter.all") : t(`org.practices.status.${value}`)}
                </Button>
              ))}
            </div>
            <ClassFilterSelect value={class_id} onChange={setClass} />
            <SessionFilterSelect classId={class_id} value={class_session_id} onChange={setSession} />
          </TableFilter>
        </div>
        <ViewModeToggle value={viewMode} onChange={setViewMode} />
      </div>

      {renderContent()}

      <PracticeSubmitDialog
        open={!!submitTarget}
        onOpenChange={(open) => !open && setSubmitTarget(null)}
        practice={submitTarget}
      />
    </div>
  )
}

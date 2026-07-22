import { Link, useNavigate } from "@tanstack/react-router"
import { CalendarCheckIcon, ExternalLinkIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { AttendanceMatrixView } from "@/components/org/classes/AttendanceMatrixView"
import { useAttendancePermissions } from "@/components/org/livesessions/use-attendance-permissions"
import { ManagerClassPicker, useManagerClasses } from "@/components/org/manager-class-picker"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { Route } from "@/routes/_auth/org/attendance/index"

export function ManagerAttendanceView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { search, class_id, order_by, order_dir, page, page_size } = Route.useSearch()
  const { can } = useAccess()
  const { canEdit } = useAttendancePermissions()

  const { classes, isLoading } = useManagerClasses(can("attendance:view_any"))

  // Fall back to the first class so teachers land on data, not a blank picker.
  const selected = classes.find((cls) => cls.id === class_id) ?? classes[0]

  const setClass = (classId: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_id: classId, page: 1 }) })
  const setPage = (nextPage: number) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, page: nextPage }) })

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col gap-4">
          <Skeleton className="h-8 w-56" />
          <Skeleton className="h-64 w-full rounded-lg" />
        </div>
      )
    }

    if (classes.length === 0) {
      return (
        <EmptyState
          icon={CalendarCheckIcon}
          title={t("org.attendance.manager.noClasses")}
          description={t("org.attendance.manager.noClassesHint")}
        />
      )
    }

    return (
      <>
        <div className="flex flex-wrap items-center gap-2">
          <ManagerClassPicker classes={classes} value={selected?.id} onChange={setClass} />
          {selected?.id && (
            <Button
              variant="ghost"
              size="sm"
              render={<Link to="/org/classes/$classId" params={{ classId: selected.id }} />}
            >
              <ExternalLinkIcon className="size-4" />
              {t("common.openClass")}
            </Button>
          )}
        </div>
        {selected?.id && (
          <AttendanceMatrixView
            key={selected.id}
            classId={selected.id}
            canEdit={canEdit}
            page={page ?? 1}
            pageSize={page_size ?? 20}
            search={search}
            orderBy={order_by}
            orderDir={order_dir}
            onPageChange={setPage}
          />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.class.attendance.title")} />
        <p className="text-muted-foreground text-sm">{t("org.attendance.manager.subtitle")}</p>
      </div>

      {renderContent()}
    </div>
  )
}

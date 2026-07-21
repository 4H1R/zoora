import { Link, useNavigate } from "@tanstack/react-router"
import { ExternalLinkIcon, GraduationCapIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { OrgGradebookView } from "@/components/org/classes/OrgGradebookView"
import { ManagerClassPicker, useManagerClasses } from "@/components/org/manager-class-picker"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { Route } from "@/routes/_auth/org/grades/index"

export function ManagerGradesView() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { class_id } = Route.useSearch()
  const { can } = useAccess()

  const { classes, isLoading } = useManagerClasses(can("gradebook:view_any"))

  // Fall back to the first class so graders land on data, not a blank picker.
  const selected = classes.find((cls) => cls.id === class_id) ?? classes[0]

  const setClass = (classId: string) =>
    navigate({ to: ".", search: (prev) => ({ ...prev, class_id: classId }) })

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col gap-5">
          <Skeleton className="h-8 w-56" />
          <Skeleton className="h-64 w-full rounded-2xl" />
        </div>
      )
    }

    if (classes.length === 0) {
      return (
        <EmptyState
          icon={GraduationCapIcon}
          title={t("org.grades.manager.noClasses")}
          description={t("org.grades.manager.noClassesHint")}
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
        {selected?.id && <OrgGradebookView key={selected.id} classId={selected.id} cls={selected} />}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <PageHeader title={t("org.class.gradebook.title")} />
        <p className="text-muted-foreground text-sm">{t("org.grades.manager.subtitle")}</p>
      </div>

      {renderContent()}
    </div>
  )
}

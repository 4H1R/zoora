import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClassesId } from "@/api/classes/classes"
import { ClassPicker } from "@/components/admin/forms/ClassSessionPicker"
import { GradebookMatrixView } from "@/components/admin/gradebook/GradebookMatrixView"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"

export const Route = createFileRoute("/_admin/admin/classes/$classId/gradebook")({
  head: () => adminHead("admin.classManagement.gradebook"),
  component: GradebookPage,
})

function GradebookPage() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()

  const { data: classData } = useGetClassesId(classId)
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={
          cls?.name
            ? `${cls.name} · ${t("admin.classManagement.gradebook")}`
            : t("admin.classManagement.gradebook")
        }
        actions={
          <Link to="/admin/classes/$classId/sessions" params={{ classId }}>
            <Button variant="outline" size="sm">
              <ArrowLeftIcon data-icon="inline-start" />
              {t("admin.classManagement.backToSessions")}
            </Button>
          </Link>
        }
      />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.gradebook.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={() => {}} disabled />
        </div>
      </Card>
      <GradebookMatrixView classId={classId} />
    </div>
  )
}

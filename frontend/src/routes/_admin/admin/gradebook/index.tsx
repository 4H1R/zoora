import { createFileRoute } from "@tanstack/react-router"
import { TrophyIcon, XIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { ClassPicker } from "@/components/admin/forms/ClassSessionPicker"
import { GradebookMatrixView } from "@/components/admin/gradebook/GradebookMatrixView"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"

export const Route = createFileRoute("/_admin/admin/gradebook/")({
  head: () => adminHead("admin.gradebook.title"),
  component: AdminGradebookPage,
})

function AdminGradebookPage() {
  const { t } = useTranslation()
  const [classId, setClassId] = useState<string | undefined>(undefined)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.gradebook.title")} />
      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-end">
        <div className="flex-1">
          <label className="mb-1.5 block text-xs font-medium">
            {t("admin.gradebook.filter.class")}
          </label>
          <ClassPicker value={classId} onChange={(id) => setClassId(id || undefined)} />
        </div>
        {classId && (
          <Button variant="outline" size="sm" onClick={() => setClassId(undefined)}>
            <XIcon data-icon="inline-start" />
            {t("admin.gradebook.filter.clear")}
          </Button>
        )}
      </Card>
      {classId ? (
        <GradebookMatrixView classId={classId} />
      ) : (
        <Card className="text-muted-foreground flex flex-col items-center gap-3 p-8 text-center text-sm">
          <TrophyIcon className="size-8 opacity-40" />
          {t("admin.gradebook.filter.selectClassFirst")}
        </Card>
      )}
    </div>
  )
}

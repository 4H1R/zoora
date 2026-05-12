import { createFileRoute } from "@tanstack/react-router"
import { TrophyIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"

export const Route = createFileRoute("/_admin/admin/classes/$classId/gradebook")({
  head: () => adminHead("admin.classManagement.gradebook"),
  component: GradebookPage,
})

function GradebookPage() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.classManagement.gradebook")} />
      <Card className="flex flex-col items-center justify-center gap-3 py-16 text-center">
        <TrophyIcon className="text-muted-foreground size-10 opacity-50" />
        <p className="text-muted-foreground text-sm">{t("common.noResults")}</p>
      </Card>
    </div>
  )
}

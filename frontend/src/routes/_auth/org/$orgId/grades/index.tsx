import { createFileRoute } from "@tanstack/react-router"
import { GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetGradebookMe } from "@/api/gradebook/gradebook"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/grades/")({
  head: () => orgHead("org.nav.grades"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const allowed = useOrgGuard(["gradebook:view_own"])

  const gradesQ = useGetGradebookMe({ query: { enabled: allowed } })
  const gradebook = (gradesQ.data?.status === 200 && gradesQ.data.data.data) || undefined
  const classes = gradebook?.classes ?? []
  const loading = gradesQ.isPending

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.grades.title")} />

      {loading ? (
        <div className="flex flex-col gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <Card key={i} className="gap-0 overflow-hidden p-0">
              <div className="border-b px-4 py-3">
                <Skeleton className="h-4 w-40" />
              </div>
              <div className="flex flex-col gap-3 p-4">
                {Array.from({ length: 3 }).map((__, j) => (
                  <div key={j} className="flex items-center justify-between">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-4 w-12" />
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </div>
      ) : classes.length === 0 ? (
        <Card className="flex flex-col items-center gap-2 px-6 py-12 text-center">
          <div className="bg-muted text-muted-foreground mb-1 flex size-12 items-center justify-center rounded-xl [&>svg]:size-6">
            <GraduationCapIcon />
          </div>
          <p className="text-muted-foreground max-w-sm text-sm">{t("org.grades.empty")}</p>
        </Card>
      ) : (
        <div className="flex flex-col gap-4">
          {classes.map((cls) => {
            const columns = cls.columns ?? []
            return (
              <Card key={cls.class_id} className="gap-0 overflow-hidden p-0">
                <div className="flex items-center gap-2 border-b px-4 py-3">
                  <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
                    <GraduationCapIcon />
                  </div>
                  <h2 className="truncate text-sm font-semibold">{cls.class_name || "—"}</h2>
                </div>
                {columns.length === 0 ? (
                  <p className="text-muted-foreground px-4 py-8 text-center text-sm">
                    {t("org.grades.noColumns")}
                  </p>
                ) : (
                  <table className="w-full text-sm">
                    <tbody className="divide-y">
                      {columns.map((col) => {
                        const value = col.id ? cls.cells?.[col.id] : undefined
                        return (
                          <tr key={col.id}>
                            <td className="text-muted-foreground px-4 py-2.5 text-start">{col.title || "—"}</td>
                            <td className="px-4 py-2.5 text-end font-medium tabular-nums">
                              {value && value.trim() ? value : "—"}
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </table>
                )}
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}

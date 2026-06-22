import type { GithubCom4H1RZooraInternalDomainAttendanceStatus as AttendanceStatus } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import {
  CalendarCheckIcon,
  CalendarXIcon,
  ClockIcon,
  ShieldCheckIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetAttendanceMe } from "@/api/attendance/attendance"
import { StatCards, type StatItem } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/$orgId/attendance/")({
  head: () => orgHead("org.nav.attendance"),
  component: RouteComponent,
})

function statusBadgeVariant(status: AttendanceStatus | undefined) {
  switch (status) {
    case "present":
      return "default" as const
    case "absent":
      return "destructive" as const
    case "late":
      return "outline" as const
    case "excused":
      return "secondary" as const
    default:
      return "ghost" as const
  }
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const allowed = useOrgGuard(["attendance:view"])

  const attendanceQ = useGetAttendanceMe(undefined, { query: { enabled: allowed } })
  const attendance = (attendanceQ.data?.status === 200 && attendanceQ.data.data.data) || undefined
  const summary = attendance?.summary
  const items = attendance?.items ?? []
  const loading = attendanceQ.isPending

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

      {loading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <Card key={i} size="sm" className="flex-row items-center gap-3 p-4">
              <div className="flex flex-1 flex-col gap-2">
                <Skeleton className="h-4 w-40" />
                <Skeleton className="h-3 w-28" />
              </div>
              <Skeleton className="h-5 w-16" />
            </Card>
          ))}
        </div>
      ) : items.length === 0 ? (
        <Card className="flex flex-col items-center gap-2 px-6 py-12 text-center">
          <div className="bg-muted text-muted-foreground mb-1 flex size-12 items-center justify-center rounded-xl [&>svg]:size-6">
            <CalendarCheckIcon />
          </div>
          <p className="text-muted-foreground max-w-sm text-sm">{t("org.attendance.empty")}</p>
        </Card>
      ) : (
        <div className="flex flex-col gap-3">
          {items.map((item) => (
            <Card key={item.id} size="sm" className="flex-row items-center gap-3 p-4">
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{item.class?.name || "—"}</p>
                <p className="text-muted-foreground mt-0.5 truncate text-xs">
                  {item.class_session?.name || item.class_session?.description || ""}
                  {item.created_at ? (item.class_session?.name ? " · " : "") + formatSessionDate(item.created_at, i18n.language, "short") : ""}
                </p>
              </div>
              <Badge variant={statusBadgeVariant(item.status)}>{t(`org.attendance.status.${item.status}`)}</Badge>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}

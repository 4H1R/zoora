import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, CalendarClockIcon, SparklesIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClassesId, useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { AttendanceSection } from "@/components/org/livesessions/AttendanceSection"
import { useAttendancePermissions } from "@/components/org/livesessions/use-attendance-permissions"
import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate, useSessionStatus } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/classes/class-sessions/$classSessionId_/attendance")({
  head: () => orgHead("org.session.attendance.title"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { classSessionId } = Route.useParams()
  const allowed = useOrgGuard(["attendance:view", "attendance:view_any", "attendance:create"])
  const { canView } = useAttendancePermissions()

  const {
    data: sessionData,
    isPending: sessionPending,
    isError: sessionError,
  } = useGetClassesSessionsSessionId(classSessionId)
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined
  const classId = session?.class_id
  const status = useSessionStatus(session?.start_time)

  const { data: classData } = useGetClassesId(classId ?? "", { query: { enabled: !!classId } })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  useBreadcrumb([
    { label: t("org.nav.classes"), to: "/org/classes" },
    {
      label: cls?.name ?? null,
      to: "/org/classes/$classId",
      params: { classId: classId ?? "" },
      loading: !cls,
    },
    {
      label: session?.name ?? null,
      to: "/org/classes/class-sessions/$classSessionId",
      params: { classSessionId },
      loading: !session,
    },
    { label: t("org.session.attendance.title") },
  ])

  if (!allowed || !canView) return null

  if (sessionPending) {
    return (
      <div className="flex flex-col gap-6 py-6">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-24 w-full rounded-2xl" />
        <div className="flex flex-col gap-2">
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
        </div>
      </div>
    )
  }

  if (sessionError || !session || !classId) {
    return (
      <div className="flex flex-col items-start gap-4 py-16">
        <h1 className="text-2xl font-semibold tracking-tight">{t("org.session.notFound.title")}</h1>
        <Button variant="outline" render={<Link to="/org/classes" />}>
          <ArrowLeftIcon className="size-4 rtl:rotate-180" />
          {t("org.session.notFound.backToClasses")}
        </Button>
      </div>
    )
  }

  const startStr = formatSessionDate(session.start_time, i18n.language, "long")

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/6%,transparent_55%)]"
      />

      <div className="pt-6">
        <Link
          to="/org/classes/class-sessions/$classSessionId"
          params={{ classSessionId }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
          {t("org.session.attendancePage.back")}
        </Link>
      </div>

      <header className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2.5">
          <Eyebrow>{t("org.session.attendance.title")}</Eyebrow>
          <SessionStatusPill status={status} size="sm" />
        </div>
        <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{session.name}</h1>
        <div className="text-muted-foreground flex flex-wrap items-center gap-x-2 gap-y-1.5 text-sm">
          <span className="inline-flex items-center gap-1.5">
            <CalendarClockIcon className="size-3.5 opacity-70" />
            {startStr}
          </span>
          {cls?.name && (
            <>
              <span className="text-muted-foreground/40">·</span>
              <span className="inline-flex items-center gap-1.5">
                <SparklesIcon className="size-3.5 opacity-70" />
                {cls.name}
              </span>
            </>
          )}
        </div>
      </header>

      <AttendanceSection classId={classId} classSessionId={classSessionId} />
    </div>
  )
}

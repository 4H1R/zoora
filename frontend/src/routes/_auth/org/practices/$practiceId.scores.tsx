import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, ClockIcon, ExternalLinkIcon, TrophyIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetPracticesId } from "@/api/practices/practices"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { PracticeGradingPanel } from "@/components/org/practices/PracticeGradingPanel"
import { usePracticePermissions } from "@/components/org/practices/use-practice-permissions"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/practices/$practiceId/scores")({
  head: () => orgHead("org.practices.scores.title"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { practiceId } = Route.useParams()
  const allowed = useOrgGuard(["practices:view", "practices:view_any"])
  const { canGrade } = usePracticePermissions()

  const { data, isPending, isError } = useGetPracticesId(practiceId)
  const practice = (data?.status === 200 && data.data.data) || undefined
  // The practice knows its own session, so the back path needs no URL params —
  // a shared link keeps its context.
  const sessionId = practice?.class_session_id

  useBreadcrumb([
    { label: t("org.nav.practices"), to: "/org/practices" },
    { label: practice?.title ?? null, loading: !practice },
  ])

  if (!allowed) return null

  if (isPending) {
    return (
      <div className="flex flex-col gap-8 py-6">
        <Skeleton className="h-4 w-40" />
        <div className="flex flex-col gap-4">
          <Skeleton className="h-3 w-24" />
          <Skeleton className="h-10 w-2/3" />
          <Skeleton className="h-5 w-96" />
        </div>
        <Skeleton className="h-32 w-full rounded-2xl" />
        <div className="flex flex-col gap-3">
          <Skeleton className="h-24 w-full rounded-2xl" />
          <Skeleton className="h-24 w-full rounded-2xl" />
          <Skeleton className="h-24 w-full rounded-2xl" />
        </div>
      </div>
    )
  }

  if (isError || !practice || !canGrade) {
    return (
      <div className="flex flex-col items-start gap-4 py-16">
        <h1 className="text-2xl font-semibold tracking-tight">{t("org.practices.scores.notFound")}</h1>
        <Button variant="outline" render={<Link to="/org/practices" />}>
          <ArrowLeftIcon className="size-4 rtl:rotate-180" />
          {t("org.practices.scores.back")}
        </Button>
      </div>
    )
  }

  const startStr = formatSessionDate(practice.start_time, i18n.language, "short")
  const endStr = formatSessionDate(practice.end_time, i18n.language, "short")

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(ellipse_at_bottom_left,var(--color-primary)/8%,transparent_55%)]"
      />

      <div className="pt-6">
        {sessionId ? (
          <Link
            to="/org/classes/class-sessions/$classSessionId"
            params={{ classSessionId: sessionId }}
            search={{ tab: "practices" }}
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
          >
            <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
            {t("org.practices.scores.backToSession")}
          </Link>
        ) : (
          <Link
            to="/org/practices"
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
          >
            <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
            {t("org.practices.scores.back")}
          </Link>
        )}
      </div>

      <header className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-3">
          <Eyebrow>{t("org.practices.scores.title")}</Eyebrow>
          <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{practice.title || "—"}</h1>
          <div className="text-muted-foreground flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
            <span className="inline-flex items-center gap-1.5 tabular-nums">
              <ClockIcon className="size-4" />
              {startStr} → {endStr}
            </span>
            {typeof practice.max_score === "number" && (
              <span className="inline-flex items-center gap-1.5 tabular-nums">
                <TrophyIcon className="size-4" />
                {formatScore(practice.max_score)}
              </span>
            )}
            {sessionId && (
              <Link
                to="/org/classes/class-sessions/$classSessionId"
                params={{ classSessionId: sessionId }}
                search={{ tab: "practices" }}
                className="hover:text-foreground inline-flex items-center gap-1.5 underline-offset-4 transition-colors hover:underline"
              >
                <ExternalLinkIcon className="size-3.5" />
                {t("org.practices.scores.openSession")}
              </Link>
            )}
          </div>
        </div>
        <span className="text-muted-foreground hidden font-mono text-[11px] tracking-[0.25em] uppercase md:inline">
          {`// ${practice.title ?? ""}`}
        </span>
      </header>

      <PracticeGradingPanel practice={practice} />
    </div>
  )
}

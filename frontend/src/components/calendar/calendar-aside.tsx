import { useTranslation } from "react-i18next"
import { useAccess } from "react-access-engine"
import { Link } from "@tanstack/react-router"
import { ChevronLeft, ChevronRight } from "lucide-react"

import { eventDotColor } from "@/lib/calendar"
import { ORG_ROUTES, type OrgRouteKey } from "@/lib/org-routes"
import { Card } from "@/components/ui/card"
import { cn } from "@/lib/utils"

const LEGEND: { type: string; key: string }[] = [
  { type: "live", key: "org.calendar.legend.live" },
  { type: "quiz", key: "org.calendar.legend.quiz" },
  { type: "practice", key: "org.calendar.legend.practice" },
  { type: "offline", key: "org.calendar.legend.offline" },
]

// Learning shortcuts shown in the aside — Online Classes first, mirroring the
// sidebar's "Learning" nav group.
const LEARNING_KEYS: OrgRouteKey[] = [
  "online-classes",
  "exams",
  "practices",
  "grades",
  "attendance",
]

// CalendarLegend maps each event-type color to its label and shows the total
// event count for the visible period.
export function CalendarLegend({ monthCount }: { monthCount: number }) {
  const { t } = useTranslation()
  return (
    <Card className="gap-3 p-4">
      <div className="text-muted-foreground flex items-center justify-between text-xs font-semibold">
        <span>{t("org.calendar.legendTitle")}</span>
        <span className="tabular-nums">
          {t("org.calendar.monthCount", { count: monthCount })}
        </span>
      </div>
      <div className="grid grid-cols-2 gap-x-4 gap-y-2">
        {LEGEND.map((l) => (
          <span
            key={l.type}
            className="text-foreground flex items-center gap-2 text-sm"
          >
            <span
              className={cn("h-2.5 w-2.5 rounded-full", eventDotColor(l.type))}
            />
            {t(l.key)}
          </span>
        ))}
      </div>
    </Card>
  )
}

// LearningLinks renders perm-gated quick links to the org's learning routes.
// Hidden entirely when the user can access none of them.
export function LearningLinks({ orgId }: { orgId: string }) {
  const { t, i18n } = useTranslation()
  const { can } = useAccess()
  const routes = LEARNING_KEYS.map((k) => ORG_ROUTES[k]).filter(
    (s) => !s.perms || s.perms.some((p) => can(p))
  )
  if (routes.length === 0) return null
  const Chevron = i18n.language === "fa" ? ChevronLeft : ChevronRight

  return (
    <Card className="gap-1 p-2">
      <p className="text-muted-foreground px-2 pt-1.5 pb-1 text-xs font-semibold">
        {t("org.nav.learning")}
      </p>
      <nav className="flex flex-col">
        {routes.map((spec) => (
          <Link
            key={spec.segment}
            to={`/org/${orgId}/${spec.segment}` as string}
            className="group hover:bg-accent flex items-center gap-2.5 rounded-lg px-2 py-2 text-sm font-medium transition-colors [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-muted-foreground group-hover:[&_svg]:text-primary"
          >
            {spec.icon}
            <span className="flex-1 truncate">{t(spec.i18nKey)}</span>
            <Chevron className="text-muted-foreground/40 group-hover:text-foreground size-3.5 transition-all group-hover:-translate-x-0.5 rtl:group-hover:translate-x-0.5" />
          </Link>
        ))}
      </nav>
    </Card>
  )
}

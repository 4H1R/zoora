import { useState } from "react"
import { useTranslation } from "react-i18next"
import { createFileRoute, Link } from "@tanstack/react-router"
import {
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Clapperboard,
  GraduationCap,
  NotebookPen,
  Video,
  type LucideIcon,
} from "lucide-react"

import { useGetCalendarEvents } from "@/api/calendar/calendar"
import {
  bucketByDay,
  dateKey,
  eventAccent,
  eventDotColor,
  eventLink,
  eventTime,
  formatDayParts,
  getMonthRange,
  isToday,
  type CalendarEvent,
} from "@/lib/calendar"
import { orgHead } from "@/lib/org-head"
import { PageHeader } from "@/components/page-header"
import { CalendarBoard } from "@/components/calendar/calendar-board"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/calendar")({
  head: () => orgHead("org.nav.calendar"),
  component: RouteComponent,
})

const LEGEND: { type: string; key: string }[] = [
  { type: "live", key: "org.calendar.legend.live" },
  { type: "quiz", key: "org.calendar.legend.quiz" },
  { type: "practice", key: "org.calendar.legend.practice" },
  { type: "offline", key: "org.calendar.legend.offline" },
]

const EVENT_ICON: Record<string, LucideIcon> = {
  live: Video,
  quiz: GraduationCap,
  practice: NotebookPen,
  offline: Clapperboard,
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { orgId } = Route.useParams()
  const lang = i18n.language

  const [month, setMonth] = useState<Date>(() => new Date())
  const [selected, setSelected] = useState<Date | undefined>(() => new Date())

  const range = getMonthRange(month, lang)
  const eventsQ = useGetCalendarEvents({ from: range.from, to: range.to })
  const events: CalendarEvent[] =
    (eventsQ.data?.status === 200 && eventsQ.data.data.data?.events) || []
  const loading = eventsQ.isPending

  const buckets = bucketByDay(events)
  const selectedKey = selected ? dateKey(selected) : ""
  const dayEvents = buckets.get(selectedKey) ?? []
  const parts = selected ? formatDayParts(selected, lang) : null
  const selectedIsToday = selected ? isToday(selected) : false

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("org.calendar.title")} />

      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        {LEGEND.map((l) => (
          <span
            key={l.type}
            className="text-muted-foreground flex items-center gap-1.5 text-sm"
          >
            <span className={cn("h-2 w-2 rounded-full", eventDotColor(l.type))} />
            {t(l.key)}
          </span>
        ))}
      </div>

      <div className="grid items-start gap-6 lg:grid-cols-[auto_1fr]">
        {loading ? (
          <Skeleton className="h-80 w-72" />
        ) : (
          <CalendarBoard
            events={events}
            month={month}
            onMonthChange={setMonth}
            selected={selected}
            onSelect={setSelected}
          />
        )}

        <Card className="overflow-hidden p-0">
          {/* Day header — what day am I looking at (weekday + long date). */}
          <div className="bg-muted/40 flex items-start justify-between gap-3 border-b p-4">
            <div className="flex flex-col gap-0.5">
              <div className="flex items-center gap-2">
                <span className="text-xl leading-none font-semibold tracking-tight">
                  {parts?.weekday ?? t("org.calendar.selectedDay")}
                </span>
                {selectedIsToday && (
                  <span className="bg-primary/10 text-primary rounded-full px-2 py-0.5 text-xs font-medium">
                    {t("org.calendar.todayBadge")}
                  </span>
                )}
              </div>
              {parts && (
                <span className="text-muted-foreground text-sm">{parts.date}</span>
              )}
            </div>
            <span className="bg-background text-muted-foreground flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium">
              <CalendarDays className="h-3.5 w-3.5" />
              {t("org.calendar.eventCount", { count: dayEvents.length })}
            </span>
          </div>

          {dayEvents.length === 0 ? (
            <div className="text-muted-foreground flex flex-col items-center gap-2 px-4 py-12 text-center">
              <CalendarDays className="h-8 w-8 opacity-40" />
              <p className="text-sm">{t("org.calendar.empty")}</p>
            </div>
          ) : (
            <ul className="flex flex-col gap-2 p-3">
              {dayEvents.map((e) => {
                const link = eventLink(orgId, e)
                const accent = eventAccent(e.type)
                const Icon = EVENT_ICON[e.type ?? ""] ?? CalendarDays
                const time = eventTime(e.start_time, lang)
                const Chevron = lang === "fa" ? ChevronLeft : ChevronRight
                return (
                  <li key={e.id}>
                    <Link
                      to={link.to}
                      params={link.params}
                      className="group hover:bg-accent flex items-stretch gap-3 rounded-lg border p-2.5 transition-colors"
                    >
                      <span
                        className={cn("w-1 shrink-0 rounded-full", accent.bar)}
                      />
                      <span
                        className={cn(
                          "flex h-9 w-9 shrink-0 items-center justify-center rounded-md",
                          accent.chipBg,
                          accent.chipText
                        )}
                      >
                        <Icon className="h-4 w-4" />
                      </span>
                      <span className="flex min-w-0 flex-1 flex-col justify-center">
                        <span className="truncate font-medium">{e.title}</span>
                        <span className="text-muted-foreground truncate text-xs">
                          {e.class_name}
                        </span>
                      </span>
                      <span className="flex shrink-0 items-center gap-1">
                        {time && (
                          <span className="text-muted-foreground text-xs tabular-nums">
                            {time}
                          </span>
                        )}
                        <Chevron className="text-muted-foreground/60 group-hover:text-foreground h-4 w-4 transition-colors" />
                      </span>
                    </Link>
                  </li>
                )
              })}
            </ul>
          )}
        </Card>
      </div>
    </div>
  )
}

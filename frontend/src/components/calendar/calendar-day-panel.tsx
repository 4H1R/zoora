import type { CalendarEvent } from "@/lib/calendar"
import type { LucideIcon } from "lucide-react"

import { Link } from "@tanstack/react-router"
import {
  CalendarDays,
  Check,
  ChevronLeft,
  ChevronRight,
  Clapperboard,
  GraduationCap,
  NotebookPen,
  Video,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { Card } from "@/components/ui/card"
import { eventAccent, eventLink, eventTime } from "@/lib/calendar"
import { cn } from "@/lib/utils"

const EVENT_ICON: Record<string, LucideIcon> = {
  live: Video,
  quiz: GraduationCap,
  practice: NotebookPen,
  offline: Clapperboard,
}

type DayParts = { weekday: string; date: string }

type CalendarDayPanelProps = {
  events: CalendarEvent[]
  parts: DayParts | null
  isToday: boolean
}

// CalendarDayPanel is the left "what's on this day" column: a hero header with
// the selected day + event count, then a vertical timeline of that day's events
// (or an empty state). Selection state lives in the parent route.
export function CalendarDayPanel({ events, parts, isToday }: CalendarDayPanelProps) {
  return (
    <Card className="overflow-hidden p-0">
      <DayHeader parts={parts} isToday={isToday} count={events.length} />
      {events.length === 0 ? (
        <EmptyDay />
      ) : (
        <ul className="flex flex-col p-3">
          {events.map((e, i) => (
            <TimelineEvent key={e.id} event={e} isFirst={i === 0} isLast={i === events.length - 1} />
          ))}
        </ul>
      )}
    </Card>
  )
}

function DayHeader({ parts, isToday, count }: { parts: DayParts | null; isToday: boolean; count: number }) {
  const { t } = useTranslation()
  return (
    <div className="relative overflow-hidden border-b">
      <div
        aria-hidden
        className="from-primary/8 pointer-events-none absolute inset-0 bg-gradient-to-bl to-transparent"
      />
      <div className="relative flex flex-wrap items-start justify-between gap-3 p-5">
        <div className="flex min-w-0 flex-col gap-1">
          <div className="flex items-center gap-2.5">
            <span className="text-2xl leading-none font-bold tracking-tight">
              {parts?.weekday ?? t("org.calendar.selectedDay")}
            </span>
            {isToday && (
              <span className="bg-primary/12 text-primary inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-semibold">
                <span className="bg-primary h-1.5 w-1.5 rounded-full" />
                {t("org.calendar.todayBadge")}
              </span>
            )}
          </div>
          {parts && <span className="text-muted-foreground text-sm">{parts.date}</span>}
        </div>
        <span className="bg-background/80 text-foreground flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-semibold tabular-nums backdrop-blur">
          <CalendarDays className="text-muted-foreground h-3.5 w-3.5" />
          {t("org.calendar.eventCount", { count })}
        </span>
      </div>
    </div>
  )
}

function EmptyDay() {
  const { t } = useTranslation()
  return (
    <div className="text-muted-foreground flex flex-col items-center gap-3 px-4 py-16 text-center">
      <span className="bg-muted/60 flex h-14 w-14 items-center justify-center rounded-2xl">
        <CalendarDays className="h-6 w-6 opacity-50" />
      </span>
      <div className="flex flex-col gap-1">
        <p className="text-foreground text-sm font-medium">{t("org.calendar.empty")}</p>
        <p className="text-xs">{t("org.calendar.emptyHint")}</p>
      </div>
    </div>
  )
}

function TimelineEvent({ event, isFirst, isLast }: { event: CalendarEvent; isFirst: boolean; isLast: boolean }) {
  const { i18n } = useTranslation()
  const lang = i18n.language
  const link = eventLink(event)
  const accent = eventAccent(event.type)
  const Icon = EVENT_ICON[event.type ?? ""] ?? CalendarDays
  const time = eventTime(event.start_time, lang)
  const Chevron = lang === "fa" ? ChevronLeft : ChevronRight
  // Events whose start time is in the past get a "done" check on their node.
  const passed = event.start_time ? new Date(event.start_time) < new Date() : false

  return (
    <li className="grid grid-cols-[3.25rem_1.25rem_1fr] items-stretch gap-x-1">
      <span className="text-muted-foreground pt-4 text-end text-xs font-medium tabular-nums">{time}</span>
      <span aria-hidden className="relative flex justify-center">
        <span
          className={cn(
            "bg-border absolute left-1/2 w-px -translate-x-1/2",
            isFirst ? "top-4" : "top-0",
            isLast ? "bottom-4" : "bottom-0"
          )}
        />
        {passed ? (
          <span
            className={cn(
              "ring-card relative z-10 mt-3 flex h-5 w-5 items-center justify-center rounded-full text-white ring-4",
              accent.bar
            )}
          >
            <Check className="h-3 w-3" strokeWidth={3} />
          </span>
        ) : (
          <span className={cn("ring-card relative z-10 mt-3.5 h-2.5 w-2.5 rounded-full ring-4", accent.bar)} />
        )}
      </span>
      <Link
        to={link.to}
        params={link.params}
        className="group hover:border-border hover:bg-card my-1.5 flex items-stretch gap-3 rounded-xl border border-transparent p-2.5 transition-all"
      >
        <span
          className={cn(
            "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-transform group-hover:scale-105",
            accent.chipBg,
            accent.chipText
          )}
        >
          <Icon className="h-4.5 w-4.5" />
        </span>
        <span className="flex min-w-0 flex-1 flex-col justify-center">
          <span className="group-hover:text-primary truncate font-semibold transition-colors">{event.title}</span>
          <span className="text-muted-foreground truncate text-xs">{event.class_name}</span>
        </span>
        <Chevron className="text-muted-foreground/40 group-hover:text-primary mt-1 h-4 w-4 shrink-0 self-center transition-all group-hover:-translate-x-0.5 rtl:group-hover:translate-x-0.5" />
      </Link>
    </li>
  )
}

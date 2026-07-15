import type { CalendarEvent } from "@/lib/calendar"

import { createFileRoute } from "@tanstack/react-router"
import { Sparkles } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetCalendarEvents } from "@/api/calendar/calendar"
import { CalendarLegend, LearningLinks } from "@/components/calendar/calendar-aside"
import { CalendarBoard } from "@/components/calendar/calendar-board"
import { CalendarDayPanel } from "@/components/calendar/calendar-day-panel"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { bucketByDay, dateKey, formatDayParts, getMonthRange, isToday } from "@/lib/calendar"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/calendar")({
  head: () => orgHead("org.nav.calendar"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const lang = i18n.language

  const [month, setMonth] = useState<Date>(() => new Date())
  const [selected, setSelected] = useState<Date | undefined>(() => new Date())

  const range = getMonthRange(month, lang)
  const eventsQ = useGetCalendarEvents({ from: range.from, to: range.to })
  const events: CalendarEvent[] = (eventsQ.data?.status === 200 && eventsQ.data.data.data?.events) || []
  const loading = eventsQ.isPending

  const buckets = bucketByDay(events)
  const dayEvents = selected ? (buckets.get(dateKey(selected)) ?? []) : []
  const parts = selected ? formatDayParts(selected, lang) : null

  function jumpToToday() {
    const now = new Date()
    setMonth(now)
    setSelected(now)
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("org.calendar.title")}
        actions={
          <Button variant="outline" size="sm" onClick={jumpToToday} className="gap-1.5">
            <Sparkles className="h-3.5 w-3.5" />
            {t("org.calendar.jumpToday")}
          </Button>
        }
      />
      <p className="text-muted-foreground -mt-4 text-sm">{t("org.calendar.subtitle")}</p>

      <div className="grid items-start gap-6 lg:grid-cols-[1fr_19rem]">
        <CalendarDayPanel events={dayEvents} parts={parts} isToday={selected ? isToday(selected) : false} />

        <aside className="flex flex-col gap-4 lg:sticky lg:top-6">
          {loading ? (
            <Skeleton className="h-80 w-full" />
          ) : (
            <CalendarBoard
              events={events}
              month={month}
              onMonthChange={setMonth}
              selected={selected}
              onSelect={setSelected}
            />
          )}

          <CalendarLegend monthCount={events.length} />
          <LearningLinks />
        </aside>
      </div>
    </div>
  )
}

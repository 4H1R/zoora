import { addDays, endOfMonth, format, isSameDay, startOfMonth } from "date-fns"
import {
  endOfMonth as jEndOfMonth,
  format as jFormat,
  startOfMonth as jStartOfMonth,
} from "date-fns-jalali"
import { faIR } from "date-fns-jalali/locale"

import type { GithubCom4H1RZooraInternalDomainCalendarEvent as CalendarEvent } from "@/api/model"

export type { CalendarEvent }

// dateKey returns a stable per-day key using LOCAL date parts (not toISOString)
// so events bucket into the day the user actually sees in their timezone.
export function dateKey(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, "0")
  const day = String(d.getDate()).padStart(2, "0")
  return `${y}-${m}-${day}`
}

// getMonthRange returns the UTC [from, to] ISO window covering the visible
// month grid, padded ±7 days for the overflow weeks. Bounds are computed in
// the active calendar system so the Jalali grid is fully covered.
export function getMonthRange(
  month: Date,
  lang: string
): { from: string; to: string } {
  const isFa = lang === "fa"
  const start = isFa ? jStartOfMonth(month) : startOfMonth(month)
  const end = isFa ? jEndOfMonth(month) : endOfMonth(month)
  return {
    from: addDays(start, -7).toISOString(),
    to: addDays(end, 7).toISOString(),
  }
}

// formatDayParts splits a date into its weekday name and long date in the
// active calendar system (Jalali for fa, Gregorian otherwise).
export function formatDayParts(
  d: Date,
  lang: string
): { weekday: string; date: string } {
  if (lang === "fa") {
    return {
      weekday: jFormat(d, "EEEE", { locale: faIR }),
      date: jFormat(d, "d MMMM yyyy", { locale: faIR }),
    }
  }
  return {
    weekday: format(d, "EEEE"),
    date: format(d, "d MMMM yyyy"),
  }
}

// eventTime renders an event's start time as localized HH:mm (Persian digits
// for fa). Returns "" when the event has no start_time.
export function eventTime(iso: string | undefined, lang: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  return lang === "fa"
    ? jFormat(d, "HH:mm", { locale: faIR })
    : format(d, "HH:mm")
}

// isToday reports whether d falls on the current local day. Calendar-system
// agnostic — same instant means same day regardless of fa/en formatting.
export function isToday(d: Date): boolean {
  return isSameDay(d, new Date())
}

export type EventAccent = { bar: string; chipBg: string; chipText: string }

export function eventAccent(type: string | undefined): EventAccent {
  switch (type) {
    case "live":
      return {
        bar: "bg-green-500",
        chipBg: "bg-green-500/10",
        chipText: "text-green-600 dark:text-green-400",
      }
    case "quiz":
      return {
        bar: "bg-amber-500",
        chipBg: "bg-amber-500/10",
        chipText: "text-amber-600 dark:text-amber-400",
      }
    case "practice":
      return {
        bar: "bg-blue-500",
        chipBg: "bg-blue-500/10",
        chipText: "text-blue-600 dark:text-blue-400",
      }
    case "offline":
    default:
      return {
        bar: "bg-gray-400",
        chipBg: "bg-gray-400/10",
        chipText: "text-gray-600 dark:text-gray-300",
      }
  }
}

// bucketByDay groups events by local day key. Each value preserves order.
export function bucketByDay(
  events: CalendarEvent[]
): Map<string, CalendarEvent[]> {
  const map = new Map<string, CalendarEvent[]>()
  for (const e of events) {
    if (!e.start_time) continue
    const key = dateKey(new Date(e.start_time))
    const bucket = map.get(key)
    if (bucket) bucket.push(e)
    else map.set(key, [e])
  }
  return map
}

export function eventDotColor(type: string | undefined): string {
  switch (type) {
    case "live":
      return "bg-green-500"
    case "quiz":
      return "bg-amber-500"
    case "practice":
      return "bg-blue-500"
    case "offline":
      return "bg-gray-400"
    default:
      return "bg-gray-400"
  }
}

export function eventLink(
  e: CalendarEvent
): { to: string; params: Record<string, string> } {
  const entity = e.entity_id ?? ""
  switch (e.type) {
    case "live":
      return { to: "/live/$liveId", params: { liveId: entity } }
    case "quiz":
      return {
        to: "/org/exams/$quizId/take",
        params: { quizId: entity },
      }
    case "offline":
      return {
        to: "/org/offlines/$offlineId",
        params: { offlineId: entity },
      }
    case "practice":
    default:
      return {
        to: "/org/classes/class-sessions/$classSessionId",
        params: { classSessionId: entity },
      }
  }
}

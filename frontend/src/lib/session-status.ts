import { useEffect, useState } from "react"

import { formatDate } from "./format-date"

export type SessionStatus = "scheduled" | "live" | "ended"

export const LIVE_WINDOW_MS = 1000 * 60 * 60 * 2

export function useNow(intervalMs = 1000) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
  return now
}

// setTimeout caps at a signed 32-bit ms (~24.8 days); a larger delay overflows
// to fire immediately, so we clamp and let the hook re-evaluate at that ceiling.
const MAX_TIMEOUT_MS = 2_147_483_647

// Live status only flips at two instants: start (scheduled→live) and
// start+window (live→ended). Polling every second to detect those is wasteful —
// instead schedule a single timeout to the exact next boundary, so consumers
// re-render ~twice per session rather than 60×/min.
export function useSessionStatus(startIso: string | undefined): SessionStatus {
  const [status, setStatus] = useState<SessionStatus>(() => getSessionStatus(startIso, Date.now()))
  useEffect(() => {
    let id: ReturnType<typeof setTimeout>
    const schedule = () => {
      const now = Date.now()
      const next = getSessionStatus(startIso, now)
      setStatus(next)
      if (next === "ended" || !startIso) return // terminal — no further transitions
      const start = new Date(startIso).getTime()
      if (Number.isNaN(start)) return
      const boundary = next === "scheduled" ? start : start + LIVE_WINDOW_MS
      const delay = Math.min(Math.max(0, boundary - now) + 50, MAX_TIMEOUT_MS)
      id = setTimeout(schedule, delay)
    }
    schedule()
    return () => clearTimeout(id)
  }, [startIso])
  return status
}

export function getSessionStatus(startIso: string | undefined, now: number): SessionStatus {
  if (!startIso) return "scheduled"
  const start = new Date(startIso).getTime()
  if (Number.isNaN(start)) return "scheduled"
  if (now < start) return "scheduled"
  if (now < start + LIVE_WINDOW_MS) return "live"
  return "ended"
}

// Thin delegate to the date-formatting source of truth (./format-date).
export function formatSessionDate(
  iso: string | undefined,
  locale: string,
  variant: "short" | "long" = "short"
): string {
  return formatDate(iso, locale, variant === "long" ? "datetime-long" : "datetime")
}

// Locale-aware relative time ("in 2 days" / "۲ روز دیگر"). Intl picks the unit
// granularity and handles fa/en wording + numerals. Empty string for no/bad date.
export function formatRelativeTime(targetIso: string | undefined, now: number, locale: string): string {
  if (!targetIso) return ""
  const target = new Date(targetIso).getTime()
  if (Number.isNaN(target)) return ""
  const diff = target - now
  const abs = Math.abs(diff)
  const min = 60_000
  const hr = 3_600_000
  const day = 86_400_000
  const week = day * 7
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" })
  if (abs < min) return rtf.format(0, "second")
  if (abs < hr) return rtf.format(Math.round(diff / min), "minute")
  if (abs < day) return rtf.format(Math.round(diff / hr), "hour")
  if (abs < week) return rtf.format(Math.round(diff / day), "day")
  if (abs < day * 30) return rtf.format(Math.round(diff / week), "week")
  if (abs < day * 365) return rtf.format(Math.round(diff / (day * 30)), "month")
  return rtf.format(Math.round(diff / (day * 365)), "year")
}

export function formatCountdown(targetIso: string | undefined, now: number): string {
  if (!targetIso) return "—"
  const target = new Date(targetIso).getTime()
  if (Number.isNaN(target)) return "—"
  const abs = Math.abs(target - now)
  const days = Math.floor(abs / 86_400_000)
  const hours = Math.floor((abs % 86_400_000) / 3_600_000)
  const minutes = Math.floor((abs % 3_600_000) / 60_000)
  const seconds = Math.floor((abs % 60_000) / 1000)
  const pad = (n: number) => String(n).padStart(2, "0")
  if (days > 0) return `${days}d ${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
}

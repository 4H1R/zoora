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

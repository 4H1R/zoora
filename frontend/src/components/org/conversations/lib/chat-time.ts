import { format, isSameDay, subDays } from "date-fns"
import { format as jFormat } from "date-fns-jalali"
import { faIR } from "date-fns-jalali/locale"

// Message timestamps render in the active calendar system: Jalali (Persian
// digits) for fa, Gregorian for everything else. Time-of-day is clock-agnostic
// so only the digit shaping differs between locales.

/** Localized time-of-day (HH:mm) for a message bubble corner. */
export function formatTimeOfDay(iso: string | undefined, lang: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  return lang === "fa" ? jFormat(d, "HH:mm", { locale: faIR }) : format(d, "HH:mm")
}

export type DayDividerParts = {
  /** "today" | "yesterday" when the date is recent, else null. */
  relative: "today" | "yesterday" | null
  /** Absolute friendly date in the active calendar system (fallback label). */
  absolute: string
}

/**
 * Split a message date into a relative bucket (today/yesterday) plus an
 * absolute friendly label. The caller localizes the relative bucket via i18n
 * and shows `absolute` for older days. Same instant => same bucket regardless
 * of fa/en calendar, so comparison is done on the raw Date.
 */
export function dayDividerParts(iso: string | undefined, lang: string): DayDividerParts {
  const d = iso ? new Date(iso) : new Date()
  const now = new Date()
  const isFa = lang === "fa"
  const absolute = isFa
    ? jFormat(d, "d MMMM yyyy", { locale: faIR })
    : format(d, "d MMMM yyyy")

  let relative: DayDividerParts["relative"] = null
  if (isSameDay(d, now)) relative = "today"
  else if (isSameDay(d, subDays(now, 1))) relative = "yesterday"

  return { relative, absolute }
}

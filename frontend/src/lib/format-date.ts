import { useTranslation } from "react-i18next"

/**
 * Single source of truth for date/time formatting across the app.
 *
 * All formatting goes through `Intl.DateTimeFormat` keyed on the active i18n
 * language, so adding a locale needs no changes here — `fa` automatically
 * resolves to the Jalali calendar with Persian digits, every other locale to
 * its own calendar. Add or tweak a format in ONE place: `PRESETS` below.
 */

export type DateVariant =
  | "date" // 1 Jan 2026
  | "datetime" // Fri, 1 Jan, 12:34
  | "datetime-long" // Fri, 1 January 2026, 12:34
  | "time" // 12:34
  | "weekday-long" // Friday, 1 January
  | "year" // 2026

const PRESETS: Record<DateVariant, Intl.DateTimeFormatOptions> = {
  date: { year: "numeric", month: "short", day: "numeric" },
  datetime: { weekday: "short", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" },
  "datetime-long": {
    weekday: "short",
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  },
  time: { hour: "2-digit", minute: "2-digit" },
  "weekday-long": { weekday: "long", month: "long", day: "numeric" },
  year: { year: "numeric" },
}

const EMPTY = "—"

/**
 * Pure formatter — use when you already have the locale (e.g. outside React or
 * when the locale is passed in). Returns `—` for missing/invalid input.
 */
export function formatDate(
  value: string | number | Date | undefined | null,
  locale: string,
  variant: DateVariant = "date"
): string {
  if (value === undefined || value === null || value === "") return EMPTY
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return EMPTY
  return new Intl.DateTimeFormat(locale, PRESETS[variant]).format(d)
}

/**
 * Same presets as `formatDate`, but returns the locale parts — for UI that
 * styles individual segments (e.g. blinking the time separator) while keeping
 * locale digits (Persian numerals) intact. Returns `[]` for invalid input.
 */
export function formatDateToParts(
  value: string | number | Date | undefined | null,
  locale: string,
  variant: DateVariant = "date"
): Intl.DateTimeFormatPart[] {
  if (value === undefined || value === null || value === "") return []
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return []
  return new Intl.DateTimeFormat(locale, PRESETS[variant]).formatToParts(d)
}

/**
 * Hook bound to the active language. Returns a formatter that defaults to the
 * `date` variant so existing date-only call sites work unchanged.
 */
export function useFormatDate() {
  const { i18n } = useTranslation()
  return (value?: string | number | Date | null, variant: DateVariant = "date") =>
    formatDate(value, i18n.language, variant)
}

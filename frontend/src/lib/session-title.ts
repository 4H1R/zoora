import { useTranslation } from "react-i18next"

import { formatDate } from "./format-date"

/**
 * Builds a human-friendly default session/room title from a start time, e.g.
 * "کلاس دوشنبه ۲۰ تیر ساعت ۱۱:۳۰" (fa, Jalali) or "Class Mon, 20 Jul at 11:30".
 *
 * Locale-aware via the active language — `fa` resolves to the Jalali calendar
 * with Persian digits automatically (see `format-date.ts`).
 */
export function useSessionTitle() {
  const { t, i18n } = useTranslation()
  return (value: string | number | Date) =>
    t("common.autoSessionTitle", {
      date: formatDate(value, i18n.language, "weekday-long"),
      time: formatDate(value, i18n.language, "time"),
    })
}

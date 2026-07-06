// Compact relative-time formatter backed by Intl.RelativeTimeFormat so it
// localizes for both en (LTR) and fa (RTL) without a date library. Falls back
// to an absolute date once an item is older than a week.
const DIVISIONS: { amount: number; unit: Intl.RelativeTimeFormatUnit }[] = [
  { amount: 60, unit: "second" },
  { amount: 60, unit: "minute" },
  { amount: 24, unit: "hour" },
  { amount: 7, unit: "day" },
]

export function formatRelativeTime(iso: string | undefined, locale: string): string {
  if (!iso) return ""
  const date = new Date(iso)
  const now = Date.now()
  let duration = (date.getTime() - now) / 1000

  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto", style: "narrow" })
  for (const { amount, unit } of DIVISIONS) {
    if (Math.abs(duration) < amount) {
      return rtf.format(Math.round(duration), unit)
    }
    duration /= amount
  }

  // Older than a week — show a short absolute date instead of "N weeks ago".
  return date.toLocaleDateString(locale, { day: "numeric", month: "short", year: "numeric" })
}

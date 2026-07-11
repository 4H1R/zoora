import type { DayButtonProps } from "react-day-picker"

import { useTranslation } from "react-i18next"

import { CalendarGregorian } from "@/components/ui/calendar/calendar-gregorian"
import { CalendarJalali } from "@/components/ui/calendar/calendar-jalali"
import { bucketByDay, dateKey, eventDotColor, type CalendarEvent } from "@/lib/calendar"
import { cn } from "@/lib/utils"

type CalendarBoardProps = {
  events: CalendarEvent[]
  month: Date
  onMonthChange: (d: Date) => void
  selected: Date | undefined
  onSelect: (d: Date | undefined) => void
}

// CalendarBoard is the shared month grid: it renders the Jalali or Gregorian
// day-picker (chosen by the active language) and overlays up to four colored
// per-type dots on days that have events. Selection/navigation is delegated to
// the parent so it can be reused by both the full calendar page and the
// dashboard widget.
export function CalendarBoard({
  events,
  month,
  onMonthChange,
  selected,
  onSelect,
}: CalendarBoardProps) {
  const { i18n } = useTranslation()
  const isFa = i18n.language === "fa"
  const buckets = bucketByDay(events)

  function DayButton(props: DayButtonProps) {
    const { day, modifiers, className, children, ...rest } = props
    const key = dateKey(day.date)
    const types = Array.from(new Set((buckets.get(key) ?? []).map((e) => e.type)))
    const selected = modifiers.selected
    return (
      <button
        {...rest}
        className={cn(
          className,
          // Stack the number over a fixed dot row so the dots never crowd the
          // digit; a taller cell gives both room to breathe.
          "relative flex h-11 w-full flex-col items-center justify-center gap-1"
        )}
      >
        <span className="leading-none">{children}</span>
        {/* Reserve the dot row on every cell so the number sits at the same
            height whether or not a day has events. */}
        <span className="pointer-events-none flex h-1.5 items-center justify-center gap-1">
          {types.slice(0, 4).map((type) => (
            <span
              key={type}
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                // On the selected (green) cell the colored dots wash out — swap
                // to crisp white so they stay legible in both light and dark.
                selected ? "bg-white" : eventDotColor(type)
              )}
            />
          ))}
        </span>
      </button>
    )
  }

  if (isFa) {
    return (
      <CalendarJalali
        mode="single"
        month={month}
        onMonthChange={onMonthChange}
        selected={selected}
        onSelect={onSelect}
        components={{ DayButton }}
      />
    )
  }
  return (
    <CalendarGregorian
      mode="single"
      month={month}
      onMonthChange={onMonthChange}
      selected={selected}
      onSelect={onSelect}
      components={{ DayButton }}
    />
  )
}

import type { CalendarEvent } from "@/lib/calendar"
import type { DayButtonProps } from "react-day-picker"

import { createContext, useContext } from "react"
import { useTranslation } from "react-i18next"

import { CalendarGregorian } from "@/components/ui/calendar/calendar-gregorian"
import { CalendarJalali } from "@/components/ui/calendar/calendar-jalali"
import { bucketByDay, dateKey, eventDotColor } from "@/lib/calendar"
import { cn } from "@/lib/utils"

type CalendarBoardProps = {
  events: CalendarEvent[]
  month: Date
  onMonthChange: (d: Date) => void
  selected: Date | undefined
  onSelect: (d: Date | undefined) => void
}

// Per-day event buckets handed to the day-picker's DayButton slot. react-day-picker's
// `components` API only passes DayButtonProps, so the buckets travel via context
// instead of a prop — letting DayButton live at module scope.
const DayBucketsCtx = createContext<Map<string, CalendarEvent[]>>(new Map())

// Renders one day cell: the number plus up to four colored per-type dots for any
// events on that day.
function DayButton(props: DayButtonProps) {
  const { day, modifiers, className, children, ...rest } = props
  const buckets = useContext(DayBucketsCtx)
  const key = dateKey(day.date)
  const types = Array.from(new Set((buckets.get(key) ?? []).map((e) => e.type)))
  const isSelected = modifiers.selected
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
              isSelected ? "bg-white" : eventDotColor(type)
            )}
          />
        ))}
      </span>
    </button>
  )
}

// CalendarBoard is the shared month grid: it renders the Jalali or Gregorian
// day-picker (chosen by the active language) and overlays up to four colored
// per-type dots on days that have events. Selection/navigation is delegated to
// the parent so it can be reused by both the full calendar page and the
// dashboard widget.
export function CalendarBoard({ events, month, onMonthChange, selected, onSelect }: CalendarBoardProps) {
  const { i18n } = useTranslation()
  const isFa = i18n.language === "fa"
  const buckets = bucketByDay(events)

  if (isFa) {
    return (
      <DayBucketsCtx.Provider value={buckets}>
        <CalendarJalali
          mode="single"
          month={month}
          onMonthChange={onMonthChange}
          selected={selected}
          onSelect={onSelect}
          components={{ DayButton }}
        />
      </DayBucketsCtx.Provider>
    )
  }
  return (
    <DayBucketsCtx.Provider value={buckets}>
      <CalendarGregorian
        mode="single"
        month={month}
        onMonthChange={onMonthChange}
        selected={selected}
        onSelect={onSelect}
        components={{ DayButton }}
      />
    </DayBucketsCtx.Provider>
  )
}

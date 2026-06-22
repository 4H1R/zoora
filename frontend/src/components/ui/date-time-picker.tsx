import { CalendarIcon, ClockIcon, XIcon } from "lucide-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { CalendarGregorian } from "@/components/ui/calendar/calendar-gregorian"
import { CalendarJalali } from "@/components/ui/calendar/calendar-jalali"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { ScrollArea } from "@/components/ui/scroll-area"
import { formatDate } from "@/lib/format-date"
import { cn } from "@/lib/utils"

export interface DateTimePickerProps {
  /** Current value as an ISO 8601 string (UTC), or `undefined` when empty. */
  value?: string
  /** Emits an ISO 8601 string (UTC), or `undefined` when cleared. */
  onChange: (value: string | undefined) => void
  /** Show hour/minute selects. Date-only when false. Default `true`. */
  showTime?: boolean
  /** Allow clearing back to empty. Default `false`. */
  clearable?: boolean
  disabled?: boolean
  /** UX-only lower bound — days before are disabled. Validation stays in Zod. */
  minDate?: Date
  /** UX-only upper bound — days after are disabled. Validation stays in Zod. */
  maxDate?: Date
  placeholder?: string
  /** Renders the error ring (wire to RHF `fieldState.invalid`). */
  invalid?: boolean
  id?: string
  className?: string
}

const HOURS = Array.from({ length: 24 }, (_, i) => i)
const MINUTES = Array.from({ length: 60 }, (_, i) => i)

function isoToDate(iso?: string): Date | undefined {
  if (!iso) return undefined
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? undefined : d
}

export function DateTimePicker({
  value,
  onChange,
  showTime = true,
  clearable = false,
  disabled = false,
  minDate,
  maxDate,
  placeholder,
  invalid,
  id,
  className,
}: DateTimePickerProps) {
  const { t, i18n } = useTranslation()
  const lang = i18n.language
  const isFa = lang === "fa"
  const [open, setOpen] = React.useState(false)

  const selected = isoToDate(value)

  // Two-digit, locale-aware (Persian digits for `fa`) for the time cells.
  const pad2 = (n: number) => new Intl.NumberFormat(lang, { minimumIntegerDigits: 2, useGrouping: false }).format(n)

  const label = value
    ? formatDate(value, lang, showTime ? "datetime" : "date")
    : (placeholder ?? t("common.dateTimePicker.placeholder"))

  // Picking a day keeps the existing time (or 00:00 when the field was empty).
  function handleDaySelect(day: Date | undefined) {
    if (!day) {
      if (clearable) onChange(undefined)
      return
    }
    const next = new Date(day)
    next.setHours(showTime ? (selected?.getHours() ?? 0) : 0, showTime ? (selected?.getMinutes() ?? 0) : 0, 0, 0)
    onChange(next.toISOString())
    if (!showTime) setOpen(false)
  }

  // Changing time before a day is picked anchors to the current day.
  function handleTime(part: "h" | "m", val: number) {
    const next = selected ? new Date(selected) : new Date()
    if (part === "h") next.setHours(val)
    else next.setMinutes(val)
    next.setSeconds(0, 0)
    onChange(next.toISOString())
  }

  function handleNow() {
    const now = new Date()
    now.setSeconds(0, 0)
    onChange(now.toISOString())
  }

  const dayDisabled = [
    ...(minDate ? [{ before: minDate }] : []),
    ...(maxDate ? [{ after: maxDate }] : []),
  ]
  const Calendar = isFa ? CalendarJalali : CalendarGregorian

  return (
    <div className={cn("relative", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          render={
            <Button
              id={id}
              type="button"
              variant="outline"
              disabled={disabled}
              aria-invalid={invalid || undefined}
              className={cn("w-full justify-start font-normal", clearable && value && "pe-9", !value && "text-muted-foreground")}
            />
          }
        >
          <CalendarIcon className="size-4 opacity-70" />
          <span className="flex-1 text-start">{label}</span>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <div className="flex flex-col sm:flex-row">
            <Calendar
              mode="single"
              selected={selected}
              onSelect={handleDaySelect}
              defaultMonth={selected}
              disabled={dayDisabled}
              autoFocus
            />
            {showTime ? (
              <div
                dir="ltr"
                className="flex w-full flex-col border-t sm:w-auto sm:border-s sm:border-t-0"
              >
                <div className="flex h-12 items-center justify-between gap-3 border-b px-3">
                  <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs font-medium">
                    <ClockIcon className="size-3.5" />
                    <span className="text-foreground font-mono text-sm tabular-nums">
                      {selected ? `${pad2(selected.getHours())}:${pad2(selected.getMinutes())}` : "--:--"}
                    </span>
                  </span>
                  <Button
                    type="button"
                    size="xs"
                    variant="ghost"
                    disabled={disabled}
                    onClick={handleNow}
                  >
                    {t("common.dateTimePicker.now")}
                  </Button>
                </div>
                <div className="flex h-56 min-h-0 sm:h-64">
                  <TimeColumn
                    open={open}
                    values={HOURS}
                    selected={selected?.getHours()}
                    onSelect={(v) => handleTime("h", v)}
                    format={pad2}
                    ariaLabel={t("common.dateTimePicker.hour")}
                    disabled={disabled}
                  />
                  <div className="bg-border w-px" />
                  <TimeColumn
                    open={open}
                    values={MINUTES}
                    selected={selected?.getMinutes()}
                    onSelect={(v) => handleTime("m", v)}
                    format={pad2}
                    ariaLabel={t("common.dateTimePicker.minute")}
                    disabled={disabled}
                  />
                </div>
              </div>
            ) : null}
          </div>
        </PopoverContent>
      </Popover>
      {clearable && value && !disabled ? (
        <button
          type="button"
          aria-label={t("common.dateTimePicker.clear")}
          onClick={() => onChange(undefined)}
          className="text-muted-foreground hover:text-foreground absolute end-2 top-1/2 -translate-y-1/2 rounded-sm p-0.5"
        >
          <XIcon className="size-4" />
        </button>
      ) : null}
    </div>
  )
}

interface TimeColumnProps {
  open: boolean
  values: number[]
  selected?: number
  onSelect: (value: number) => void
  format: (n: number) => string
  ariaLabel: string
  disabled?: boolean
}

function TimeColumn({ open, values, selected, onSelect, format, ariaLabel, disabled }: TimeColumnProps) {
  const activeRef = React.useRef<HTMLButtonElement>(null)

  // Center the active cell each time the popover opens so the current value is visible.
  React.useEffect(() => {
    if (open && activeRef.current) {
      activeRef.current.scrollIntoView({ block: "center" })
    }
  }, [open, selected])

  return (
    <ScrollArea className="h-full">
      <div role="listbox" aria-label={ariaLabel} className="flex flex-col gap-0.5 p-2">
        {values.map((v) => {
          const active = v === selected
          return (
            <button
              key={v}
              ref={active ? activeRef : undefined}
              type="button"
              role="option"
              aria-selected={active}
              aria-label={`${ariaLabel} ${format(v)}`}
              disabled={disabled}
              onClick={() => onSelect(v)}
              className={cn(
                "flex h-8 w-12 shrink-0 items-center justify-center rounded-md font-mono text-sm tabular-nums outline-none transition-colors",
                "hover:bg-muted focus-visible:ring-ring/50 focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50",
                active && "bg-primary text-primary-foreground hover:bg-primary"
              )}
            >
              {format(v)}
            </button>
          )
        })}
      </div>
    </ScrollArea>
  )
}

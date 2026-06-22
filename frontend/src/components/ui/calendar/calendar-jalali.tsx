import type { ChevronProps } from "react-day-picker"

import { faIR } from "date-fns-jalali/locale"
import * as React from "react"
import { DayPicker } from "react-day-picker/jalali"

import { cn } from "@/lib/utils"

import { getClassNames } from "./calendar.styles"
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"

export type CalendarJalaliProps = React.ComponentProps<typeof DayPicker>

function CalendarJalali({
  components,
  showOutsideDays = true,
  className,
  classNames,
  ...props
}: CalendarJalaliProps) {
  return (
    <DayPicker
      dir="rtl"
      className={cn("p-3", className)}
      noonSafe
      components={{
        Chevron,
        ...components,
      }}
      locale={faIR}
      showOutsideDays={showOutsideDays}
      classNames={{
        ...getClassNames(props),
        ...classNames,
      }}
      {...props}
    />
  )
}
CalendarJalali.displayName = "CalendarJalali"

function Chevron({ orientation }: ChevronProps) {
  if (orientation === "left") return <ChevronRightIcon className="h-4 w-4" />
  if (orientation === "right") return <ChevronLeftIcon className="h-4 w-4" />
  return <span />
}

export { CalendarJalali }

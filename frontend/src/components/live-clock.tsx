import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { formatDate, formatDateToParts } from "@/lib/format-date"
import { cn } from "@/lib/utils"

function useNow() {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(id)
  }, [])
  return now
}

// Persian (fa) gets Jalali calendar + Persian digits via Intl automatically.
// Time is split into parts so the separator can blink without breaking locale
// digits — formatToParts keeps fa's Persian numerals intact.
export function LiveClock({ className }: { className?: string }) {
  const { i18n } = useTranslation()
  const now = useNow()
  const locale = i18n.language

  const timeParts = formatDateToParts(now, locale, "time")
  const date = formatDate(now, locale, "weekday-long")

  const tick = now.getSeconds() % 2 === 0

  return (
    <div
      className={cn(
        "group border-border/60 bg-card/50 hover:border-border hover:bg-accent/40 relative flex items-center gap-2.5 rounded-lg border px-2.5 py-1 shadow-xs transition-colors",
        className
      )}
    >
      {/* heartbeat — a steady green tick that breathes once per second */}
      <span className="relative flex size-2 shrink-0 items-center justify-center">
        <span
          className={cn(
            "bg-success/60 absolute inline-flex size-full rounded-full transition-all duration-700 ease-out",
            tick ? "scale-100 opacity-0" : "scale-50 opacity-70"
          )}
        />
        <span className="bg-success size-1.5 rounded-full" />
      </span>

      <div className="flex flex-col leading-none">
        <span className="text-foreground font-mono text-sm font-semibold tracking-tight tabular-nums">
          {timeParts.map((part, i) =>
            part.type === "literal" ? (
              <span key={i} className={cn("transition-opacity duration-300", tick ? "opacity-100" : "opacity-30")}>
                {part.value}
              </span>
            ) : (
              <span key={i}>{part.value}</span>
            )
          )}
        </span>
        <span className="text-muted-foreground tracking-caps mt-0.5 hidden text-[0.625rem] font-medium uppercase lg:inline">
          {date}
        </span>
      </div>
    </div>
  )
}

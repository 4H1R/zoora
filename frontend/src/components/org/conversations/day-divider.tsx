import { useTranslation } from "react-i18next"

import { dayDividerParts } from "./lib/chat-time"

interface DayDividerProps {
  /** ISO timestamp of the first message on this calendar day. */
  date: string | undefined
}

/**
 * Centered date pill separating messages from different calendar days. Shows a
 * friendly "Today"/"Yesterday" for recent days and an absolute date otherwise,
 * localized to the active calendar (Jalali for fa).
 */
export function DayDivider({ date }: DayDividerProps) {
  const { t, i18n } = useTranslation()
  const { relative, absolute } = dayDividerParts(date, i18n.language)
  const label = relative ? t(`conversations.thread.${relative}`) : absolute

  return (
    <div className="flex items-center justify-center py-4">
      <span className="bg-muted/70 text-muted-foreground ring-border/50 rounded-full px-3 py-1 text-xs font-medium shadow-sm ring-1 backdrop-blur-sm">
        {label}
      </span>
    </div>
  )
}

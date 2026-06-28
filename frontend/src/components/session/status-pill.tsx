import { useTranslation } from "react-i18next"

import { type SessionStatus } from "@/lib/session-status"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

interface SessionStatusPillProps {
  status: SessionStatus
  size?: "sm" | "md"
  className?: string
}

export function SessionStatusPill({ status, size = "md", className }: SessionStatusPillProps) {
  const { t } = useTranslation()
  const dotSize = size === "sm" ? "size-1.5" : "size-2"
  const padding = size === "sm" ? "px-2 py-0.5" : "px-2.5 py-1"

  if (status === "live") {
    return (
      <span
        className={cn(
          "bg-destructive/10 text-destructive inline-flex items-center gap-1.5 rounded-full font-mono text-xs tracking-[0.25em] uppercase",
          padding,
          className
        )}
      >
        <span className={cn("relative flex", dotSize)}>
          <span className="bg-destructive absolute inline-flex h-full w-full animate-ping rounded-full opacity-75" />
          <span className={cn("bg-destructive relative inline-flex rounded-full", dotSize)} />
        </span>
        {t("status.live")}
      </span>
    )
  }

  const variant = status === "ended" ? "outline" : "secondary"
  const label = status === "ended" ? t("status.ended") : t("status.scheduled")

  return (
    <Badge variant={variant} className={cn("font-mono text-xs tracking-[0.25em] uppercase", className)}>
      {label}
    </Badge>
  )
}

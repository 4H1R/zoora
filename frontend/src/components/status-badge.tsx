import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

type Status = "live" | "scheduled" | "ended" | "processing" | "failed" | "draft"

interface StatusBadgeProps {
  status: Status
  className?: string
  children?: React.ReactNode
}

const statusConfig: Record<Status, { bg: string; text: string; dot: string; pulse?: boolean; mono?: boolean }> = {
  live: { bg: "bg-[#dc2626]", text: "text-white", dot: "bg-white", pulse: true, mono: true },
  scheduled: { bg: "bg-[var(--green-50)]", text: "text-[var(--green-800)]", dot: "bg-[var(--green-600)]" },
  ended: { bg: "bg-muted", text: "text-muted-foreground", dot: "bg-muted-foreground" },
  processing: { bg: "bg-[#fffbeb]", text: "text-[#92400e]", dot: "bg-[#d97706]" },
  failed: { bg: "bg-[#fef2f2]", text: "text-[#991b1b]", dot: "bg-[#dc2626]" },
  draft: { bg: "bg-[#eff6ff]", text: "text-[#1e40af]", dot: "bg-[#2563eb]" },
}

export function StatusBadge({ status, className, children }: StatusBadgeProps) {
  const { t } = useTranslation()
  const config = statusConfig[status]

  return (
    <span
      className={cn(
        "inline-flex items-center gap-[5px] rounded-full px-2 py-1 text-[11px] leading-none font-medium",
        config.bg,
        config.text,
        config.mono && "tracking-[0.06em]",
        className
      )}
    >
      <span className={cn("size-1.5 rounded-full", config.dot, config.pulse && "animate-pulse-dot")} />
      {children ?? t(`status.${status}`)}
    </span>
  )
}

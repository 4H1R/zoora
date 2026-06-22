import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

export type PracticeStatus = "upcoming" | "to_submit" | "submitted" | "graded" | "missed"

const STATUS_STYLES: Record<PracticeStatus, string> = {
  upcoming: "border-border bg-muted text-muted-foreground",
  to_submit: "border-primary/25 bg-primary/10 text-primary",
  submitted: "border-sky-500/25 bg-sky-500/10 text-sky-700 dark:text-sky-400",
  graded: "border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
  missed: "border-destructive/25 bg-destructive/10 text-destructive",
}

export function PracticeStatusBadge({
  status,
  className,
}: {
  status?: string
  className?: string
}) {
  const { t } = useTranslation()
  const key = (status ?? "to_submit") as PracticeStatus
  const style = STATUS_STYLES[key] ?? STATUS_STYLES.to_submit
  return (
    <Badge variant="outline" className={cn("font-medium", style, className)}>
      {t(`org.practices.status.${key}`)}
    </Badge>
  )
}

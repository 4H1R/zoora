import type {
  GithubCom4H1RZooraInternalDomainTicketStatus as TicketStatus,
  GithubCom4H1RZooraInternalDomainTicketType as TicketType,
} from "@/api/model"

import {
  CircleDotIcon,
  GraduationCapIcon,
  HelpCircleIcon,
  LockIcon,
  MessageCircleReplyIcon,
  TagIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

// Status is never encoded by color alone: each badge pairs its tint with an
// icon + translated label so open/answered/closed stay distinguishable.
const STATUS_STYLES: Record<string, { className: string; icon: React.ReactNode }> = {
  open: {
    className: "border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-400",
    icon: <CircleDotIcon />,
  },
  answered: {
    className: "border-sky-500/25 bg-sky-500/10 text-sky-700 dark:text-sky-400",
    icon: <MessageCircleReplyIcon />,
  },
  closed: {
    className: "border-border bg-muted text-muted-foreground",
    icon: <LockIcon />,
  },
}

export function TicketStatusBadge({ status, className }: { status?: TicketStatus; className?: string }) {
  const { t } = useTranslation()
  const key = status ?? "open"
  const style = STATUS_STYLES[key] ?? STATUS_STYLES.open
  return (
    <Badge variant="outline" className={cn("font-medium", style.className, className)}>
      {style.icon}
      {t(`tickets.status.${key}`)}
    </Badge>
  )
}

const TYPE_ICONS: Record<string, React.ReactNode> = {
  question: <HelpCircleIcon />,
  grade_objection: <GraduationCapIcon />,
  other: <TagIcon />,
}

export function TicketTypeBadge({ type, className }: { type?: TicketType; className?: string }) {
  const { t } = useTranslation()
  const key = type ?? "other"
  return (
    <Badge variant="outline" className={cn("text-muted-foreground font-medium", className)}>
      {TYPE_ICONS[key] ?? TYPE_ICONS.other}
      {t(`tickets.type.${key}`)}
    </Badge>
  )
}

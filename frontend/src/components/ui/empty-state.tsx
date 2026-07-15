import type { LucideIcon } from "lucide-react"
import type { ReactNode } from "react"

import { cn } from "@/lib/utils"

interface EmptyStateProps {
  /** Lucide icon component shown above the title. */
  icon: LucideIcon
  title: string
  description?: string
  /** Optional action area (e.g. a create button), rendered below the text. */
  children?: ReactNode
  className?: string
}

/** Shared empty/no-results state. Use everywhere a list, grid, or section has
 * nothing to show — whether it's an active filter returning nothing or a
 * create-prompt for empty data. Pass an action via `children`. */
export function EmptyState({ icon: Icon, title, description, children, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1",
        className
      )}
    >
      <Icon className="text-muted-foreground size-8" />
      <h3 className="text-foreground text-lg font-semibold tracking-tight">{title}</h3>
      {description && <p className="text-muted-foreground max-w-md text-sm leading-relaxed">{description}</p>}
      {Boolean(children) && <div className="mt-2">{children}</div>}
    </div>
  )
}

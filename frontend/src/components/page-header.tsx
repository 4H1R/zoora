import { cn } from "@/lib/utils"

export function PageHeader({
  title,
  actions,
  className,
}: {
  title: string
  actions?: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex items-end justify-between gap-6", className)}>
      <div className="min-w-0">
        <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </div>
  )
}

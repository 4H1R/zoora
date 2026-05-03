import { Card, CardContent } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export type StatItem = {
  icon: React.ReactNode
  label: string
  value?: number
  loading?: boolean
  detail?: React.ReactNode
}

export function StatCard({
  icon,
  label,
  value,
  loading,
  detail,
  className,
  style,
}: StatItem & { className?: string; style?: React.CSSProperties }) {
  return (
    <Card
      size="sm"
      className={cn(
        "relative overflow-hidden transition-all duration-[--dur-slow] ease-[--ease-out]",
        "animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both duration-300",
        className
      )}
      style={style}
    >
      <div className="bg-muted-foreground/30 absolute inset-y-0 start-0 w-0.5" />

      <CardContent className="flex items-center gap-3 pt-3">
        <div className="bg-muted text-muted-foreground flex size-9 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
          {icon}
        </div>

        <div className="min-w-0 flex-1">
          <p className="text-muted-foreground truncate text-xs font-medium tracking-wide uppercase">{label}</p>

          {loading ? (
            <Skeleton className="mt-1 h-7 w-14" />
          ) : (
            <div className="flex items-baseline gap-1.5">
              <span className="text-xl font-semibold tracking-tight tabular-nums">{(value ?? 0).toLocaleString()}</span>
              {detail && (
                <>
                  <span className="text-border text-xs">·</span>
                  <span className="text-muted-foreground truncate text-xs">{detail}</span>
                </>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function StatCards({ stats, className }: { stats: StatItem[]; className?: string }) {
  return (
    <div className={cn("grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3", className)}>
      {stats.map((stat, i) => (
        <StatCard key={stat.label} {...stat} style={{ animationDelay: `${i * 60}ms` }} />
      ))}
    </div>
  )
}

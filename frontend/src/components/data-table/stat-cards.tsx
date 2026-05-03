import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

export type StatItem = {
  icon: React.ReactNode
  label: string
  value?: number
  loading?: boolean
  detail?: React.ReactNode
}

export function StatCard({ icon, label, value, loading, detail, className }: StatItem & { className?: string }) {
  return (
    <Card size="sm" className={cn(className)}>
      <CardHeader>
        <div className="flex items-center gap-2">
          <span className="text-muted-foreground [&>svg]:size-3.5">{icon}</span>
          <CardTitle className="text-muted-foreground text-[11px] font-medium tracking-wide uppercase">
            {label}
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-8 w-16" />
        ) : (
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold tracking-tight tabular-nums">{(value ?? 0).toLocaleString()}</span>
            {detail && <span className="text-muted-foreground text-xs">{detail}</span>}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export function StatCards({ stats, className }: { stats: StatItem[]; className?: string }) {
  return (
    <div className={cn("grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3", className)}>
      {stats.map((stat) => (
        <StatCard key={stat.label} {...stat} />
      ))}
    </div>
  )
}

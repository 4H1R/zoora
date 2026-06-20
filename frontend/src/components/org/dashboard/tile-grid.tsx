import { Link } from "@tanstack/react-router"

import { Card } from "@/components/ui/card"
import { cn } from "@/lib/utils"

export type DashboardTileSpec = {
  key: string
  label: string
  icon: React.ReactNode
  to: string
  params?: Record<string, string>
}

export function DashboardTile({
  label,
  icon,
  to,
  params,
  style,
}: Omit<DashboardTileSpec, "key"> & { style?: React.CSSProperties }) {
  return (
    <Link to={to} params={params} className="group block" style={style}>
      <Card
        size="sm"
        className={cn(
          "relative h-full flex-col items-center justify-center gap-3 overflow-hidden p-6 text-center",
          "hover:border-primary/40 hover:bg-muted/40 transition-colors duration-[--dur-slow] ease-[--ease-out]",
          "animate-in fade-in-0 slide-in-from-bottom-2 fill-mode-both duration-300"
        )}
      >
        <div className="bg-muted text-muted-foreground group-hover:bg-primary/10 group-hover:text-primary flex size-12 items-center justify-center rounded-xl transition-colors [&>svg]:size-6">
          {icon}
        </div>
        <span className="text-sm font-medium">{label}</span>
      </Card>
    </Link>
  )
}

export function TileGrid({ tiles, className }: { tiles: DashboardTileSpec[]; className?: string }) {
  return (
    <div className={cn("grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4", className)}>
      {tiles.map(({ key, ...tile }, i) => (
        <DashboardTile key={key} {...tile} style={{ animationDelay: `${i * 60}ms` }} />
      ))}
    </div>
  )
}
